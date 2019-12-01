package sessions

import (
	"errors"
	"net/http"
)

const headerAuthorization = "Authorization"
const paramAuthorization = "auth"
const schemeBearer = "Bearer "

// ErrNoSessionID is used when no session ID was found in the Authorization header
var ErrNoSessionID = errors.New("no session ID found in " + headerAuthorization + " header")

// ErrInvalidScheme is used when the authorization scheme is not supported
var ErrInvalidScheme = errors.New("authorization scheme not supported")

// BeginSession creates a new SessionID, saves the `sessionState` to the store, adds an
// Authorization header to the response with the SessionID, and returns the new SessionID
func BeginSession(signingKey string, store Store, sessionState interface{}, w http.ResponseWriter) (SessionID, error) {
	//TODO:
	//- create a new SessionID
	//- save the sessionState to the store
	//- add a header to the ResponseWriter that looks like this:
	//    "Authorization: Bearer <sessionID>"
	//  where "<sessionID>" is replaced with the newly-created SessionID
	//  (note the constants declared for you above, which will help you avoid typos)
	
	// Create new sessionID
	mySessionID, err := NewSessionID(signingKey)
	if err != nil {
		return InvalidSessionID, err
	}
	// Save sessionID and state to the store
	store.Save(mySessionID, sessionState)
	// Set authorization
	w.Header().Set(headerAuthorization, schemeBearer+string(mySessionID))

	return mySessionID, nil
}

// GetSessionID extracts and validates the SessionID from the request headers
func GetSessionID(r *http.Request, signingKey string) (SessionID, error) {
	//TODO: get the value of the Authorization header,
	//or the "auth" query string parameter if no Authorization header is present,
	//and validate it. If it's valid, return the SessionID. If not
	//return the validation error.

	// Check that authorization header is valid, if it is blank use the
	// auth query string parameter
	authHeader := r.Header.Get(headerAuthorization)
	if authHeader == "" {
		authParams, err := r.URL.Query()["auth"]

		if !err || len(authParams[0]) < 1 {
			return InvalidSessionID, ErrNoSessionID
		}
		authHeader = authParams[0]
	}

	authHeader = authHeader[7:]

	mySessionID, err := ValidateID(authHeader, signingKey)
	if err != nil {
		return InvalidSessionID, err
	}

	return mySessionID, nil
}

// GetState extracts the SessionID from the request,
// gets the associated state from the provided store into
// the `sessionState` parameter, and returns the SessionID
func GetState(r *http.Request, signingKey string, store Store, sessionState interface{}) (SessionID, error) {
	//TODO: get the SessionID from the request, and get the data
	//associated with that SessionID from the store.

	// Extract session id
	mySessionID, err := GetSessionID(r, signingKey)
	if err != nil {
		return InvalidSessionID, err
	}
	// Extract the session state
	err = store.Get(mySessionID, sessionState)
	if err != nil {
		return InvalidSessionID, err
	}

	return mySessionID, nil
}

// EndSession extracts the SessionID from the request,
// and deletes the associated data in the provided store, returning
// the extracted SessionID.
func EndSession(r *http.Request, signingKey string, store Store) (SessionID, error) {
	//TODO: get the SessionID from the request, and delete the
	//data associated with it in the store.

	// Extracts the sessionID from the http request
	mySessionID, err := GetSessionID(r, signingKey)
	if err != nil {
		return InvalidSessionID, err
	}
	// Deletes the associated data
	store.Delete(mySessionID)

	return mySessionID, nil
}
