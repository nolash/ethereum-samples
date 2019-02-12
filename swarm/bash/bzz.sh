#!/bin/bash

dir="${BASH_SOURCE%/*}"
if [ ! -d $dir ]
then
	dir=$PWD
else
	dir=`realpath $PWD`
fi

helpout=$(cat <<EOF
bzz - example help output header
\n
\nsome other description text
EOF
)

. $dir/bzz.inc


