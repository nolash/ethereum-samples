# swarm sqlite vfs 

This work is currently at "hello world" stage

The demo does the following:

* Create a new sqlite db with specified number of rows and size of blob data
* Create a LocalDPA and add db to it
* Prove that sqlite can call back the supplied hash used to open to the relevant go method
* Seek and read from DPA according to sqlite read call
* perform actual prepared (select) statement on database and output rows

The `demo/main_hello.c` file and its `Makefile` are only for separate testing of the `c` part and has no other relevance.

## USAGE 

There are command line args. Run with `-h` to see them

## ISSUES

1. On my system I sometimes have to provide `$GOPATH` explicitly to the linker to find the right package library:

```
go run -v -ldflags "-L $GOPATH/pkg/linux_amd64" main_hello.go
```
