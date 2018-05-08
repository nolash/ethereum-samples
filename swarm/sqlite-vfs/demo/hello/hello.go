package hello

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

/*
#cgo LDFLAGS: -lsqlite3
#include <stdlib.h>
#include <sqlite3.h>
#include "hello.h"
*/
import "C"

var (
	vfs        *C.struct_sqlite3_vfs
	chunkFiles []*chunkFile
	dpa        *storage.DPA
	debug      bool
)

type chunkFile struct {
	reader io.ReadSeeker
	size   int64
}

//export GoBzzOpen
func GoBzzOpen(name *C.char, fd *C.int) C.int {
	hex := C.GoStringN(name, 64) // must be specific, sqlite often mangles the 0 terminator
	hash := common.HexToHash(hex)
	key := storage.Key(hash[:])
	log.Debug("retrieve", "key", key, "name", name, "hex", hex)
	r := dpa.Retrieve(key)
	sz, err := r.Size(nil)
	if err != nil {
		log.Error("dpa size query fail", "err", err)
		return 1
	}
	chunkfile := &chunkFile{
		reader: r,
		size:   sz,
	}
	chunkFiles = append(chunkFiles, chunkfile)
	*fd = C.int(len(chunkFiles) - 1)
	return 0
}

//export GoBzzFileSize
func GoBzzFileSize(c_fd C.int) C.longlong {
	c := int(c_fd)
	if !isValidFD(c) {
		return -1
	}
	log.Trace(fmt.Sprintf("reporting filesize: %d", chunkFiles[c].size))
	return C.longlong(chunkFiles[c].size)
}

//export GoBzzRead
func GoBzzRead(c_fd C.int, p unsafe.Pointer, amount C.int, offset C.longlong) int64 {

	// check if we have this reader
	c := int(c_fd)
	if !isValidFD(c) {
		return -1
	}

	// seek and retrieve from dpa
	file := chunkFiles[c]
	file.reader.Seek(int64(offset), 0)
	data := make([]byte, amount)
	c, err := file.reader.Read(data)
	if err != nil && err != io.EOF {
		log.Warn("read error", "err", err, "read", c)
		return -1
	}

	// not sure about this pointer handling, looks risky
	var pp []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&pp))
	hdr.Len = int(amount)
	hdr.Data = uintptr(p)
	copy(pp, data[:amount])

	log.Trace(fmt.Sprintf("returning data: '%x'...'%x'", data[:16], data[amount-16:amount]))
	return int64(len(data))
}

// open database with using dpa
func Open(key storage.Key) error {
	keystr := key.String()
	bzzhash := C.CString(keystr)
	defer C.free(unsafe.Pointer(bzzhash))
	r := C.bzzvfs_open(bzzhash)
	if r != C.SQLITE_OK {
		return fmt.Errorf("sqlite open fail: %d", r)
	}
	return nil
}

// execute query on database using dpa
func Exec(sql string) error {
	csql := C.CString(sql)
	defer C.free(unsafe.Pointer(csql))
	res := make([]byte, 1024)
	cres := C.CString(string(res))
	defer C.free(unsafe.Pointer(cres))
	log.Trace(fmt.Sprintf("executing %s... ", sql))
	r := C.bzzvfs_exec(C.int(len(sql)), csql, 1024, cres)
	if r != C.SQLITE_OK {
		return fmt.Errorf("sqlite exec fail (%d): %s", r, C.GoString(cres))
	}
	return nil
}

// register bzz vfs
func Init(newdpa *storage.DPA) error {
	r := C.bzzvfs_register()
	if r != C.SQLITE_OK {
		fmt.Errorf("%d", r)
	}
	dpa = newdpa
	if debug {
		C.bzzvfs_debug(1)
	} else {
		C.bzzvfs_debug(0)
	}
	return nil
}

func Close() {
	C.bzzvfs_close()
}

func Debug(active bool) {
	debug = active
}

func isValidFD(fd int) bool {
	if fd > len(chunkFiles) {
		return false
	}
	return true
}
