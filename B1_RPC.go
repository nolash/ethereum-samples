// RPC hello world
package main

import (
	"net"
	"os"

	"github.com/ethereum/go-ethereum/rpc"

	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

// set up an object that can contain the API methods
type FooAPI struct {
}

// a valid API method is exported, has a pointer receiver and returns error as last argument
// the method will be called with <registeredname>_helloWorld
// (first letter in method is lowercase, module name and method name separated by underscore)
func (api *FooAPI) HelloWorld() (string, error) {
	return "foobar", nil
}

func main() {

	// set up the RPC server
	rpcsrv := rpc.NewServer()
	err := rpcsrv.RegisterName("foo", &FooAPI{})
	if err != nil {
		demo.Log.Crit("Register API method(s) fail", "err", err)
	}

	// create IPC endpoint
	ipcpath := ".demo.ipc"
	ipclistener, err := net.Listen("unix", ipcpath)
	if err != nil {
		demo.Log.Crit("IPC endpoint create fail", "err", err)
	}
	defer os.Remove(ipcpath)

	// mount RPC server on IPC endpoint
	// it will automatically detect and serve any valid methods
	go func() {
		err = rpcsrv.ServeListener(ipclistener)
		if err != nil {
			demo.Log.Crit("Mount RPC on IPC fail", "err", err)
		}
	}()

	// create an IPC client
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		demo.Log.Crit("IPC dial fail", "err", err)
	}

	// call the RPC method
	var result string
	err = rpcclient.Call(&result, "foo_helloWorld")
	if err != nil {
		demo.Log.Crit("RPC call fail", "err", err)
	}

	// inspect the results
	demo.Log.Info("RPC return value", "reply", result)

	// bring down the server
	rpcsrv.Stop()
}
