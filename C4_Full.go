// Node stack with ping/pong and API reporting
package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rpc"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
	"sync"
	"time"
)

var (
	stackW = &sync.WaitGroup{}
)

type FooPingMsg struct {
	Pong    bool
	Created time.Time
}

// the service we want to offer on the node
// it must implement the node.Service interface
type fooService struct {
	pongcount int
	pingC     map[discover.NodeID]chan struct{}
}

func newFooService() *fooService {
	return &fooService{
		pingC: make(map[discover.NodeID]chan struct{}),
	}
}

// specify API structs that carry the methods we want to use
func (self *fooService) APIs() []rpc.API {
	return []rpc.API{
		rpc.API{
			Namespace: "foo",
			Version:   "42",
			Service:   newFooAPI(self.pingC),
			Public:    true,
		},
	}
}

// the p2p.Protocol to run
// sends a ping to its peer, waits pong
func (self *fooService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		p2p.Protocol{
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
						msg, err := rw.ReadMsg()
						if err != nil {
							demo.Log.Warn("Receive p2p message fail", "err", err)
							break
						}

						// decode the message and check the contents
						var decodedmsg FooPingMsg
						err = msg.Decode(&decodedmsg)
						if err != nil {
							demo.Log.Error("Decode p2p message fail", "err", err)
							break
						}

						// if we get a pong, update our pong counter
						if decodedmsg.Pong {
							self.pongcount++
							demo.Log.Debug("received pong", "peer", p)
						} else {
							demo.Log.Debug("received ping", "peer", p)
							pingmsg := &FooPingMsg{
								Pong:    true,
								Created: time.Now(),
							}
							err := p2p.Send(rw, 0, pingmsg)
							if err != nil {
								demo.Log.Error("Send p2p message fail", "err", err)
								break
							}
							demo.Log.Debug("sent pong", "peer", p)
						}
					}
				}()

				// pings are invoked through the API using a channel
				// when this channel is closed we quit the protocol
				for {
					_, ok := <-self.pingC[p.ID()]
					if !ok {
						demo.Log.Debug("break protocol", "peer", p)
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
					demo.Log.Info("sending ping", "peer", p, "count", pingcount)
				}

				return nil
			},
		},
	}
}

func (self *fooService) Start(srv *p2p.Server) error {
	return nil
}

func (self *fooService) Stop() error {
	return nil
}

// Specify the API
// in this example we don't care about who the pongs comes from, we count them all
// note it is a bit fragile; we don't check for closed channels
type FooAPI struct {
	pongcount int
	pingC     map[discover.NodeID]chan struct{}
}

func newFooAPI(pingC map[discover.NodeID]chan struct{}) *FooAPI {
	return &FooAPI{
		pingC: pingC,
	}
}

// invoke a single ping
func (api *FooAPI) Ping(id discover.NodeID) error {
	api.pingC[id] <- struct{}{}
	return nil
}

// quit the ping protocol
func (api *FooAPI) Quit(id discover.NodeID) error {
	demo.Log.Debug("quitting API", "peer", id)
	if api.pingC[id] == nil {
		return fmt.Errorf("unknown peer")
	}
	close(api.pingC[id])
	return nil
}

// return the amounts of pongs received
func (api *FooAPI) PongCount() (int, error) {
	return api.pongcount, nil
}

func main() {

	// create the two nodes
	stack_one, err := demo.NewServiceNode(demo.P2PDefaultPort, 0, 0)
	if err != nil {
		demo.Log.Crit("Create servicenode #1 fail", "err", err)
	}
	stack_two, err := demo.NewServiceNode(demo.P2PDefaultPort+1, 0, 0)
	if err != nil {
		demo.Log.Crit("Create servicenode #2 fail", "err", err)
	}

	foosvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return newFooService(), nil
	}

	// register adds the service to the services the servicenode starts when started
	err = stack_one.Register(foosvc)
	if err != nil {
		demo.Log.Crit("Register service in servicenode #1 fail", "err", err)
	}
	err = stack_two.Register(foosvc)
	if err != nil {
		demo.Log.Crit("Register service in servicenode #2 fail", "err", err)
	}

	// start the nodes
	err = stack_one.Start()
	if err != nil {
		demo.Log.Crit("servicenode #1 start failed", "err", err)
	}
	err = stack_two.Start()
	if err != nil {
		demo.Log.Crit("servicenode #2 start failed", "err", err)
	}

	// connect to the servicenode RPCs
	// subscribe to events
	rpcclient_one, err := rpc.Dial(fmt.Sprintf("%s/%s", demo.Datadir(demo.P2PDefaultPort), demo.IPCName))
	if err != nil {
		demo.Log.Crit("connect to servicenode #1 IPC fail", "err", err)
	}
	rpcclient_two, err := rpc.Dial(fmt.Sprintf("%s/%s", demo.Datadir(demo.P2PDefaultPort+1), demo.IPCName))
	if err != nil {
		demo.Log.Crit("connect to servicenode #2 IPC fail", "err", err)
	}

	// verify that the initial pong counts are 0
	var count int
	err = rpcclient_one.Call(&count, "foo_pongCount")
	if err != nil {
		demo.Log.Crit("servicenode #1 pongcount RPC failed", "err", err)
	}
	demo.Log.Info("servicenode #1 before ping", "pongcount", count)

	err = rpcclient_two.Call(&count, "foo_pongCount")
	if err != nil {
		demo.Log.Crit("servicenode #2 pongcount RPC failed", "err", err)
	}
	demo.Log.Info("servicenode #2 before ping", "pongcount", count)

	// connect the nodes
	// subscribe to events
	eventOneC := make(chan *p2p.PeerEvent)
	eventTwoC := make(chan *p2p.PeerEvent)

	srv_one := stack_one.Server()
	srv_two := stack_two.Server()

	sub_one := srv_one.SubscribeEvents(eventOneC)
	sub_two := srv_two.SubscribeEvents(eventTwoC)

	p2pnode_two := srv_two.Self()
	srv_one.AddPeer(p2pnode_two)

	// fork and do the pinging
	stackW.Add(2)
	pingmax_one := 4
	pingmax_two := 2
	go func() {
		ev := <-eventOneC
		if ev.Type != "add" {
			demo.Log.Error("server #1 expected peer add", "eventtype", ev.Type)
			stackW.Done()
			return
		}
		demo.Log.Debug("server #1 connected", "peer", ev.Peer)
		for i := 0; i < pingmax_one; i++ {
			err := rpcclient_one.Call(nil, "foo_ping", ev.Peer)
			if err != nil {
				demo.Log.Error("server #1 RPC ping fail", "err", err)
				stackW.Done()
				break
			}
		}
		err := rpcclient_one.Call(nil, "foo_quit", ev.Peer)
		if err != nil {
			demo.Log.Crit("server #1 RPC quit fail", "err", err)
		}
		ev = <-eventOneC
		if ev.Type != "drop" {
			demo.Log.Error("server #1 expected peer drop", "eventtype", ev.Type)
		}
		stackW.Done()
	}()

	go func() {
		ev := <-eventTwoC
		if ev.Type != "add" {
			demo.Log.Error("expected peer add", "eventtype", ev.Type)
			stackW.Done()
			return
		}
		demo.Log.Debug("server #2 connected", "peer", ev.Peer)
		for i := 0; i < pingmax_two; i++ {
			err := rpcclient_two.Call(nil, "foo_ping", ev.Peer)
			if err != nil {
				demo.Log.Error("server #2 RPC ping fail", "err", err)
				stackW.Done()
				break
			}
		}
		err := rpcclient_two.Call(nil, "foo_quit", ev.Peer)
		if err != nil {
			demo.Log.Crit("server #2 RPC quit fail", "err", err)
		}
		ev = <-eventTwoC
		if ev.Type != "drop" {
			demo.Log.Error("expected peer drop", "eventtype", ev.Type)
		}
		stackW.Done()
	}()

	// wait for every to finish
	// add a grace period for pongs to arrive
	stackW.Wait()
	demo.Log.Debug("Waiting a bit for protocols to finish")
	time.Sleep(time.Millisecond * 250)

	// inspect the result
	err = rpcclient_one.Call(&count, "foo_pongCount")
	if err != nil {
		demo.Log.Crit("servicenode #1 pongcount RPC failed", "err", err)
	}
	demo.Log.Info("servicenode #1 after ping", "pongcount", count)

	err = rpcclient_two.Call(&count, "foo_pongCount")
	if err != nil {
		demo.Log.Crit("servicenode #2 pongcount RPC failed", "err", err)
	}
	demo.Log.Info("servicenode #2 after ping", "pongcount", count)

	// bring down the servicenode
	sub_one.Unsubscribe()
	sub_two.Unsubscribe()
	stack_one.Stop()
	stack_two.Stop()
}
