#!/bin/bash

dbuuid=$(`which uuidgen`)
SQLITE=`which sqlite3`
DBNAME="hello_${dbuuid}.sqlite3"

$SQLITE $DBNAME <<EOF
CREATE TABLE hello (
	k int not null,
	v blob
)
EOF

for i in {0..4096}; do
	dd if=/dev/urandom of=.data.sql bs=1024 count=6 &> /dev/null
$SQLITE $DBNAME <<EOF
	INSERT INTO hello (k, v) 
	VALUES ($i, readfile(".data.sql"))
EOF

	echo -ne "writing... $i\r"
done
echo 
echo done
