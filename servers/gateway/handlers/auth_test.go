package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"serverside-final-project/servers/gateway/models/users"
	"serverside-final-project/servers/gateway/sessions"
	"testing"
	"time"
)

//Set request content type header to wrong type and look for
//status code http.StatusUnsupportedMediaType to be returned
func TestWrongContentTypeUsersHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "/v1/users", nil)
	req.Header.Set("Content-Type", "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	handler := http.HandlerFunc(context.UsersHandler)
	handler.ServeHTTP(rr, req)
	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusUnsupportedMediaType {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnsupportedMediaType)
	}
}

//Set body to non json content and a bad request
//status should be returned
func TestNonJSONFormatUsersHandler(t *testing.T) {
	r := bytes.NewReader([]byte("hello"))
	req, err := http.NewRequest("POST", "/v1/users", r)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	handler := http.HandlerFunc(context.UsersHandler)
	handler.ServeHTTP(rr, req)
	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

//Set json body with an invalid user and
//validation error should be returned
func TestInvalidUserUsersHandler(t *testing.T) {
	newUser := &users.NewUser{Email: "s", Password: "12", PasswordConf: "123", UserName: "s", FirstName: "Stanley", LastName: "Wu"}
	buffer, err := json.Marshal(newUser)
	r := bytes.NewReader(buffer)
	req, err := http.NewRequest("POST", "/v1/users", r)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	handler := http.HandlerFunc(context.UsersHandler)
	handler.ServeHTTP(rr, req)
	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

//If json body has a valid new user, check that
//user is inserted into the database, a new session was created,
//application/json is set, and a status 201 is returned
func TestValidNewUserUsersHandler(t *testing.T) {
	newUser := &users.NewUser{Email: "stanley@gmail.com", Password: "123456", PasswordConf: "123456", UserName: "swu", FirstName: "Stanley", LastName: "Wu"}
	buffer, err := json.Marshal(newUser)
	r := bytes.NewReader(buffer)
	req, err := http.NewRequest("POST", "/v1/users", r)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	handler := http.HandlerFunc(context.UsersHandler)
	handler.ServeHTTP(rr, req)
	// Check that StatusCreated code is returned
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
	// Check that application/json is set as content-type
	if content := rr.Header().Get("Content-type"); content != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			content, "application/json")
	}
	// Check response body is correct by comparing user input into
	// request with response body values
	user, _ := newUser.ToUser()
	bodyUser := &users.User{}
	buff := []byte(rr.Body.String())
	if err := json.Unmarshal(buff, bodyUser); err != nil {
		fmt.Printf("error unmarshaling JSON: %v\n", err)
	}
	// Cannot check email and passhash because they are omitted when encoded to json
	if bodyUser.UserName != user.UserName || bodyUser.FirstName != user.FirstName || bodyUser.LastName != user.LastName || bodyUser.PhotoURL != user.PhotoURL {
		t.Errorf("Error with response body")
	}
	// Check the testuserstore to see that hardcoded value equals return inputted values
	testuserstore := &users.TestUserStore{Client: "client"}
	hardCodedUser, _ := testuserstore.GetByEmail("stanley@gmail.com")
	if hardCodedUser.Email != user.Email || hardCodedUser.FirstName != user.FirstName || hardCodedUser.LastName != user.LastName ||
		hardCodedUser.PhotoURL != user.PhotoURL || hardCodedUser.UserName != user.UserName {
		t.Errorf("Database returned wrong information")
	}
}

//If HTTP method besides POST is used,
//a statusmethodnotallowed error should occur
func TestWrongHTTPMethodUsersHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/v1/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	handler := http.HandlerFunc(context.UsersHandler)
	handler.ServeHTTP(rr, req)
	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

//Test that user cannot access this handler
//if not logged in
func TestNotAuthenticatedSpecificUserHandler(t *testing.T) {
	//There is no session header passed in to simulate
	//user not signed in
	req, err := http.NewRequest("GET", "/v1/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	handler := http.HandlerFunc(context.SpecificUserHandler)
	handler.ServeHTTP(rr, req)
	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

//If no user is found with given id,
//a statusnotfound code should be returned
func TestBadIDSpecificUserHandler(t *testing.T) {
	Context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	rr, context := CreateNewUser(Context)

	req, err := http.NewRequest("GET", "/v1/users/2", nil)
	if err != nil {
		t.Fatal(err)
	}
	//Get the response authorization header and put into request header
	//to simulate user being signed in
	req.Header.Set("Authorization", rr.Header().Get("Authorization"))
	rrTwo := httptest.NewRecorder()
	handler := http.HandlerFunc(context.SpecificUserHandler)
	handler.ServeHTTP(rrTwo, req)
	// Check the status code is what we expect.
	if status := rrTwo.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

//If user id is correct (using 1), a user struct should be returned
func TestValidUserIDNumberSpecificUserHandler(t *testing.T) {
	Context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	rr, context := CreateNewUser(Context)

	req, err := http.NewRequest("GET", "/v1/users/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	//Get the response authorization header and put into request header
	//to simulate user being signed in
	req.Header.Set("Authorization", rr.Header().Get("Authorization"))
	rrTwo := httptest.NewRecorder()
	handler := http.HandlerFunc(context.SpecificUserHandler)
	handler.ServeHTTP(rrTwo, req)
	// Check the status code is what we expect.
	if status := rrTwo.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	// Check that application/json is set as content-type
	if content := rrTwo.Header().Get("Content-type"); content != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			content, "application/json")
	}
	// Check response body is correct by comparing user input into
	// request with response body values
	newUser := &users.NewUser{Email: "stanley@gmail.com", Password: "123456", PasswordConf: "123456", UserName: "swu", FirstName: "Stanley", LastName: "Wu"}
	user, _ := newUser.ToUser()
	bodyUser := &users.User{}
	buff := []byte(rrTwo.Body.String())
	if err := json.Unmarshal(buff, bodyUser); err != nil {
		fmt.Printf("error unmarshaling JSON: %v\n", err)
	}
	// Cannot check email and passhash because they are omitted when encoded to json
	if bodyUser.UserName != user.UserName || bodyUser.FirstName != user.FirstName || bodyUser.LastName != user.LastName || bodyUser.PhotoURL != user.PhotoURL {
		t.Errorf("Error with response body")
	}
}

//If user id is correct (using "me"), a user struct should be returned
func TestValidUserIDMeSpecificUserHandler(t *testing.T) {
	Context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	rr, context := CreateNewUser(Context)

	req, err := http.NewRequest("GET", "/v1/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	//Get the response authorization header and put into request header
	//to simulate user being signed in
	req.Header.Set("Authorization", rr.Header().Get("Authorization"))
	rrTwo := httptest.NewRecorder()
	handler := http.HandlerFunc(context.SpecificUserHandler)
	handler.ServeHTTP(rrTwo, req)
	// Check the status code is what we expect.
	if status := rrTwo.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	// Check that application/json is set as content-type
	if content := rrTwo.Header().Get("Content-type"); content != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			content, "application/json")
	}
	// Check response body is correct by comparing user input into
	// request with response body values
	newUser := &users.NewUser{Email: "stanley@gmail.com", Password: "123456", PasswordConf: "123456", UserName: "swu", FirstName: "Stanley", LastName: "Wu"}
	user, _ := newUser.ToUser()
	bodyUser := &users.User{}
	buff := []byte(rrTwo.Body.String())
	if err := json.Unmarshal(buff, bodyUser); err != nil {
		fmt.Printf("error unmarshaling JSON: %v\n", err)
	}
	// Cannot check email and passhash because they are omitted when encoded to json
	if bodyUser.UserName != user.UserName || bodyUser.FirstName != user.FirstName || bodyUser.LastName != user.LastName || bodyUser.PhotoURL != user.PhotoURL {
		t.Errorf("Error with response body")
	}
}

//Test if the user ID in the request URL is not "me" or does not
//match the currently-authenticated user, respond with StatusForbidden (403)
func TestUnauthenticatedUserIDSpecificUserHandler(t *testing.T) {
	Context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	rr, context := CreateNewUser(Context)

	//Authenticated userID is 1, test with UserID 2
	req, err := http.NewRequest("PATCH", "/v1/users/2", nil)
	if err != nil {
		t.Fatal(err)
	}
	//Get the response authorization header and put into request header
	//to simulate user being signed in
	req.Header.Set("Authorization", rr.Header().Get("Authorization"))
	rrTwo := httptest.NewRecorder()
	handler := http.HandlerFunc(context.SpecificUserHandler)
	handler.ServeHTTP(rrTwo, req)
	// Check the status code is what we expect.
	if status := rrTwo.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusForbidden)
	}
}

//Test if header does not start with application/json,
//return http.StatusUnsupportedMediaType (415)
func TestWrongContentTypeSpecificUserHandler(t *testing.T) {
	Context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	rr, context := CreateNewUser(Context)

	req, err := http.NewRequest("PATCH", "/v1/users/1", nil)
	req.Header.Set("Content-Type", "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	//Get the response authorization header and put into request header
	//to simulate user being signed in
	req.Header.Set("Authorization", rr.Header().Get("Authorization"))
	rrTwo := httptest.NewRecorder()
	handler := http.HandlerFunc(context.SpecificUserHandler)
	handler.ServeHTTP(rrTwo, req)
	// Check the status code is what we expect.
	if status := rrTwo.Code; status != http.StatusUnsupportedMediaType {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnsupportedMediaType)
	}
}

//Test that status 200 is returned if input is an updates
//struct that can update user's profile
func TestUpdateSpecificUserHandler(t *testing.T) {
	Context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	rr, context := CreateNewUser(Context)

	update := &users.Updates{FirstName: "John", LastName: "Smith"}
	buffer, _ := json.Marshal(update)
	r := bytes.NewReader(buffer)
	req, err := http.NewRequest("PATCH", "/v1/users/me", r)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}
	//Get the response authorization header and put into request header
	//to simulate user being signed in
	req.Header.Set("Authorization", rr.Header().Get("Authorization"))
	rrTwo := httptest.NewRecorder()
	handler := http.HandlerFunc(context.SpecificUserHandler)
	handler.ServeHTTP(rrTwo, req)
	// Check that StatusCreated code is returned
	if status := rrTwo.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	// Check that application/json is set as content-type
	if content := rrTwo.Header().Get("Content-type"); content != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			content, "application/json")
	}
	// Check response body is correctly updated
	// with update values
	updateUser := &users.Updates{}
	buff := []byte(rrTwo.Body.String())
	if err := json.Unmarshal(buff, updateUser); err != nil {
		fmt.Printf("error unmarshaling JSON: %v\n", err)
	}
	// Check that first and last name were updated
	if update.FirstName != updateUser.FirstName || update.LastName != updateUser.LastName {
		t.Errorf("Error with response body")
	}
}

//If HTTP method besides GET and PATCH is used,
//a statusmethodnotallowed error should occur
func TestWrongHTTPMethodSpecificUserHandler(t *testing.T) {
	Context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	rr, context := CreateNewUser(Context)

	req, err := http.NewRequest("POST", "/v1/users/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	//Get the response authorization header and put into request header
	//to simulate user being signed in
	req.Header.Set("Authorization", rr.Header().Get("Authorization"))
	rrTwo := httptest.NewRecorder()
	handler := http.HandlerFunc(context.SpecificUserHandler)
	handler.ServeHTTP(rrTwo, req)
	// Check the status code is what we expect.
	if status := rrTwo.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

//Create new user to simulate being signed in
func CreateNewUser(context *Context) (http.ResponseWriter, *Context) {
	newUser := &users.NewUser{Email: "stanley@gmail.com", Password: "123456", PasswordConf: "123456", UserName: "swu", FirstName: "Stanley", LastName: "Wu"}
	buffer, _ := json.Marshal(newUser)
	r := bytes.NewReader(buffer)
	req, _ := http.NewRequest("POST", "/v1/users", r)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(context.UsersHandler)
	handler.ServeHTTP(rr, req)
	return rr, context
}

func TestSessionsHandlerMethodType(t *testing.T) {

	// Test if we pass in a GET instead of a POST method
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	handler := http.HandlerFunc(context.SessionsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestSessionsHandlerContentType(t *testing.T) {

	// Test we get http.StatusUnsupportedMediaType error when we pass in wrong Content-Type
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "not/application/json")

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))
	handler := http.HandlerFunc(context.SessionsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnsupportedMediaType {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestSessionsHandlerCorrectInput(t *testing.T) {

	rr := httptest.NewRecorder()

	userCredentials := &users.Credentials{Email: "stanley@gmail.com", Password: "123456"}
	buffer, err := json.Marshal(userCredentials)
	r := bytes.NewReader(buffer)

	req, err := http.NewRequest("POST", "", r)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))

	handler := http.HandlerFunc(context.SessionsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Check that application/json is set as content-type
	if content := rr.Header().Get("Content-type"); content != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			content, "application/json")
	}
	// Check response body is correct by comparing user input into
	// request with response body values
	newUser := &users.NewUser{Email: "stanley@gmail.com", Password: "123456", PasswordConf: "123456", UserName: "swu", FirstName: "Stanley", LastName: "Wu"}
	user, _ := newUser.ToUser()

	bodyUser := &users.User{}
	buff := []byte(rr.Body.String())
	if err := json.Unmarshal(buff, bodyUser); err != nil {
		fmt.Printf("error unmarshaling JSON: %v\n", err)
	}

	if bodyUser.UserName != user.UserName || bodyUser.FirstName != user.FirstName || bodyUser.LastName != user.LastName || bodyUser.PhotoURL != user.PhotoURL {
		t.Errorf("Error with response body")
	}
}

func TestSessionsHandlerBadCredentialsStruct(t *testing.T) {

	rr := httptest.NewRecorder()

	// Not a json struct
	buffer := []byte("hi")
	r := bytes.NewReader(buffer)

	req, err := http.NewRequest("POST", "", r)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))

	handler := http.HandlerFunc(context.SessionsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestSessionsHandlerBadCredentials(t *testing.T) {

	rr := httptest.NewRecorder()

	// Non-existing user
	userCredentials := &users.Credentials{Email: "sanley@gmail.com", Password: "123456"}
	buffer, err := json.Marshal(userCredentials)
	r := bytes.NewReader(buffer)

	req, err := http.NewRequest("POST", "", r)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))

	handler := http.HandlerFunc(context.SessionsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

func TestSessionsHandlerBadPassword(t *testing.T) {

	rr := httptest.NewRecorder()

	// Incorrect Password
	userCredentials := &users.Credentials{Email: "stanley@gmail.com", Password: "hi"}
	buffer, err := json.Marshal(userCredentials)
	r := bytes.NewReader(buffer)

	req, err := http.NewRequest("POST", "", r)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))

	handler := http.HandlerFunc(context.SessionsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

func TestSessionsHandlerIPAddress(t *testing.T) {

	rr := httptest.NewRecorder()

	// Incorrect Password
	userCredentials := &users.Credentials{Email: "stanley@gmail.com", Password: "123456"}
	buffer, err := json.Marshal(userCredentials)
	r := bytes.NewReader(buffer)

	req, err := http.NewRequest("POST", "", r)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "1,2")

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))

	handler := http.HandlerFunc(context.SessionsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
}

func TestSpecificSessionHandlerCorrectInput(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodDelete, "/v1/sessions/mine", nil)
	if err != nil {
		t.Fatal(err)
	}

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))

	handler := http.HandlerFunc(context.SpecificSessionHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Body.String(); status != "Signed out" {
		t.Errorf("handler returned wrong status message: got %v want %v",
			status, "Signed out")
	}
}

func TestSpecificSessionHandlerBadMethodBadURL(t *testing.T) {
	rr := httptest.NewRecorder()

	// Method is not DELETE
	req, err := http.NewRequest(http.MethodPost, "/v1/sessions/mine", nil)
	if err != nil {
		t.Fatal(err)
	}

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))

	handler := http.HandlerFunc(context.SpecificSessionHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status message: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestSpecificSessionHandlerBadURL(t *testing.T) {
	rr := httptest.NewRecorder()

	// Method is not DELETE
	req, err := http.NewRequest(http.MethodDelete, "/v1/sessions/notmine", nil)
	if err != nil {
		t.Fatal(err)
	}

	context := NewContext("key", sessions.NewMemStore(3*time.Minute, 3*time.Minute), users.NewTestUserStore("client"))

	handler := http.HandlerFunc(context.SpecificSessionHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status message: got %v want %v",
			status, http.StatusForbidden)
	}
}
