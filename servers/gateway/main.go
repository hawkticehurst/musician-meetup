package main

import (
	"assignments-hawkticehurst/servers/gateway/handlers"
	"assignments-hawkticehurst/servers/gateway/models/users"
	"assignments-hawkticehurst/servers/gateway/sessions"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

//main is the main entry point for the server
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

	stringSummaryAddr := strings.Split(os.Getenv("SUMMARYADDR"), ",")
	stringMessageAddr := strings.Split(os.Getenv("MESSAGESADDR"), ",")
	if len(stringSummaryAddr) == 0 || len(stringMessageAddr) == 0 {
		err := fmt.Errorf("Environment variables SUMMARYADDR and MESSAGESADDR should be set.\nSUMMARYADDR: %s\nMESSAGESADDR: %s", stringSummaryAddr, stringMessageAddr)
		fmt.Println(err.Error())
		os.Exit(1)
	}

	lenSummary := len(stringSummaryAddr)
	urlSummaryAddr := make([]*url.URL, lenSummary)
	for i, stringURL := range stringSummaryAddr {
		urlAddr, _ := url.Parse(stringURL)
		urlSummaryAddr[i] = urlAddr
	}

	lenMessage := len(stringMessageAddr)
	urlMessageAddr := make([]*url.URL, lenMessage)
	for i, stringURL := range stringMessageAddr {
		urlAddr, _ := url.Parse(stringURL)
		urlMessageAddr[i] = urlAddr
	}

	summaryDirector := CustomDirector(urlSummaryAddr, sessionKey, redisStore)
	messageDirector := CustomDirector(urlMessageAddr, sessionKey, redisStore)

	summaryProxy := &httputil.ReverseProxy{Director: summaryDirector}
	messagingProxy := &httputil.ReverseProxy{Director: messageDirector}

	//handlers.ReadIncomingMessagesFromRabbit()

	mux.Handle("/v1/summary", summaryProxy)
	mux.Handle("/v1/channels", messagingProxy)
	mux.Handle("/v1/channels/", messagingProxy)
	mux.Handle("/v1/messages/", messagingProxy)
	mux.Handle("/v1/events", messagingProxy)
	mux.Handle("/v1/events/", messagingProxy)

	mux.HandleFunc("/v1/users", hctx.UsersHandler)
	mux.HandleFunc("/v1/users/", hctx.SpecificUserHandler)
	mux.HandleFunc("/v1/sessions", hctx.SessionsHandler)
	mux.HandleFunc("/v1/sessions/", hctx.SpecificSessionHandler)

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
			log.Printf("Error: Retrieving Session ID: %v", err)
			r.Header["X-User"] = nil
		} else {
			log.Println("No error getting sessionID")
			sessionState := &handlers.SessionState{}
			sessions.GetState(r, sessionKey, redisstore, sessionState)
			user := sessionState.User
			bytes, _ := json.Marshal(user)
			log.Println("User JSON:")
			log.Println(string(bytes))
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
