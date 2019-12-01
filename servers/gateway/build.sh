#!/bin/bash
GOOS=linux go build
echo "✅  Linux Go Build Complete"
docker build -t stan9920/gatewayserver .
echo "✅  Local Gateway Docker Build Complete"
go clean
echo "✅  Linux Go Clean Complete"
