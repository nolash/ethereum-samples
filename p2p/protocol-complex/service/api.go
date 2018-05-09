package service

import (
	"../protocol"
)

type DemoAPI struct {
	service *Demo
}

func newDemoAPI(s *Demo) *DemoAPI {
	return &DemoAPI{
		service: s,
	}
}

func (self *DemoAPI) Submit(data []byte, difficulty uint8) (protocol.ID, error) {
	return self.service.submitRequest(data, difficulty)
}

func (self *DemoAPI) Stop() error {
	//self.service.running = false
	return nil
}

func (self *DemoAPI) SetDifficulty(d uint8) error {
	self.service.mu.Lock()
	defer self.service.mu.Unlock()
	self.service.maxDifficulty = d
	return nil
}
