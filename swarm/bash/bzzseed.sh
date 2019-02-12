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

# help output head
helpout=$(cat <<EOF
bzzseed - script for upload of all dirs in subdir
\n
\nsuitable for use with crontab for seeding swarm content
\n
\nuses http on localhost 8500 unless BZZAPI is set
\n
EOF
)

. $dir/bzz.inc



dir=${@%/}

[ ! -d $DIR ] && exit 1;

bzzapi="http://localhost:8500"
[ ! -z $BZZAPI ] && bzzapi=$BZZAPI

echo using $bzzapi

in=0
out=0
for f in $dir/*; do
	((in++))
	swarm --bzzapi $bzzapi --recursive  up $f && ((out++))
done

[ $in -ne $out ] && logger "swarm seeder fail only $out/$in succeeded"
