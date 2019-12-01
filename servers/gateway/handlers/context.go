package handlers

import (
	"assignments-hawkticehurst/servers/gateway/models/users"
	"assignments-hawkticehurst/servers/gateway/sessions"
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
