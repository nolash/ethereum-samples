SRCDIR=./hello/
TMPNAME=./tmp
TMPDIR=${TMPNAME}/

all: test

test:
	mkdir -p ${TMPNAME}
	gcc -g -c -DTEST_HELLO=1 -o ${TMPDIR}hello.o ${SRCDIR}hello.c
	gcc -g -c -I${SRCDIR} -o ${TMPDIR}main_hello.o ./main_hello.c
	gcc -o hello.bin ${TMPDIR}hello.o ${TMPDIR}main_hello.o -lsqlite3
	rm -rf ${TMPNAME}

clean:
	rm -f *.bin
