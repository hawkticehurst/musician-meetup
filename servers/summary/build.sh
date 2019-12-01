#!/bin/bash
GOOS=linux go build
echo "✅  Linux Go Build Complete"
docker build -t piercecave/summary .
echo "✅  Local Docker Build Complete"
go clean
echo "✅  Linux Go Clean Complete"
