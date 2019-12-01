#!/bin/bash
source ./build.sh
docker push piercecave/summary
echo "✅  Local Docker Push Complete"
ssh -oStrictHostKeyChecking=no ec2-user@server.info441summary.me 'bash -s' < update-server.sh
echo "🎊  Server Deployment Complete!"
read -p "Press any key..."
