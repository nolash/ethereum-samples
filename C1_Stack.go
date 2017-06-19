// set up boilerplate service node and start it
package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/node"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

const (
	p2pDefaultPort = 30100
	ipcName        = "demo.ipc"
	datadirPrefix  = ".data_"
)

func main() {
	var err error

	// set up the local service node
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", p2pDefaultPort)
	cfg.IPCPath = ipcName
	cfg.DataDir = fmt.Sprintf("%s%d", datadirPrefix, p2pDefaultPort)
	stack, err := node.New(cfg)
	if err != nil {
		demo.Log.Crit("ServiceNode create fail", "err", err)
	}
	err = stack.Start()
	if err != nil {
		demo.Log.Crit("ServiceNode start fail", "err", err)
	}
}
