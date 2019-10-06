#!/bin/bash

#IMG="swarm:pss"
IMG="ethdevops/swarm:latest"
WSFLAGS="--ws --wsapi=pss,bzz,admin --wsaddr=0.0.0.0 --wsorigins=*"
BZZFLAGS="--httpaddr=0.0.0.0"

case $1 in
	"start")
		#docker network create --internal --subnet 10.1.3.0/24 pssnet
		docker network create --subnet 10.1.3.0/24 pssnet
		#docker run -p 8500:8500 -p 8546:8546 -d --ip 10.1.3.11 --network pssnet --rm -e PATH=/:/bin:/usr/bin:/usr/local/bin -e PASSWORD=tralala --name pss1 -h ps1 $IMG $BZZFLAGS $WSFLAGS
		docker run -d --ip 10.1.3.11 --network pssnet --rm -e PATH=/:/bin:/usr/bin:/usr/local/bin -e PASSWORD=tralala --name pss1 -h ps1 $IMG $BZZFLAGS $WSFLAGS
		# wait for nodes to start
		sleep 5
		ENODE=`docker exec pss1 /geth attach /root/.ethereum/bzzd.ipc --exec admin.nodeInfo.enode`
		ENODENEW=`echo -n $ENODE | sed -e "s/^\"\(.*\)@127.0.0.1:\([0-9]*\).*$/\1@10.1.3.11:\2/g"`
		echo "b$i: '$ENODE'"
		echo "a$i: '$ENODENEW'"
		for i in {2..4}; do
			#docker run -d --ip 10.1.3.1$i -p 1850$i:8500 --network pssnet --rm -e PATH=/:/bin:/usr/bin:/usr/local/bin -e PASSWORD=tralala --name pss$i -h ps$i $IMG --nat extip:10.1.3.1$i --bootnodes=$ENODENEW $BZZFLAGS $WSFLAGS
			docker run -d --ip 10.1.3.1$i --network pssnet --rm -e PATH=/:/bin:/usr/bin:/usr/local/bin -e PASSWORD=tralala --name pss$i -h ps$i $IMG --nat extip:10.1.3.1$i --bootnodes=$ENODENEW $BZZFLAGS $WSFLAGS
		done
		;;
	"status")
		for i in {1..4}; do
			docker exec pss$i /geth attach /root/.ethereum/bzzd.ipc --exec admin.peers
		done
		;;
	"stop")
		for i in {4..1}; do
			docker stop pss$i
		done
		docker network rm pssnet
		;;
esac
