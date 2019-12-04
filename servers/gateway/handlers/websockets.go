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
		// This function's purpose is to reject websocket upgrade requests if the
		// origin of the websockete handshake request is coming from unknown domains.
		// This prevents some random domain from opening up a socket with your server.
		// TODO: make sure you modify this for your HW to check if r.Origin is your host

		return true
	},
}

// Data structure containing every current websocket connection
var socketStore *SocketStore = NewSocketStore()

// WebSocketConnectionHandler upgrades a client connection to a WebSocket connection,
// regardless of what method is used in the request
func (hc *Context) WebSocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Break 1, before getsessionid")

	// Check if user is authenticated (i.e. logged in)
	_, err := sessions.GetSessionID(r, hc.SessionIDKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Println("Websockets could not get session id")
		return
	}

	log.Println("Break 2, before getstate")

	// Get user information
	sessionState := &SessionState{}
	sessions.GetState(r, hc.SessionIDKey, hc.SessionStore, sessionState)
	user := sessionState.User

	log.Println("Break 3, before get origin")
	log.Printf("Origin Header in websocket.go: %s", r.Header.Get("Origin"))

	// Upgrade the connection to a web socket connection
	if r.Header.Get("Origin") != "https://client.info441summary.me" {
		// http.Error(w, "Websocket Connection Refused", 403)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Websocket Connection Refused"))
		log.Println("Websocket Connection Refused")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//http.Error(w, "Failed to open websocket connection", 401)
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
	//defer conn.Close()
	log.Println("[AMQP] Connection Opened")

	// Open a RabbitMQ channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	//defer ch.Close()
	log.Println("[AMQP] Channel Opened")

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
	log.Println("[AMQP] Queue Declared")
	log.Println(q.Name)

	// messages, err := ch.Consume("events", "", true, false, false, false, nil)
	// if err != nil {
	// 	log.Fatalf("Failed to declare a consumer: %v", err)
	// }

	// go func() {
	// 	log.Println("FIRST CONSUME: I'm inside the go routine for messages")
	// 	for mymessage := range messages {
	// 		log.Println("FIRST CONSUME: HI")
	// 		log.Printf("Received a message: %s", mymessage.Body)
	// 	}
	// }()

	msgs, err := ch.Consume(
		"events", // queue
		"",       // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		log.Fatalf("Failed to declare a consumer: %v", err)
	}
	log.Println("[AMQP] Consumer Declared")

	go func() {
		for msg := range msgs {
			log.Println("HELLO")
			log.Printf("Received a message: %s", msg.Body)
			log.Println("Delivery:")

			msg.Ack(false)
			newMsg := &Message{}
			json.Unmarshal(msg.Body, newMsg)

			for userID, conn := range socketStore.Connections {
				// Write data to WebSocket connection
				if err := conn.WriteMessage(TextMessage, []byte(msg.Body)); err != nil {
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
						log.Println("New Message: ")
						log.Println(newMsg.UserMessage)
						data = []byte(newMsg.UserMessage)
					case "message-delete":
						data = []byte(newMsg.UserMessageID)
					}

					// Write data to WebSocket connection
					if err := conn.WriteMessage(TextMessage, data); err != nil {
						fmt.Println("Error writing message to WebSocket connection.", err)
					}
				}
			}
		}
	}()

	body := "Hello World!"
	err = ch.Publish(
		"",       // exchange
		"events", // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
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
