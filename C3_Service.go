// Node stack API using HTTP and WS
package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

var (
	msgCount      = 5
	p2pPort       = 30100
	ipcpath       = ".demo.ipc"
	datadirPrefix = ".data_"
)

// the service we want to offer on the node
// it must implement the node.Service interface
type fooService struct {
	V int
}

// specify API structs that carry the methods we want to use
func (self *fooService) APIs() []rpc.API {
	return []rpc.API{
		rpc.API{
			Namespace: "foo",
			Version:   "0.42",
			Service:   &FooAPI{self.V},
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

func (api *FooAPI) GetNumber() (int, error) {
	return api.V, nil
}

type BarAPI struct {
}

func (api *BarAPI) Double(n int) (int, error) {
	return 2 * n, nil
}

func main() {

	// set up the service node with HTTP and WS
	// modules to be available through the different interfaces must be specified explicitly
	// Note that IPC exports ALL modules implicitly
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", p2pPort)
	cfg.IPCPath = ipcpath
	cfg.DataDir = fmt.Sprintf("%s%d", datadirPrefix, p2pPort)

	// HTTP parameters - both module "foo" and "bar"
	cfg.HTTPHost = node.DefaultHTTPHost
	cfg.HTTPPort = node.DefaultHTTPPort
	cfg.HTTPModules = append(cfg.HTTPModules, "foo", "bar")

	// Websocket parameters - only module "foo"
	cfg.WSHost = node.DefaultWSHost
	cfg.WSPort = node.DefaultWSPort
	cfg.WSModules = append(cfg.WSModules, "foo", "baz")

	// create the node instance with the config
	stack, err := node.New(cfg)
	if err != nil {
		demo.Log.Crit("ServiceNode create fail", "err", err)
	}
	defer os.RemoveAll(stack.DataDir())

	// wrapper function for servicenode to start the service
	foosvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return &fooService{42}, nil
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
	defer os.RemoveAll(cfg.DataDir)

	// the numbers we will pass to the api
	var number int
	var doublenumber int

	// connect to the RPC
	rpcclient_ipc, err := rpc.Dial(fmt.Sprintf("%s/%s", cfg.DataDir, cfg.IPCPath))

	// Using IPC, get the number from the FooApi
	err = rpcclient_ipc.Call(&number, "foo_getNumber")
	if err != nil {
		demo.Log.Crit("IPC RPC getnumber failed", "err", err)
	}
	demo.Log.Info("IPC", "getnumber", number)

	// Pass it to BarApi which doubles it
	err = rpcclient_ipc.Call(&doublenumber, "bar_double", number)
	if err != nil {
		demo.Log.Crit("IPC RPC double failed", "err", err)
	}
	demo.Log.Info("IPC", "double", doublenumber)

	// Same operation with HTTP
	// HTTP has both Apis
	number = 0
	doublenumber = 0

	rpcclient_http, err := rpc.Dial(fmt.Sprintf("http://%s:%d", node.DefaultHTTPHost, node.DefaultHTTPPort))

	err = rpcclient_http.Call(&number, "foo_getNumber")
	if err != nil {
		demo.Log.Crit("HTTP RPC getnumber failed", "err", err)
	}
	demo.Log.Info("HTTP", "getnumber", number)
	err = rpcclient_http.Call(&doublenumber, "bar_double", number)
	if err != nil {
		demo.Log.Crit("HTTP RPC double failed", "err", err)
	}
	demo.Log.Info("HTTP", "double", doublenumber)

	// Same operation with WS
	// we only added the first module to the WS interface, so the second call will fail
	number = 0
	doublenumber = 0

	rpcclient_ws, err := rpc.Dial(fmt.Sprintf("ws://%s:%d", node.DefaultWSHost, node.DefaultWSPort))

	err = rpcclient_ws.Call(&number, "foo_getNumber")
	if err != nil {
		demo.Log.Crit("WS RPC getnumber failed", "err", err)
	}
	demo.Log.Info("WS", "getnumber", number)
	err = rpcclient_ws.Call(&doublenumber, "bar_double", number)
	if err == nil {
		demo.Log.Crit("WS RPC double should have failed!")
	}
	demo.Log.Info("WS (double expected fail)", "err", err)

	// bring down the servicenode
	err = stack.Stop()
	if err != nil {
		demo.Log.Crit("Node stop fail", "err", err)
	}
}
