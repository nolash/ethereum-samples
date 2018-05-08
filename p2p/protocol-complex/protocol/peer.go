package protocol

import (
	"errors"

	"github.com/ethereum/go-ethereum/p2p/protocols"
)

// Every protocol that wants to send messages back to sender, must have a peer abstraction
// this is because devp2p doesn't let us know about which peer is the sender
type DemoPeer struct {
	*protocols.Peer
	skillsHandler  func(*Skills, *protocols.Peer) error
	statusHandler  func(*Status, *protocols.Peer) error
	requestHandler func(*Request, *protocols.Peer) error
	resultHandler  func(*Result, *protocols.Peer) error
}

// Dispatcher for incoming messages
func (self *DemoPeer) Handle(msg interface{}) error {
	if typ, ok := msg.(*Skills); ok {
		return self.skillsHandler(typ, self.Peer)
	}
	if typ, ok := msg.(*Status); ok {
		return self.statusHandler(typ, self.Peer)
	}
	if typ, ok := msg.(*Request); ok {
		return self.requestHandler(typ, self.Peer)
	}
	if typ, ok := msg.(*Result); ok {
		return self.resultHandler(typ, self.Peer)
	}
	return errors.New("unknown message type")
}
