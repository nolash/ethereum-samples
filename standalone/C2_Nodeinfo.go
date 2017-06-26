// Different ways of accessing RPC API on a servicenode
package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"os"
)

const (
	p2pDefaultPort = 30100
	datadirPrefix  = ".datadir_"
	ipcName        = ".demo.ipc"
)

func datadir(port int) string {
	return fmt.Sprintf("./%s%d", datadirPrefix, port)
}

func init() {
	loglevel := log.LvlTrace
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)
}

func main() {

	// start servicenode
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", p2pDefaultPort)
	cfg.IPCPath = ipcName
	cfg.DataDir = datadir(p2pDefaultPort)
	stack, err := node.New(cfg)
	if err != nil {
		log.Crit("ServiceNode create fail", "err", err)
	}
	err = stack.Start()
	if err != nil {
		log.Crit("ServiceNode start fail", "err", err)
	}

	// get the info directly via the p2p server object
	p2pserver := stack.Server()
	localnodeinfo := p2pserver.NodeInfo()
	log.Info("Nodeinfo from p2p.Server", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

	// get the nodeinfo via ServiceNode IPC
	localnodeinfo = &p2p.NodeInfo{}
	rpcclient, err := stack.Attach()
	err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
	if err != nil {
		log.Crit("Could not get rpcclient via p2p.Server", "err", err)

	}
	log.Info("Nodeinfo from IPC via ServiceNode", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

	// get the nodeinfo via external IPC
	rpcclient, err = rpc.Dial(fmt.Sprintf("%s/%s", datadir(p2pDefaultPort), ipcName))

	if err != nil {
		log.Crit("Could not get rpcclient via p2p.Server", "err", err)
	}
	localnodeinfo = &p2p.NodeInfo{}
	rpcclient, err = stack.Attach()
	err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
	log.Info("Nodeinfo from IPC via external call", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

	err = stack.Stop()
	if err != nil {
		log.Crit("Node stop fail", "err", err)
	}
}
