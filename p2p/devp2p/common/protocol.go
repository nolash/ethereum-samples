package common

import (
	"fmt"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rpc"
	"time"
)

const (
	FooProtocolName       = "fooping"
	FooProtocolVersion    = 42
	FooProtocolMaxMsgSize = 1024
)

type FooPingMsg struct {
	Pong    bool
	Created time.Time
}

var (
	FooMessages = []interface{}{
		&FooPingMsg{},
	}
	FooProtocol = protocols.Spec{
		Name:       FooProtocolName,
		Version:    FooProtocolVersion,
		MaxMsgSize: FooProtocolMaxMsgSize,
		Messages:   FooMessages,
	}
)

// the service we want to offer on the node
// it must implement the node.Service interface
type FooService struct {
	pongcount int
	pingC     map[enode.ID]chan struct{}
}

func NewFooService() *FooService {
	return &FooService{
		pingC: make(map[enode.ID]chan struct{}),
	}
}

// specify API structs that carry the methods we want to use
func (self *FooService) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "foo",
			Version:   "42",
			Service:   NewFooAPI(self.pingC, &self.pongcount),
			Public:    true,
		},
	}
}

// the p2p.Protocol to run
// sends a ping to its peer, waits pong
func (self *FooService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "fooping",
			Version: 666,
			Length:  1,
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {

				// create the channel when a connection is made
				self.pingC[p.ID()] = make(chan struct{})
				pingcount := 0

				// create the message structure

				// we don't know if we're awaiting anything at the time of the kill so this subroutine will run till the application ends
				go func() {
					for {
						Log.Debug("in pong catch")
						msg, err := rw.ReadMsg()
						if err != nil {
							Log.Warn("Receive p2p message fail", "err", err)
							break
						}

						Log.Debug("in pong catch after readmsg")
						// decode the message and check the contents
						var decodedmsg FooPingMsg
						err = msg.Decode(&decodedmsg)
						if err != nil {
							Log.Error("Decode p2p message fail", "err", err)
							break
						}

						// if we get a pong, update our pong counter
						if decodedmsg.Pong {
							self.pongcount++
							Log.Debug("received pong", "peer", p, "count", self.pongcount)
						} else {
							Log.Debug("received ping", "peer", p)
							pingmsg := &FooPingMsg{
								Pong:    true,
								Created: time.Now(),
							}
							err := p2p.Send(rw, 0, pingmsg)
							if err != nil {
								Log.Error("Send p2p message fail", "err", err)
								break
							}
							Log.Debug("sent pong", "peer", p)
						}
					}
				}()

				// pings are invoked through the API using a channel
				// when this channel is closed we quit the protocol
				for {
					_, ok := <-self.pingC[p.ID()]
					if !ok {
						Log.Debug("break protocol", "peer", p)
						break
					}
					pingmsg := &FooPingMsg{
						Pong:    false,
						Created: time.Now(),
					}
					err := p2p.Send(rw, 0, pingmsg)
					if err != nil {
						return fmt.Errorf("Send p2p message fail: %v", err)
					}
					pingcount++
					Log.Info("sending ping", "peer", p, "count", pingcount)
				}

				return nil
			},
		},
	}
}

func (self *FooService) Start(srv *p2p.Server) error {
	return nil
}

func (self *FooService) Stop() error {
	return nil
}

// Specify the API
// in this example we don't care about who the pongs comes from, we count them all
// note it is a bit fragile; we don't check for closed channels
type FooAPI struct {
	running   bool
	pongcount *int
	pingC     map[enode.ID]chan struct{}
}

func NewFooAPI(pingC map[enode.ID]chan struct{}, pongcount *int) *FooAPI {
	return &FooAPI{
		running:   true,
		pingC:     pingC,
		pongcount: pongcount,
	}
}

func (api *FooAPI) Increment() {
	*api.pongcount++
}

// invoke a single ping
func (api *FooAPI) Ping(id enode.ID) error {
	if api.running {
		api.pingC[id] <- struct{}{}
	}
	return nil
}

// quit the ping protocol
func (api *FooAPI) Quit(id enode.ID) error {
	Log.Debug("quitting API", "peer", id)
	if api.pingC[id] == nil {
		return fmt.Errorf("unknown peer")
	}
	api.running = false
	close(api.pingC[id])
	return nil
}

// return the amounts of pongs received
func (api *FooAPI) PongCount() (int, error) {
	return *api.pongcount, nil
}
