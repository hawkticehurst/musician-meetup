package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"serverside-final-project/servers/gateway/models/users"
	"serverside-final-project/servers/gateway/sessions"
	"strconv"
	"strings"
	"time"
)

// UsersHandler creates new user accounts
func (hc *Context) UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-type")
		if strings.HasPrefix(contentType, "application/json") {
			newUser := &users.NewUser{}
			dec := json.NewDecoder(r.Body)
			if err := dec.Decode(newUser); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			err := newUser.Validate()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			user, _ := newUser.ToUser()
			insertedUser, err := hc.UserStore.Insert(user)
			if err != nil {
				fmt.Printf("Error inserting user into database: %v\n", err)
			}
			sessionState := NewSessionState(time.Now(), insertedUser)
			_, sessionErr := sessions.BeginSession(hc.SessionIDKey, hc.SessionStore, sessionState, w)
			if sessionErr != nil {
				fmt.Printf("Error creating session: %v\n", err)
				return
			}

			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			userJSON, _ := json.Marshal(insertedUser)
			w.Write(userJSON)
		} else {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte("Request body must be in JSON"))
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// SpecificUserHandler handle requests for specific user
func (hc *Context) SpecificUserHandler(w http.ResponseWriter, r *http.Request) {
	_, err := sessions.GetSessionID(r, hc.SessionIDKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		URL := r.URL.RequestURI()
		i := strings.LastIndex(URL, "/")
		var idValue int64
		var UserID string = URL[i+1 : len(URL)]
		if UserID == "me" {
			sessionState := &SessionState{}
			sessions.GetState(r, hc.SessionIDKey, hc.SessionStore, sessionState)
			user := sessionState.User
			idValue = user.ID
		} else {
			idValue, _ = strconv.ParseInt(UserID, 10, 64)
		}
		user, err := hc.UserStore.GetByID(idValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		userJSON, _ := json.Marshal(user)
		w.Write(userJSON)
	} else if r.Method == http.MethodPatch {
		URL := r.URL.RequestURI()
		i := strings.LastIndex(URL, "/")
		UserID := URL[i+1 : len(URL)]
		sessionState := &SessionState{}
		sessions.GetState(r, hc.SessionIDKey, hc.SessionStore, sessionState)
		if UserID != "me" {
			userID, _ := strconv.ParseInt(UserID, 10, 64)
			if userID != sessionState.User.ID {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Invalid UserID"))
				return
			}
		}
		contentType := r.Header.Get("Content-type")
		if !strings.HasPrefix(contentType, "application/json") {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte("Request body must be in JSON"))
			return
		}
		updates := &users.Updates{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(updates); err != nil {
			fmt.Printf("error decoding JSON: %v\n", err)
			return
		}
		sessionState.User.ApplyUpdates(updates)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		userJSON, _ := json.Marshal(sessionState.User)
		w.Write(userJSON)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// SessionsHandler handles requests for the "sessions" resource, and
// allows clients to begin a new session using an existing user's credentials.
func (hc *Context) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			credentials := &users.Credentials{}
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(credentials); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			user, err := hc.UserStore.GetByEmail(credentials.Email)
			if err != nil {
				time.Sleep(time.Second)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("invalid credentials"))
				return
			}

			if user.Authenticate(credentials.Password) != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("invalid credentials"))
				return
			}

			sessionState := NewSessionState(time.Now(), user)
			sessions.BeginSession(hc.SessionIDKey, hc.SessionStore, sessionState, w)

			clientIP := r.Header.Get("X-Forwarded-For")
			if len(clientIP) != 0 {
				ipList := strings.Split(clientIP, ", ")
				clientIP = ipList[0]
			} else {
				clientIP = r.RemoteAddr
			}

			hc.UserStore.LogUser(user.ID, sessionState.Time, clientIP)

			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			userJSON, _ := json.Marshal(user)
			w.Write(userJSON)
		} else {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte("Request body must be in JSON"))
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// SpecificSessionHandler handles requests related to a specific
// authenticated session
func (hc *Context) SpecificSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		pathElements := strings.Split(r.URL.RequestURI(), "/")
		if pathElements[len(pathElements)-1] != "mine" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Last element of URL must be mine, got " + pathElements[len(pathElements)-1]))
			return
		}
		sessions.EndSession(r, hc.SessionIDKey, hc.SessionStore)
		w.Write([]byte("Signed out"))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
