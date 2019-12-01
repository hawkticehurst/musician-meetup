package handlers

import "net/http"

// CORSHeader is a middleware handler that adds CORS headers to a response
type CORSHeader struct {
	handler http.Handler
}

// NewCORSHeader constructs a new ResponseHeader middleware handler
func NewCORSHeader(handlerToWrap http.Handler) *CORSHeader {
	return &CORSHeader{handlerToWrap}
}

// ServeHTTP handles the request by adding CORS response headers
// If the request header method is OPTION return StatusOK so the
// actual client request can be sent
func (ch *CORSHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, PATCH, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Authorization")
	w.Header().Set("Access-Control-Max-Age", "600")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	ch.handler.ServeHTTP(w, r)
}
