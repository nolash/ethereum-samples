// Different ways of accessing RPC API on a servicenode
package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

func main() {

	// start servicenode
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", demo.P2PDefaultPort)
	cfg.IPCPath = demo.IPCName
	cfg.DataDir = demo.Datadir(demo.P2PDefaultPort)
	stack, err := node.New(cfg)
	if err != nil {
		demo.Log.Crit("ServiceNode create fail", "err", err)
	}
	err = stack.Start()
	if err != nil {
		demo.Log.Crit("ServiceNode start fail", "err", err)
	}

	// get the info directly via the p2p server object
	p2pserver := stack.Server()
	localnodeinfo := p2pserver.NodeInfo()
	demo.Log.Info("Nodeinfo from p2p.Server", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

	// get the nodeinfo via ServiceNode IPC
	localnodeinfo = &p2p.NodeInfo{}
	rpcclient, err := stack.Attach()
	err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
	if err != nil {
		demo.Log.Crit("Could not get rpcclient via p2p.Server", "err", err)

	}
	demo.Log.Info("Nodeinfo from IPC via ServiceNode", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

	// get the nodeinfo via external IPC
	rpcclient, err = rpc.Dial(fmt.Sprintf("%s/%s", demo.Datadir(demo.P2PDefaultPort), demo.IPCName))

	if err != nil {
		demo.Log.Crit("Could not get rpcclient via p2p.Server", "err", err)
	}
	localnodeinfo = &p2p.NodeInfo{}
	rpcclient, err = stack.Attach()
	err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
	demo.Log.Info("Nodeinfo from IPC via external call", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

	err = stack.Stop()
	if err != nil {
		demo.Log.Crit("Node stop fail", "err", err)
	}
}
