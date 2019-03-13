#!/bin/bash

for f in *.go; do
	go run $f
	res=$?
	if [ $res -gt 0 ]; then
		echo $f returned $res
		exit 1	
	fi	
done
