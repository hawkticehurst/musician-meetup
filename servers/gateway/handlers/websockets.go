package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"serverside-final-project/servers/gateway/sessions"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
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
		return true
	},
}

// Data structure containing every current websocket connection
var socketStore *SocketStore = NewSocketStore()

// WebSocketConnectionHandler upgrades a client connection to a WebSocket connection,
// regardless of what method is used in the request
func (hc *Context) WebSocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated (i.e. logged in)
	_, err := sessions.GetSessionID(r, hc.SessionIDKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get user information
	sessionState := &SessionState{}
	sessions.GetState(r, hc.SessionIDKey, hc.SessionStore, sessionState)
	user := sessionState.User

	// Upgrade the connection to a web socket connection
	if r.Header.Get("Origin") != "https://client.info441summary.me" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Websocket Connection Refused"))
		log.Println("Websocket Connection Refused")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Failed to open websocket connection"))
		log.Println("Failed to open websocket connectio")
		return
	}

	socketStore.Set(user.ID, conn)

	// Invoke a goroutine for handling control messages from this connection
	go (func(userID int64, conn *websocket.Conn) {
		defer conn.Close()
		defer socketStore.Delete(userID)

		for {
			messageType, data, err := conn.ReadMessage()
			if messageType == TextMessage || messageType == BinaryMessage {
				log.Printf("Client says %v\n", data)
				log.Printf("Writing %s to all sockets\n", string(data))
				if err := conn.WriteMessage(TextMessage, data); err != nil {
					log.Println("Error writing message to WebSocket connection.", err)
				}
			} else if messageType == CloseMessage {
				log.Println("Close message received.")
				break
			} else if err != nil {
				log.Println("Error reading message.")
				break
			}
		}
	})(user.ID, conn)
}

// ReadIncomingMessagesFromRabbit connects to a RabbitMQ server and starts a go
// routine that reads in new RabbitMQ messages and writes their contents to the
// correct WebSocket connections
func ReadIncomingMessagesFromRabbit() {
	// Connect to RabbitMQ server
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmqserver:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	// Open a RabbitMQ channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}

	// Connect to RabbitMQ Queue
	q, err := ch.QueueDeclare(
		"events", // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to declare a consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			msg.Ack(false)
			newMsg := &Message{}
			json.Unmarshal(msg.Body, newMsg)

			for userID, socketconn := range socketStore.Connections {
				// Write data to WebSocket connection
				if err := socketconn.WriteMessage(TextMessage, []byte(msg.Body)); err != nil {
					fmt.Println("Error writing message to WebSocket connection.", err)
				}
				// Case: The channel is private and user is a member OR the channel is public
				// AKA NOT(the channel is private and user is NOT a member)
				if !(len(newMsg.UserIDs) > 0 && !contains(userID, newMsg.UserIDs)) {
					data := []byte("Unrecognized Control Message Type")
					switch newMsg.Type {
					case "channel-new", "channel-update":
						data = []byte(newMsg.Channel)
					case "channel-delete":
						data = []byte(newMsg.ChannelID)
					case "message-new", "message-update":
						log.Println(newMsg.UserMessage)
						data = []byte(newMsg.UserMessage)
					case "message-delete":
						data = []byte(newMsg.UserMessageID)
					}

					// Write data to WebSocket connection
					if err := socketconn.WriteMessage(TextMessage, data); err != nil {
						fmt.Println("Error writing message to WebSocket connection.", err)
					}
				}
			}
		}
	}()

	// body := "Hello World!"
	// err = ch.Publish(
	// 	"",       // exchange
	// 	"events", // routing key
	// 	false,    // mandatory
	// 	false,    // immediate
	// 	amqp.Publishing{
	// 		ContentType: "text/plain",
	// 		Body:        []byte(body),
	// 	})
	// if err != nil {
	// 	log.Fatalf("Failed to publish: %v", err)
	// }
}

func contains(userID int64, userIDs []int64) bool {
	for _, ID := range userIDs {
		if ID == userID {
			return true
		}
	}
	return false
}
