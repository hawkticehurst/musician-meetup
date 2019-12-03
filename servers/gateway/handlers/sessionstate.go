package handlers

import (
	"serverside-final-project/servers/gateway/models/users"
	"time"
)

// SessionState struct to contain the information about SessionState
type SessionState struct {
	Time time.Time   `json:"time"`
	User *users.User `json:"user"`
}

// NewSessionState constructs a new SessionState struct
func NewSessionState(time time.Time, user *users.User) *SessionState {
	return &SessionState{time, user}
}
