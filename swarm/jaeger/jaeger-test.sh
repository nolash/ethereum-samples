#!/bin/bash

case $1 in
	"start")
		docker network create --internal --subnet 10.1.2.0/24 swarminternal
		docker run -d --ip 10.1.2.2 --network swarminternal --rm --name jaeger -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 -p 16686:16686 jaegertracing/all-in-one:1.8
		docker run -d --ip 10.1.2.11 --network swarminternal --rm -e PATH=/:/bin:/usr/bin:/usr/local/bin -e PASSWORD=tralala --name bzz1 -h bz1 swarm:trace --tracing --tracing.endpoint 10.1.2.2:6831 --sync-update-delay 1s --bzzport 80 --httpaddr 0.0.0.0 --tracing.svc "bz1"
		# wait for nodes to start
		sleep 10
		ENODE=`docker exec bzz1 /geth attach /root/.ethereum/bzzd.ipc --exec admin.nodeInfo.enode`
		ENODENEW=`echo -n $ENODE | sed -e "s/^\"\(.*\)@127.0.0.1:\([0-9]*\)?.*$/\1@10.1.2.11:\2/g"`
		echo "b$i: '$ENODE'"
		echo "a$i: '$ENODENEW'"
		for i in {2..4}; do
			docker run -d --ip 10.1.2.1$i --network swarminternal --rm -e PATH=/:/bin:/usr/bin:/usr/local/bin -e PASSWORD=tralala --name bzz$i -h bz$i swarm:trace --tracing --tracing.endpoint 10.1.2.2:6831 --sync-update-delay 1s --bzzport 80 --httpaddr 0.0.0.0 --tracing.svc "bz$i" --bootnodes $ENODENEW
		done
		;;
	"smoke")
		docker exec bzz1 /swarm-smoke --verbosity 5 --single --hosts 10.1.2.11,10.1.2.14 upload_and_sync
		;;
	"status")
		for i in {1..4}; do
			docker exec bzz$i /geth attach /root/.ethereum/bzzd.ipc --exec admin.peers
		done
		;;
	"stop")
		for i in {4..1}; do
			docker stop bzz$i
		done
		docker stop jaeger
		docker network rm swarminternal
		;;
esac
