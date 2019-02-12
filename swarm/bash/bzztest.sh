#!/usr/bin/env /bin/bash

if [ "$BZZ" == "" ]; then
	echo "please set the BZZ environment variable to the swarm executable path"
	exit 1
fi

if [ ! -f $BZZ ]; then
	echo "swarm executable '$BZZ' not found"
	exit 1
fi

if [ "$BZZPATH" == "" ]; then
	BZZPATH=.
fi

BZZDIR=.ethereum
BZZHOST=127.0.0.1
BZZPORT=30399 
BZZPSSPORT=8546

function helpout {
	echo "usage: [-h ip] [-p bzzport] [-pp websocketport] [datadir]"
	echo ""
	echo "Uses the swarm executable specified by the BZZ environment variable."
	echo "Datadir is looked up relative to BZZPATH."
	echo "If BZZPATH is not set, current directory is used."
	exit 1
}

while test $# -gt 0; do
	case "$1" in
		-h)
			shift
			if test $# -gt 0; then
				BZZHOST=$1
				shift
			else
				helpout
			fi
			;;
		-pp)
			shift
			if test $# -gt 0; then
				BZZPSSPORT=$1
				shift
			else
				helpout
			fi
			;;
		-p)
			shift
			if test $# -gt 0; then
				BZZPORT=$1
				shift
			else
				helpout
			fi
			;;

		*)
			break
			;;
	esac
done

if [ ! "$1" == "" ]; then
	BZZDIR=$1
	echo "i have 1, is '$1'"
fi

if [ ! -d ${BZZPATH}/$BZZDIR ]; then
	echo "bzz datadir $BZZPATH/$BZZDIR does not exist, exiting ..."
	exit 1
fi

cd $BZZPATH

BZZACCOUNT=`find $BZZDIR/keystore -name UTC--* -print -quit | sed s/.*\.[0-9]*Z--//`

if [ "$BZZACCOUNT" == "" ]; then
	echo "no valid keystore entries in $BZZPATH/$BZZDIR/keystore, exiting ..."
	exit 1
fi

echo "using datadir $BZZPATH/$BZZDIR"
echo "using bzzaccount $BZZACCOUNT"
echo "using host $BZZHOST/$BZZPSSPORT"

echo tralala | $BZZ --nat extip:$BZZHOST --datadir $BZZPATH/$BZZDIR --bzzaccount $BZZACCOUNT --port $BZZPORT --wsport $BZZPSSPORT --ws --pss --ethapi '' --verbosity 6 2> $BZZPATH/$BZZDIR.log

