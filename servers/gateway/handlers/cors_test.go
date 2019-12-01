package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCORSHeader(t *testing.T) {
	handler := NewCORSHeader(http.HandlerFunc(MockHandler))

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "https://example.com/v1/mock", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the content-type is what we expect
	if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
		t.Errorf("Content type header does not match: got %v want %v", ctype, "application/json")
	}

	// Check the Access-Control-Allow-Origin is what we expect
	if ctype := rr.Header().Get("Access-Control-Allow-Origin"); ctype != "*" {
		t.Errorf("Access-Control-Allow-Origin header does not match: got %v want %v", ctype, "*")
	}

	// Check the Access-Control-Allow-Methods is what we expect
	if ctype := rr.Header().Get("Access-Control-Allow-Methods"); ctype != "GET, PUT, POST, PATCH, DELETE" {
		t.Errorf("Access-Control-Allow-Methods header does not match: got %v want %v", ctype, "GET, PUT, POST, PATCH, DELETE")
	}

	// Check the Access-Control-Allow-Headers is what we expect
	if ctype := rr.Header().Get("Access-Control-Allow-Headers"); ctype != "Content-Type, Authorization" {
		t.Errorf("Access-Control-Allow-Headers header does not match: got %v want %v", ctype, "Content-Type, Authorization")
	}

	// Check the Access-Control-Expose-Headers is what we expect
	if ctype := rr.Header().Get("Access-Control-Expose-Headers"); ctype != "Authorization" {
		t.Errorf("Access-Control-Expose-Headers header does not match: got %v want %v", ctype, "Authorization")
	}

	// Check the Access-Control-Max-Age is what we expect
	if ctype := rr.Header().Get("Access-Control-Max-Age"); ctype != "600" {
		t.Errorf("Access-Control-Max-Age header does not match: got %v want %v", ctype, "600")
	}

	// Check to make sure the response body is what we expect and has not been
	// manipulated by the CORS handler
	// (THIS SHOULD NEVER THROW AN ERROR BECAUSE CORS MIDDLEWARE DOES NOT MANIPULATE THE BODY)
	expected := `{"mockHandler": true}`
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

// MockHandler is a fake handler function for the purposes of testing
// Since we only care about testing the CORS middleware the contents/functionality
// of the fake handler does not matter
func MockHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	io.WriteString(w, `{"mockHandler": true}`)
}
