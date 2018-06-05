// set up boilerplate service node and start it
package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/node"

	demo "./common"
)

const (
	p2pDefaultPort = 30100
	ipcpath        = ".demo.ipc"
	datadirPrefix  = ".data_"
)

func main() {
	// set up the service node
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", p2pDefaultPort)
	cfg.IPCPath = ipcpath
	cfg.DataDir = fmt.Sprintf("%s%d", datadirPrefix, p2pDefaultPort)

	// create the node instance with the config
	stack, err := node.New(cfg)
	if err != nil {
		demo.Log.Crit("ServiceNode create fail", "err", err)
	}

	// start the node
	err = stack.Start()
	if err != nil {
		demo.Log.Crit("ServiceNode start fail", "err", err)
	}
	defer os.RemoveAll(stack.DataDir())

	// shut down
	err = stack.Stop()
	if err != nil {
		demo.Log.Crit("Node stop fail", "err", err)
	}
}
