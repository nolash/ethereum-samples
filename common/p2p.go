package common

import (
	"fmt"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"reflect"
	"time"
)

var (
	fooprotocolmessages = []interface{}{
		&FooMsg{},
		&FooOtherMsg{},
		&FooUnhandledMsg{},
	}
)

// implements github.com/ethereum/go-ethereum/node.Service interface
//
// this is the service that contains the protocol we want to run
// is it passed to the ServiceNode, and is started when the node is started
type fooService struct {
	run func(*p2p.Peer, p2p.MsgReadWriter) error
}

// convenience constructor for the service
// the run-function passed here will be invoked on the peer when a p2p tcp connection is made
func NewFooService(run func(p *p2p.Peer, rw p2p.MsgReadWriter) error) (fooService, error) {
	return fooService{
		run: run,
	}, nil
}

// lowlevel p2p protocol implementation
// used for connections that solely use the /p2p package
func (self fooService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		p2p.Protocol{
			Name:    FooProtocolName,
			Version: FooProtocolVersion,
			Length:  uint64(len(fooprotocolmessages)),
			Run:     self.run,
		},
	}
}

// this function is handles all incoming messages for the foo-service protocol
func FooHandler(msg interface{}) error {
	foomsg, ok := msg.(*FooMsg)
	if ok {
		Log.Debug("Received foomsg!", "msg", foomsg)
		return nil
	}

	fooothermsg, ok := msg.(*FooOtherMsg)
	if ok {
		Log.Debug("Received fooothermsg!", "msg", fooothermsg)
		return nil
	}
	return fmt.Errorf("If you reach here, you forgot to make a handler for the message type %v", reflect.TypeOf(msg))
}

// we have two types of messages we want to send
type FooMsg struct {
	Serial uint
}

type FooOtherMsg struct {
	Created time.Time
}

// ... and a third that is registered (so it can be sent) but has no incoming handler
type FooUnhandledMsg struct {
	Content []byte
}

// everything below here is also needed to satisfy the node.Service interface
func (self fooService) APIs() []rpc.API {
	return []rpc.API{}
}

func (self fooService) Start(srv *p2p.Server) error {
	return nil
}

func (self fooService) Stop() error {
	return nil
}
