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

docker push $DOCKERNAME/gatewayserver
docker push $DOCKERNAME/messagingserver
docker push $DOCKERNAME/meetupserver
docker push $DOCKERNAME/mysqldb
docker push $DOCKERNAME/gatewayserver
echo "✅  Push Docker Containers to DockerHub"

ssh -oStrictHostKeyChecking=no ec2-user@api.info441summary.me 'bash -s' < update-servers.sh $DOCKERNAME
echo "🎊  Server Deployment Complete!"