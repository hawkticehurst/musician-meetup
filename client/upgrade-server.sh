#!/bin/bash

export DOCKERNAME=$1

docker rm -f webclient
echo "✅  Current Docker Container Stopped & Removed"
docker pull $DOCKERNAME/webclient
echo "✅  Lastest Docker Image Pulled To Server"
docker run -d --name webclient -p 443:443 -p 80:80 -v /etc/letsencrypt:/etc/letsencrypt:ro $DOCKERNAME/webclient
echo "✅  Updated Docker Container Successfully Running"