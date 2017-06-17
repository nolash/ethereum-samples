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

// this is the service that contains the protocol we want to run
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

// this is the handler of the incoming message
func fooHandler(msg interface{}) error {
	foomsg, ok := msg.(*FooMsg)
	if ok {
		Log.Debug("foomsg!", "msg", foomsg)
		return nil
	}

	fooothermsg, ok := msg.(*FooOtherMsg)
	if ok {
		Log.Debug("fooothermsg!", "msg", fooothermsg)
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

type FooUnhandledMsg struct {
	Content []byte
}

// we really only care about the protocol part
// everything else is minimal
func (self fooService) APIs() []rpc.API {
	return []rpc.API{}
}

func (self fooService) Start(srv *p2p.Server) error {
	return nil
}

func (self fooService) Stop() error {
	return nil
}
