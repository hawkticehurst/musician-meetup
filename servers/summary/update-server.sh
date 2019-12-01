#!/bin/bash
docker rm -f summary
echo "✅  Current Docker Container Stopped & Removed"
docker pull piercecave/summary
echo "✅  Lastest Docker Image Pulled To Server"
docker run -d --name summary piercecave/summary
echo "✅  Updated Docker Container Successfully Running"
