package handlers

import (
	"sync"

	"github.com/gorilla/websocket"
)

// SocketStore represents a map of userIDs to WebSocket connections that is
// safe for concurrent use
type SocketStore struct {
	Connections map[int64]*websocket.Conn
	mx          sync.RWMutex
}

// NewSocketStore constructs a new map of userIDs and WebSocket connections
func NewSocketStore() *SocketStore {
	return &SocketStore{
		Connections: map[int64]*websocket.Conn{},
	}
}

// Set adds a new userID and WebSocket connection to the map
func (c *SocketStore) Set(userID int64, wsConn *websocket.Conn) {
	c.mx.Lock()
	c.Connections[userID] = wsConn
	c.mx.Unlock()
}

// Get retrieves the WebSocket connection for a given userID
func (c *SocketStore) Get(userID int64) *websocket.Conn {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.Connections[userID]
}

// Delete removes the WebSocket connection for a given userID
func (c *SocketStore) Delete(userID int64) {
	c.mx.Lock()
	delete(c.Connections, userID)
	c.mx.Unlock()
}
