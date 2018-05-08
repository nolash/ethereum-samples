#!/bin/bash

function f_strip {
	for l in $(cat $1 | grep -ve ^pragma | grep -ve ^import); do
		echo $l
	done
}

IFS=$'\n'
echo "pragma solidity ^0.4.0;"
echo
for i in lib/*.*; do
	>&2 echo "adding $i"
	f_strip $i
done
src=( Store.sol AbstractMain.sol Upgrader.sol Main.sol MainUpgrade.sol )
for i in $(seq 0 $((${#src[@]}-1))); do
	f=${src[$i]}
	>&2 echo "adding $f"
	f_strip $f
done

