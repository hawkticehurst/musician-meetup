#!/bin/bash

echo "Please enter your DockerHub username: "
read name
export DOCKERNAME=$name

docker build -t $DOCKERNAME/webclient .
echo "✅  Local Docker Build Complete"
docker rm -f webclient
echo "✅  Current Docker Container Stopped & Removed"
docker run -d --name webclient -p 443:443 -p 80:80 -v /etc/letsencrypt:/etc/letsencrypt:ro $DOCKERNAME/webclient
echo "✅  Updated Docker Container Successfully Running"
echo "🎊  Local Client Deployment Complete!"