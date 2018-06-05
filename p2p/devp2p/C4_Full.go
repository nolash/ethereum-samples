// Node stack with ping/pong and API reporting
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rpc"

	demo "./common"
)

var (
	p2pPort       = 30100
	ipcpath       = ".demo.ipc"
	datadirPrefix = ".data_"
	stackW        = &sync.WaitGroup{}
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

// specify API structs that carry the methods we want to use
func (self *fooService) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "foo",
			Version:   "42",
			Service: &FooAPI{
				running:   true,
				pongcount: &self.pongcount,
				pingC:     self.pingC,
			},
			Public: true,
		},
	}
}

// the p2p.Protocol to run
// sends a ping to its peer, waits pong
func (self *fooService) Protocols() []p2p.Protocol {
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
						// listen for new message
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
						// if not, send pong
						if decodedmsg.Pong {
							self.pongcount++
							demo.Log.Debug("received pong", "peer", p, "count", self.pongcount)
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
					// wait for signal to send ping
					_, ok := <-self.pingC[p.ID()]
					if !ok {
						demo.Log.Debug("break protocol", "peer", p)
						break
					}

					// send ping
					pingmsg := &FooPingMsg{
						Pong:    false,
						Created: time.Now(),
					}

					// either handler or sender should be asynchronous, otherwise we might deadlock
					go p2p.Send(rw, 0, pingmsg)
					pingcount++
					demo.Log.Info("sent ping", "peer", p, "count", pingcount)
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
	running   bool
	pongcount *int
	pingC     map[discover.NodeID]chan struct{}
}

func (api *FooAPI) Increment() {
	*api.pongcount++
}

// invoke a single ping
func (api *FooAPI) Ping(id discover.NodeID) error {
	if api.running {
		api.pingC[id] <- struct{}{}
	}
	return nil
}

// quit the ping protocol
func (api *FooAPI) Quit(id discover.NodeID) error {
	demo.Log.Debug("quitting API", "peer", id)
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

// set up the local service node
func newServiceNode(port int, httpport int, wsport int, modules ...string) (*node.Node, error) {
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", port)
	cfg.P2P.EnableMsgEvents = true
	cfg.P2P.NoDiscovery = true
	cfg.IPCPath = ipcpath
	cfg.DataDir = fmt.Sprintf("%s%d", datadirPrefix, port)
	if httpport > 0 {
		cfg.HTTPHost = node.DefaultHTTPHost
		cfg.HTTPPort = httpport
	}
	if wsport > 0 {
		cfg.WSHost = node.DefaultWSHost
		cfg.WSPort = wsport
		cfg.WSOrigins = []string{"*"}
		for i := 0; i < len(modules); i++ {
			cfg.WSModules = append(cfg.WSModules, modules[i])
		}
	}
	stack, err := node.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("ServiceNode create fail: %v", err)
	}
	return stack, nil
}

func main() {

	// create the two nodes
	stack_one, err := newServiceNode(p2pPort, 0, 0)
	if err != nil {
		demo.Log.Crit("Create servicenode #1 fail", "err", err)
	}
	stack_two, err := newServiceNode(p2pPort+1, 0, 0)
	if err != nil {
		demo.Log.Crit("Create servicenode #2 fail", "err", err)
	}

	// wrapper function for servicenode to start the service
	foosvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return &fooService{
			pingC: make(map[discover.NodeID]chan struct{}),
		}, nil
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
	rpcclient_one, err := rpc.Dial(filepath.Join(stack_one.DataDir(), ipcpath))
	if err != nil {
		demo.Log.Crit("connect to servicenode #1 IPC fail", "err", err)
	}
	defer os.RemoveAll(stack_one.DataDir())

	rpcclient_two, err := rpc.Dial(filepath.Join(stack_two.DataDir(), ipcpath))
	if err != nil {
		demo.Log.Crit("connect to servicenode #2 IPC fail", "err", err)
	}
	defer os.RemoveAll(stack_two.DataDir())

	// display that the initial pong counts are 0
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

	// get the server instances
	srv_one := stack_one.Server()
	srv_two := stack_two.Server()

	// subscribe to peerevents
	eventOneC := make(chan *p2p.PeerEvent)
	sub_one := srv_one.SubscribeEvents(eventOneC)

	eventTwoC := make(chan *p2p.PeerEvent)
	sub_two := srv_two.SubscribeEvents(eventTwoC)

	// connect the nodes
	p2pnode_two := srv_two.Self()
	srv_one.AddPeer(p2pnode_two)

	// fork and do the pinging
	stackW.Add(2)
	pingmax_one := 4
	pingmax_two := 2

	go func() {

		// when we get the add event, we know we are connected
		ev := <-eventOneC
		if ev.Type != "add" {
			demo.Log.Error("server #1 expected peer add", "eventtype", ev.Type)
			stackW.Done()
			return
		}
		demo.Log.Debug("server #1 connected", "peer", ev.Peer)

		// send the pings
		for i := 0; i < pingmax_one; i++ {
			err := rpcclient_one.Call(nil, "foo_ping", ev.Peer)
			if err != nil {
				demo.Log.Error("server #1 RPC ping fail", "err", err)
				stackW.Done()
				break
			}
		}

		// wait for all msgrecv events
		// pings we receive, and pongs we expect from pings we sent
		for i := 0; i < pingmax_two+pingmax_one; {
			ev := <-eventOneC
			demo.Log.Warn("msg", "type", ev.Type, "i", i)
			if ev.Type == "msgrecv" {
				i++
			}
		}

		stackW.Done()
	}()

	// mirrors the previous go func
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

		for i := 0; i < pingmax_one+pingmax_two; {
			ev := <-eventTwoC
			if ev.Type == "msgrecv" {
				demo.Log.Warn("msg", "type", ev.Type, "i", i)
				i++
			}
		}

		stackW.Done()
	}()

	// wait for the two ping pong exchanges to finish
	stackW.Wait()

	// tell the API to shut down
	// this will disconnect the peers and close the channels connecting API and protocol
	err = rpcclient_one.Call(nil, "foo_quit", srv_two.Self().ID)
	if err != nil {
		demo.Log.Error("server #1 RPC quit fail", "err", err)
	}
	err = rpcclient_two.Call(nil, "foo_quit", srv_one.Self().ID)
	if err != nil {
		demo.Log.Error("server #2 RPC quit fail", "err", err)
	}

	// disconnect will generate drop events
	for {
		ev := <-eventOneC
		if ev.Type == "drop" {
			break
		}
	}
	for {
		ev := <-eventTwoC
		if ev.Type == "drop" {
			break
		}
	}

	// proudly inspect the results
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

	// bring down the servicenodes
	sub_one.Unsubscribe()
	sub_two.Unsubscribe()
	stack_one.Stop()
	stack_two.Stop()
}
