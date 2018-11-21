#!/bin/bash

# defaults
remoteurl="http://swarm-gateways.net"
url="http://127.0.0.1:8500"

# flag store vars
o_raw=0
o_remote=0

while test $# -gt 0
do
	case $1 in
		-h|--help)
			echo -e $helpout
			echo
			echo "-h,--help	display this help"
			echo "-x	raw output"
			echo "-r	use fallback gateway"
			exit 0	
			;;
		-x)
			shift
			o_raw=1	
			;;
		-r)
			shift
			o_remote=1
			;;
		-*)
			shift
			;;
		*)	
			break
			;;
	esac
done

f=$1

# remote flag sets gateways as api
if [ $o_remote -gt 0 ]
then
	url=$remoteurl
elif [ ! -z $BZZAPI ]
then
	url=$BZZAPI
fi

