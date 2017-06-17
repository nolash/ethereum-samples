// three ways of accessing the API from a started node
package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

var (
	quitC = make(chan bool)
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

	go func() {
		// get the info directly via the p2p server object
		p2pserver := demo.ServiceNode.Server()
		localnodeinfo := p2pserver.NodeInfo()
		demo.Log.Info("Nodeinfo from p2p.Server", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

		// get the nodeinfo via ServiceNode IPC
		localnodeinfo = &p2p.NodeInfo{}
		rpcclient, err := demo.ServiceNode.Attach()
		err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
		if err != nil {
			demo.Log.Error("Could not get rpcclient via p2p.Server", "err", err)
			quitC <- false
			return

		}
		demo.Log.Info("Nodeinfo from IPC via ServiceNode", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

		// get the nodeinfo via external IPC
		rpcclient, err = rpc.Dial(fmt.Sprintf("%s/%s", cfg.DataDir, cfg.IPCPath))
		if err != nil {
			demo.Log.Error("Could not get rpcclient via p2p.Server", "err", err)
			quitC <- false
		}
		localnodeinfo = &p2p.NodeInfo{}
		rpcclient, err = demo.ServiceNode.Attach()
		err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
		demo.Log.Info("Nodeinfo from IPC via external call", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)
		quitC <- true
	}()

	ok := <-quitC
	if !ok {
		demo.Log.Crit("Oh-ooh, Spaghettios")
	}
	demo.Log.Info("That's all, folks")
}
