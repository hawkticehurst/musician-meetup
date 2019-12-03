package handlers

import (
	"assignments-hawkticehurst/servers/gateway/models/users"
	"assignments-hawkticehurst/servers/gateway/sessions"
	"net/http"
)

// Context struct to contain the information about the context
type Context struct {
	SessionIDKey string         `json:"sessionIDKey"`
	SessionStore sessions.Store `json:"sessionStore"`
	UserStore    users.Store    `json:"userStore"`
}

// NewContext constructs a new Context struct,
// ensuring that the dependencies are valid values
func NewContext(sessionIDKey string, sessionStore sessions.Store, userStore users.Store) *Context {
	if sessionStore == nil {
		panic("nil Redis session")
	}
	if userStore == nil {
		panic("nil MySQL session")
	}
	return &Context{sessionIDKey, sessionStore, userStore}
}

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		// This function's purpose is to reject websocket upgrade requests if the
// 		// origin of the websockete handshake request is coming from unknown domains.
// 		// This prevents some random domain from opening up a socket with your server.
// 		// TODO: make sure you modify this for your HW to check if r.Origin is your host
// 		return true
// 	},
// }

// XXXWebSocketConnectionHandler upgrades a client connection to a WebSocket connection,
// regardless of what method is used in the request
func (hc *Context) XXXWebSocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// handle the websocket handshake
	if r.Header.Get("Origin") != "https://client.info441summary.me" {
		http.Error(w, "Websocket Connection Refused", 403)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to open websocket connection", 401)
	}
	// do something with connection
	conn.WriteMessage(1, []byte("Hello jackass\n"))
	w.Write([]byte("Made connection"))
}
