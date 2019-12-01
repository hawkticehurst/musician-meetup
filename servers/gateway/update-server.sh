#!/bin/bash
docker rm -f gatewayserver
docker rm -f mysqlserver
docker rm -f redisserver
echo "✅  Current Docker Containers Stopped & Removed"
docker network rm info441network
echo "✅  Current Docker Network Stopped & Removed"
docker pull stan9920/gatewayserver
docker pull stan9920/mysqldb
echo "✅  Lastest Docker Images Pulled To Server"
export TLSCERT=/etc/letsencrypt/live/server.info441summary.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/server.info441summary.me/privkey.pem
export MYSQL_ROOT_PASSWORD="testpwd"
export SESSIONKEY="key"
export REDISADDR="redisserver:6379"
export DSN="root:testpwd@tcp(mysqlserver:3306)/infodb"
echo "✅  Environment Variables Set"
docker network create -d bridge info441network
echo "✅  Docker network created"
docker run -d --network info441network --name gatewayserver -p 443:443 -v /etc/letsencrypt:/etc/letsencrypt:ro -e TLSCERT=$TLSCERT -e TLSKEY=$TLSKEY -e REDISADDR=$REDISADDR -e SESSIONKEY=$SESSIONKEY -e DSN=$DSN stan9920/gatewayserver
docker run -d --network info441network --name mysqlserver -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e MYSQL_DATABASE=infodb stan9920/mysqldb
docker run -d --network info441network --name redisserver redis
echo "✅  Updated Docker Container Successfully Running"

