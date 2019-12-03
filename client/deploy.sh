#!/bin/bash

docker build -t piercecave/web_client .
echo "✅  Local Docker Build Complete"
docker login
docker push piercecave/web_client
echo "✅  Local Docker Push Complete"
ssh -oStrictHostKeyChecking=no ec2-user@client.info441summary.me 'bash -s' < upgrade-server.sh 
echo "🎊  Client Deployment Complete!"
read -p "Press any key..."