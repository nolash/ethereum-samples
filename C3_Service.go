// Node stack API using HTTP and WS
package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

const (
	msgCount = 5
)

// the service we want to offer on the node
// it must implement the node.Service interface
type fooService struct {
	V int
}

func newFooService(v int) *fooService {
	return &fooService{
		V: v,
	}
}

// specify API structs that carry the methods we want to use
func (self *fooService) APIs() []rpc.API {
	return []rpc.API{
		rpc.API{
			Namespace: "foo",
			Version:   "0.42",
			Service:   newFooAPI(self.V),
			Public:    true,
		},
		rpc.API{
			Namespace: "bar",
			Version:   "0.666",
			Service:   &BarAPI{},
			Public:    true,
		},
	}
}

// these are needed to satisfy the node.Service interface
// in this example they do nothing
func (self *fooService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

func (self *fooService) Start(srv *p2p.Server) error {
	return nil
}

func (self *fooService) Stop() error {
	return nil
}

// remember that API structs to be offered MUST be exported
type FooAPI struct {
	V int
}

func newFooAPI(v int) *FooAPI {
	return &FooAPI{
		V: v,
	}
}

func (api *FooAPI) GetNumber() (int, error) {
	return api.V, nil
}

type BarAPI struct {
}

func (api *BarAPI) Double(n int) (int, error) {
	return 2 * n, nil
}

func main() {

	// start servicenode with HTTP and WS
	// modules to be available through the different interfaces must be specified explicitly
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", demo.P2PDefaultPort)
	cfg.IPCPath = demo.IPCName
	cfg.DataDir = demo.Datadir(demo.P2PDefaultPort)
	cfg.HTTPHost = node.DefaultHTTPHost
	cfg.HTTPPort = node.DefaultHTTPPort
	cfg.HTTPModules = append(cfg.HTTPModules, "foo", "bar")
	cfg.WSHost = node.DefaultWSHost
	cfg.WSPort = node.DefaultWSPort
	cfg.WSModules = append(cfg.WSModules, "foo")
	stack, err := node.New(cfg)
	if err != nil {
		demo.Log.Crit("ServiceNode create fail", "err", err)
	}

	// wrapper function for servicenode to start the service
	foosvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return newFooService(42), nil
	}

	// register adds the service to the services the servicenode starts when started
	err = stack.Register(foosvc)
	if err != nil {
		demo.Log.Crit("Register service in ServiceNode failed", "err", err)
	}

	// start the node
	// after this all features served by the node are available
	// thus we can call the API
	err = stack.Start()
	if err != nil {
		demo.Log.Crit("ServiceNode start failed", "err", err)
	}

	// IPC exports all modules implicitly
	var startnumber int
	var resultnumber int

	rpcclient_ipc, err := rpc.Dial(fmt.Sprintf("%s/%s", demo.Datadir(demo.P2PDefaultPort), demo.IPCName))

	err = rpcclient_ipc.Call(&startnumber, "foo_getNumber")
	if err != nil {
		demo.Log.Crit("IPC RPC getnumber failed", "err", err)
	}
	demo.Log.Info("IPC", "getnumber", startnumber)
	err = rpcclient_ipc.Call(&resultnumber, "bar_double", startnumber)
	if err != nil {
		demo.Log.Crit("IPC RPC double failed", "err", err)
	}
	demo.Log.Info("IPC", "double", resultnumber)

	// we added both modules to the HTTP interface
	startnumber = 0
	resultnumber = 0

	rpcclient_http, err := rpc.Dial(fmt.Sprintf("http://%s:%d", node.DefaultHTTPHost, node.DefaultHTTPPort))

	err = rpcclient_http.Call(&startnumber, "foo_getNumber")
	if err != nil {
		demo.Log.Crit("HTTP RPC getnumber failed", "err", err)
	}
	demo.Log.Info("HTTP", "getnumber", startnumber)
	err = rpcclient_http.Call(&resultnumber, "bar_double", startnumber)
	if err != nil {
		demo.Log.Crit("HTTP RPC double failed", "err", err)
	}
	demo.Log.Info("HTTP", "double", resultnumber)

	// we only added the first module to the WS interface, so the second call will fail
	startnumber = 0
	resultnumber = 0

	rpcclient_ws, err := rpc.Dial(fmt.Sprintf("ws://%s:%d", node.DefaultWSHost, node.DefaultWSPort))

	err = rpcclient_ws.Call(&startnumber, "foo_getNumber")
	if err != nil {
		demo.Log.Crit("WS RPC getnumber failed", "err", err)
	}
	demo.Log.Info("WS", "getnumber", startnumber)
	err = rpcclient_ws.Call(&resultnumber, "bar_double", startnumber)
	if err == nil {
		demo.Log.Crit("WS RPC double should have failed!")
	}
	demo.Log.Info("WS (expected fail)", "err", err)

	// bring down the servicenode
	stack.Stop()
}
