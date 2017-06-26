// set up boilerplate service node and start it
package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"os"
)

const (
	p2pDefaultPort = 30100
	ipcName        = "demo.ipc"
	datadirPrefix  = ".data_"
)

func init() {
	loglevel := log.LvlTrace
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)
}

func main() {
	var err error

	// set up the local service node
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", p2pDefaultPort)
	cfg.IPCPath = ipcName
	cfg.DataDir = fmt.Sprintf("%s%d", datadirPrefix, p2pDefaultPort)
	stack, err := node.New(cfg)
	if err != nil {
		log.Crit("ServiceNode create fail", "err", err)
	}
	err = stack.Start()
	if err != nil {
		log.Crit("ServiceNode start fail", "err", err)
	}
}
