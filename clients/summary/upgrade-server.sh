#!/bin/bash
docker rm -f web_client
echo "✅  Current Docker Container Stopped & Removed"
docker pull piercecave/web_client
echo "✅  Lastest Docker Image Pulled To Server"
docker run -d --name web_client -p 443:443 -p 80:80 -v /etc/letsencrypt:/etc/letsencrypt:ro piercecave/web_client
echo "✅  Updated Docker Container Successfully Running"