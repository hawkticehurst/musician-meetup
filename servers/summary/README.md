# Summary Microservice

This directory contains the source code for your existing page summary API refactored into a microservice. Create a simple HTTP server in Go, and move your existing handler in the `servers/gateway/handlers/summary.go` file to this server. Your gateway should forward requests for the existing search API to this new microservice instead of handling it in the gateway itself.
