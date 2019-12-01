#!/bin/bash

echo "Please enter your DockerHub username: "
read name
export DOCKERNAME=$name

cd ../gateway/
GOOS=linux go build
echo "✅  Linux Go Build Complete"
docker build -t $DOCKERNAME/gatewayserver .
echo "✅  Local Gateway Docker Build Complete"
go clean
echo "✅  Linux Go Clean Complete"
cd ../messaging/

docker build -t $DOCKERNAME/messagingserver .
docker build -t $DOCKERNAME/mysqldb ../db/
echo "✅  Local Messaging Server & MySQL Docker Builds Complete"
docker rm -f messagingserver
docker rm -f mysqlserver
docker rm -f gatewayserver
docker rm -f redisserver
echo "✅  Current Docker Containers Stopped & Removed"
docker network rm devnetwork
echo "✅  Current Docker Network Stopped & Removed"
export HOST="mysqlserver"
export PORT="3306"
export USER="root"
export MYSQL_ROOT_PASSWORD="testpwd"
export DATABASE="infodb"
export SESSIONKEY="key"
export REDISADDR="redisserver:6379"
export DSN="root:testpwd@tcp(mysqlserver:3306)/infodb"
# export TLSCERT=/etc/letsencrypt/fullchain.pem
# export TLSKEY=/etc/letsencrypt/privkey.pem
export TLSCERT=/etc/letsencrypt/fullchain.pem
export TLSKEY=/etc/letsencrypt/privkey.pem
echo "✅  Environment Variables Set"
docker network create devnetwork
echo "✅  Docker Network Created"
docker run -d --network devnetwork --name messagingserver -p 4000:80 -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e HOST=$HOST -e PORT=$PORT -e USER=$USER -e DATABASE=$DATABASE $DOCKERNAME/messagingserver
docker run -d --network devnetwork --name mysqlserver -e MYSQL_USER=$USER -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e MYSQL_DATABASE=$DATABASE $DOCKERNAME/mysqldb
# NOTE: -v /Users/hawk/letsencrypt:/etc/letsencrypt:ro means that the self signed certificates 
# are in folder in my home directory because absolute file paths are required in this flag
# So create a folder in your home directory called "letencrypt" and then run the command to create
# the self signed certs within that folder and make sure to change '/Users/hawk/letsencrypt' to 
# '<Your home directory>/letsencrypt'
docker run -d --network devnetwork --name gatewayserver -p 443:443 -v C:/Users/pierc/letsencrypt:/etc/letsencrypt:ro -e TLSCERT=$TLSCERT -e TLSKEY=$TLSKEY -e REDISADDR=$REDISADDR -e SESSIONKEY=$SESSIONKEY -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e DSN=$DSN $DOCKERNAME/gatewayserver
docker run -d --network devnetwork --name redisserver redis
echo "✅  Docker Containers Successfully Running"
read -p "Press any key..."