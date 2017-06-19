// querying the p2p Server through RPC
package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

// set up an object that can contain the API methods
type FooAPI struct {
	server *p2p.Server
}

// assign the server object as a member for use in the API call
func NewFooAPI(srv *p2p.Server) *FooAPI {
	return &FooAPI{
		server: srv,
	}
}

// a valid API method has a pointer receiver and returns error as last argument
func (api *FooAPI) NodeInfo() (*p2p.NodeInfo, error) {
	return api.server.NodeInfo(), nil
}

func main() {

	// make a new private key
	privkey, err := crypto.GenerateKey()
	if err != nil {
		demo.Log.Crit("Generate private key failed", "err", err)
	}

	// set up p2p server
	cfg := p2p.Config{
		PrivateKey: privkey,
		Name:       common.MakeName("foo", "42"),
	}
	srv := p2p.Server{
		Config: cfg,
	}

	// attempt to start the p2p server
	err = srv.Start()
	if err != nil {
		demo.Log.Crit("Start p2p.Server failed", "err", err)
	}

	// set up the RPC server
	rpcsrv := rpc.NewServer()
	err = rpcsrv.RegisterName("foo", NewFooAPI(&srv))
	if err != nil {
		demo.Log.Crit("Register API method(s) fail", "err", err)
	}

	// create IPC endpoint
	ipcpath := "demo.ipc"
	ipclistener, err := rpc.CreateIPCListener(ipcpath)
	if err != nil {
		demo.Log.Crit("IPC endpoint create fail", "err", err)
	}

	// mount RPC server on IPC endpoint
	go func() {
		err = rpcsrv.ServeListener(ipclistener)
		if err != nil {
			demo.Log.Crit("Mount RPC on IPC fail", "err", err)
		}
	}()

	// create a IPC client
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		demo.Log.Crit("IPC dial fail", "err", err)
	}

	// call the RPC method
	var nodeinfo p2p.NodeInfo
	err = rpcclient.Call(&nodeinfo, "foo_nodeInfo")
	if err != nil {
		demo.Log.Crit("RPC call fail", "err", err)
	}
	demo.Log.Info("server started", "enode", nodeinfo.Enode, "name", nodeinfo.Name, "ID", nodeinfo.ID, "IP", nodeinfo.IP)

	// bring down the servers
	rpcsrv.Stop()
	srv.Stop()
}
