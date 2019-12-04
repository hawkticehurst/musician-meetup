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
	"github.com/gorilla/websocket"
)

// Control messages for websocket
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

// Message represents a RabbitMQ message
type Message struct {
	Type          string
	Channel       string `json:"channel"`
	ChannelID     string `json:"channelID"`
	UserMessage   string `json:"message"`
	UserMessageID string `json:"messageID"`
	UserIDs       []int64
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// This function's purpose is to reject websocket upgrade requests if the
		// origin of the websockete handshake request is coming from unknown domains.
		// This prevents some random domain from opening up a socket with your server.
		// TODO: make sure you modify this for your HW to check if r.Origin is your host

		return true
	},
}

// Data structure containing every current websocket connection
var socketStore *handlers.SocketStore = handlers.NewSocketStore()

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

	stringSummaryAddr := strings.Split(os.Getenv("SUMMARYADDR"), ",")
	stringMessageAddr := strings.Split(os.Getenv("MESSAGESADDR"), ",")
	stringMeetupAddr := strings.Split(os.Getenv("MEETUPADDR"), ",")
	if len(stringSummaryAddr) == 0 || len(stringMessageAddr) == 0 || len(stringMeetupAddr) == 0 {
		err := fmt.Errorf("Environment variables SUMMARYADDR, MESSAGESADDR, and MEETUPADDR should be set.\nSUMMARYADDR: %s\nMESSAGESADDR: %s\nMEETUPADDR: %s", stringSummaryAddr, stringMessageAddr, stringMeetupAddr)
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

	lenMeetup := len(stringMeetupAddr)
	urlMeetupAddr := make([]*url.URL, lenMeetup)
	for i, stringURL := range stringMeetupAddr {
		urlAddr, _ := url.Parse(stringURL)
		urlMeetupAddr[i] = urlAddr
	}

	summaryDirector := CustomDirector(urlSummaryAddr, sessionKey, redisStore)
	messageDirector := CustomDirector(urlMessageAddr, sessionKey, redisStore)
	meetupDirector := CustomDirector(urlMeetupAddr, sessionKey, redisStore)

	summaryProxy := &httputil.ReverseProxy{Director: summaryDirector}
	messagingProxy := &httputil.ReverseProxy{Director: messageDirector}
	meetupProxy := &httputil.ReverseProxy{Director: meetupDirector}

	mux.Handle("/v1/summary", summaryProxy)
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
			//log.Printf("Error: Retrieving Session ID: %v", err)
			r.Header["X-User"] = nil
		} else {
			//log.Println("No error getting sessionID")
			sessionState := &handlers.SessionState{}
			sessions.GetState(r, sessionKey, redisstore, sessionState)
			user := sessionState.User
			bytes, _ := json.Marshal(user)
			//log.Println("User JSON:")
			//log.Println(string(bytes))
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

func contains(userID int64, userIDs []int64) bool {
	for _, ID := range userIDs {
		if ID == userID {
			return true
		}
	}
	return false
}
