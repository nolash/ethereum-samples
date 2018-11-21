#!/bin/bash

# preamble

# find include location
dir="${BASH_SOURCE%/*}"
if [ ! -d $dir ]
then
	dir=$PWD
else
	dir=`realpath $PWD`
fi

# hader
helpout=$(cat <<EOF
bzzp - quick swarm paste script
\n
\nuses http on localhost 8500 unless BZZAPI is set
EOF
)

. $dir/bzz.inc

# start main script
# initialize result url components

# if we don't have a file specified, we write a temp file from stdin
if [ -z $f ]
then
	f=`mktemp`
	[ $? -gt 0 ] && >&2 echo "unable to create temporary file" && exit 1
	cat - > $f
fi

# check if file has content
[ ! -s $f ] && >&2 echo "cannot post empty file $f" && exit 1

# get type of file
bzzmime=`file --mime-type -b $f`

# build the swarm command and result url parts
bzzcmd="swarm --bzzapi $url --mime $bzzmime"
prefix=$url
postfix=""

# omit manifest creation on raw
if [ $o_raw -gt 0 ]
then
	bzzcmd="$bzzcmd --manifest=false"
	prefix="${prefix}/bzz-raw:/"
	postfix="/?content_type=$bzzmime"
else
	prefix="${prefix}/bzz:/"
fi

$bzzcmd up $f | while read -r a
do
	if [[ $a =~ ^([a-zA-Z0-9]+)$ ]];
	then
		bzzhash=${BASH_REMATCH[1]}
		echo "$prefix$bzzhash/$postfix"
	fi
#		bzzhash=${BASH_REMATCH[1]}
#	if [[ $a =~ (\"hash\": \"([a-zA-Z0-9]+)\") ]];
#	then
#		bzzhash=${BASH_REMATCH[2]}
#		echo "$prefix$bzzhash$postfix/"
#	elif [[ $a =~ ^([a-zA-Z0-9]+)$ ]];
#	then
#		bzzhash=${BASH_REMATCH[1]}
#		echo "$prefix$bzzhash/$postfix"
#	fi
done
