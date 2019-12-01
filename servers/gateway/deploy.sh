#!/bin/bash
source build.sh
docker push stan9920/gatewayserver
echo "✅  Local Gateway Docker Push Complete"
docker build -t stan9920/mysqldb ../db/
echo "✅  Local MySQL Docker Build Complete"
docker push stan9920/mysqldb
echo "✅  MySQL Docker Push Complete"
ssh -oStrictHostKeyChecking=no ec2-user@server.info441summary.me 'bash -s' < update-server.sh
echo "🎊  Server Deployment Complete!"
