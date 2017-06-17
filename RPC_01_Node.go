// set up a service node and start it
package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/node"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

func main() {
	var err error

	// set up the local service node
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", demo.P2PDefaultPort)
	cfg.IPCPath = demo.IPCName
	cfg.DataDir = fmt.Sprintf("%s/%s%d", demo.OurDir, demo.DatadirPrefix, demo.P2PDefaultPort)
	demo.ServiceNode, err = node.New(cfg)
	if err != nil {
		demo.Log.Crit("ServiceNode create fail", "err", err)
	}
	err = demo.ServiceNode.Start()
	if err != nil {
		demo.Log.Crit("ServiceNode start fail", "err", err)
	}

	demo.Log.Info("That's all, folks")
}
