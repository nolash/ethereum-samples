package service

import (
	"fmt"
	"sync"

	"../protocol"
)

const (
	defaultSubmitsCapacity = 1000
)

type submitStore struct {
	serial uint64 // last request id sent from this node

	// handle submits
	entries  []*protocol.Request               // a wrapping array cache of requests used to retrieve the request data on a result response
	cursor   int                               // the current write position on the wrapping array cache
	idx      map[protocol.ID]*protocol.Request // index to look up the request cache though a request id
	capacity int                               // size of request cache (wrap threshold)

	mu sync.RWMutex
}

func newSubmitStore() *submitStore {
	return &submitStore{
		entries:  make([]*protocol.Request, defaultSubmitsCapacity),
		idx:      make(map[protocol.ID]*protocol.Request),
		capacity: defaultSubmitsCapacity,
	}
}

// add submits to entry cache
func (self *submitStore) Put(req *protocol.Request, id protocol.ID) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if _, ok := self.idx[id]; ok {
		return fmt.Errorf("entry already exists")
	}

	self.cursor++
	self.cursor %= self.capacity
	if self.entries[self.cursor] != nil {
		delete(self.idx, self.entries[self.cursor].Id)
	}
	self.entries[self.cursor] = req
	self.idx[id] = req
	return nil
}

func (self *submitStore) Have(id protocol.ID) bool {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.have(id)
}

func (self *submitStore) have(id protocol.ID) bool {
	_, ok := self.idx[id]
	return ok
}

func (self *submitStore) GetData(id protocol.ID) []byte {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.have(id) {
		return self.idx[id].Data
	}
	return nil
}

func (self *submitStore) GetDifficulty(id protocol.ID) uint8 {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.have(id) {
		return self.idx[id].Difficulty
	}
	return 0
}

func (self *submitStore) IncSerial() uint64 {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.serial++
	return self.serial
}

func (self *submitStore) LastSerial() uint64 {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.serial
}
