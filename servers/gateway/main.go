package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"serverside-final-project/servers/gateway/handlers"
	"serverside-final-project/servers/gateway/models/users"
	"serverside-final-project/servers/gateway/sessions"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

// main is the main entry point for the server
func main() {
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":443"
	}

	tlsKeyPath := os.Getenv("TLSKEY")
	tlsCertPath := os.Getenv("TLSCERT")
	if len(tlsKeyPath) == 0 || len(tlsCertPath) == 0 {
		err := fmt.Errorf("Environment variables TLSKEY and TLSCERT should be set.\nTLSKEY: %s\nTLSCERT: %s", tlsKeyPath, tlsCertPath)
		fmt.Println(err.Error())
		os.Exit(1)
	}
	sessionKey := os.Getenv("SESSIONKEY")
	reddisAddr := os.Getenv("REDISADDR")
	dsn := os.Getenv("DSN")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     reddisAddr,
		Password: "",
		DB:       0,
	})

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	redisStore := sessions.NewRedisStore(redisClient, time.Hour*100000)
	sqlStore := users.NewMySQLStore(db)

	hctx := handlers.NewContext(sessionKey, redisStore, sqlStore)

	mux := http.NewServeMux()
	wrappedMux := handlers.NewCORSHeader(mux)

	stringMessageAddr := strings.Split(os.Getenv("MESSAGESADDR"), ",")
	stringMeetupAddr := strings.Split(os.Getenv("MEETUPADDR"), ",")
	if len(stringMessageAddr) == 0 || len(stringMeetupAddr) == 0 {
		err := fmt.Errorf("Environment variables MESSAGESADDR, and MEETUPADDR should be set.\nMESSAGESADDR: %s\nMEETUPADDR: %s", stringMessageAddr, stringMeetupAddr)
		fmt.Println(err.Error())
		os.Exit(1)
	}

	lenMessage := len(stringMessageAddr)
	urlMessageAddr := make([]*url.URL, lenMessage)
	for i, stringURL := range stringMessageAddr {
		urlAddr, _ := url.Parse(stringURL)
		urlMessageAddr[i] = urlAddr
	}

	lenMeetup := len(stringMeetupAddr)
	urlMeetupAddr := make([]*url.URL, lenMeetup)
	for i, stringURL := range stringMeetupAddr {
		urlAddr, _ := url.Parse(stringURL)
		urlMeetupAddr[i] = urlAddr
	}

	messageDirector := CustomDirector(urlMessageAddr, sessionKey, redisStore)
	meetupDirector := CustomDirector(urlMeetupAddr, sessionKey, redisStore)

	messagingProxy := &httputil.ReverseProxy{Director: messageDirector}
	meetupProxy := &httputil.ReverseProxy{Director: meetupDirector}

	mux.Handle("/v1/channels", messagingProxy)
	mux.Handle("/v1/channels/", messagingProxy)
	mux.Handle("/v1/messages/", messagingProxy)
	mux.Handle("/v1/events", meetupProxy)
	mux.Handle("/v1/events/", meetupProxy)

	mux.HandleFunc("/v1/users", hctx.UsersHandler)
	mux.HandleFunc("/v1/users/", hctx.SpecificUserHandler)
	mux.HandleFunc("/v1/sessions", hctx.SessionsHandler)
	mux.HandleFunc("/v1/sessions/", hctx.SpecificSessionHandler)

	handlers.ReadIncomingMessagesFromRabbit()
	mux.HandleFunc("/v1/ws", hctx.WebSocketConnectionHandler)

	log.Printf("Server is listening at %s...", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlsCertPath, tlsKeyPath, wrappedMux))
}

// Director represents a director function
type Director func(r *http.Request)

// CustomDirector returns a director function that will be executed in a reverse proxy call
func CustomDirector(targets []*url.URL, sessionKey string, redisstore *sessions.RedisStore) Director {
	var counter int32
	counter = 0
	return func(r *http.Request) {
		_, err := sessions.GetSessionID(r, sessionKey)
		if err != nil {
			r.Header["X-User"] = nil
		} else {
			sessionState := &handlers.SessionState{}
			sessions.GetState(r, sessionKey, redisstore, sessionState)
			user := sessionState.User
			bytes, _ := json.Marshal(user)
			r.Header.Add("X-User", string(bytes[:]))
		}

		targ := targets[counter%int32(len(targets))]
		atomic.AddInt32(&counter, 1) // note, to be extra safe, we’ll need to use mutexes
		if len(targ.Host) > 0 {
			r.Host = targ.Host
			r.URL.Host = targ.Host
		} else {
			r.Host = targ.String()
			r.URL.Host = targ.String()
		}
		if len(targ.Scheme) > 0 {
			r.URL.Scheme = targ.Scheme
		} else {
			r.URL.Scheme = "http"
		}
	}
}
