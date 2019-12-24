#!/bin/bash

echo "Please enter your DockerHub username: "
read name
export DOCKERNAME=$name

docker build -t $DOCKERNAME/webclient .
echo "✅  Local Docker Build Complete"
docker login
docker push $DOCKERNAME/webclient
echo "✅  Local Docker Push Complete"
ssh -oStrictHostKeyChecking=no ec2-user@client.info441summary.me 'bash -s' < upgrade-server.sh $DOCKERNAME
echo "🎊  Client Deployment Complete!"