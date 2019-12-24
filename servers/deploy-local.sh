#!/bin/bash

echo "Please enter your DockerHub username: "
read name
export DOCKERNAME=$name

cd ./gateway/
GOOS=linux go build
echo "✅  Linux Go Build Complete"
docker build -t $DOCKERNAME/gatewayserver .
echo "✅  Local Gateway Docker Build Complete"
go clean
echo "✅  Linux Go Clean Complete"
cd ./../

docker build -t $DOCKERNAME/messagingserver ./messaging/
docker build -t $DOCKERNAME/meetupserver ./meetup/
docker build -t $DOCKERNAME/mysqldb ./db/
echo "✅  Local Docker Builds Complete"

docker rm -f messagingserver
docker rm -f meetupserver
docker rm -f mysqlserver
docker rm -f gatewayserver
docker rm -f redisserver
docker rm -f rabbitmqserver
echo "✅  Current Docker Containers Stopped & Removed"

docker volume rm $(docker volume ls -qf dangling=true)
echo "✅  Docker Volumes Removed"

docker network rm backendnetwork
echo "✅  Current Docker Network Stopped & Removed"

export HOST="mysqlserver"
export PORT="3306"
export USER="root"
export MYSQL_ROOT_PASSWORD="testpwd"
export DATABASE="infodb"
export SESSIONKEY="key"
export REDISADDR="redisserver:6379"
export MESSAGESADDR="messagingserver"
export MEETUPADDR="meetupserver"
export DSN="root:testpwd@tcp(mysqlserver:3306)/infodb"
export TLSCERT=/etc/letsencrypt/live/api.info441summary.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/api.info441summary.me/privkey.pem
echo "✅  Environment Variables Set"

docker network create -d bridge backendnetwork
echo "✅  Docker Network Created"

docker run -d --network backendnetwork --hostname my-rabbit --name rabbitmqserver rabbitmq:3-management
docker run -d --network backendnetwork --name messagingserver --restart unless-stopped -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e HOST=$HOST -e PORT=$PORT -e USER=$USER -e DATABASE=$DATABASE $DOCKERNAME/messagingserver
docker run -d --network backendnetwork --name meetupserver --restart unless-stopped -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e HOST=$HOST -e PORT=$PORT -e USER=$USER -e DATABASE=$DATABASE $DOCKERNAME/meetupserver
docker run -d --network backendnetwork --name mysqlserver -e MYSQL_USER=$USER -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e MYSQL_DATABASE=$DATABASE $DOCKERNAME/mysqldb
docker run -d --network backendnetwork --name gatewayserver --restart unless-stopped -p 443:443 -v /etc/letsencrypt:/etc/letsencrypt:ro -e TLSCERT=$TLSCERT -e TLSKEY=$TLSKEY -e REDISADDR=$REDISADDR -e MESSAGESADDR=$MESSAGESADDR -e MEETUPADDR=$MEETUPADDR -e SESSIONKEY=$SESSIONKEY -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e DSN=$DSN $DOCKERNAME/gatewayserver
docker run -d --network backendnetwork --name redisserver redis
echo "✅  Docker Containers Successfully Running"
echo "🎊  Server Deployment Complete!"