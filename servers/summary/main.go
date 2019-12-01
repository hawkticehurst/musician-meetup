package main

import (
	"assignments-hawkticehurst/servers/summary/handlers"
	"log"
	"net/http"
	"os"
)

// main is the main entry point for the server
func main() {
	// Get the value of the ADDR environment variable
	addr := os.Getenv("ADDR")

	// If it's blank, default to port 80 for requests
	// addressed to any host
	if len(addr) == 0 {
		addr = ":80"
	}

	// Create a new mux (router)
	// The mux calls different functions for different
	// resource paths
	mux := http.NewServeMux()

	// Tell the mux to call the handlers.SummaryHandler
	// function when the "/v1/summary" URL path is requested
	mux.HandleFunc("/v1/summary", handlers.SummaryHandler)

	// Start the web server using the mux as the root
	// handler, and report any errors that occur.
	// The ListenAndServe() function is blocking so
	// this program will continue until killed.
	log.Printf("Server is listening at %s...", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
