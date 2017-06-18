// using RPC through HTTP and Websockets
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
	// add these parameters to activate HTTP
	cfg.HTTPHost = node.DefaultHTTPHost
	cfg.HTTPPort = node.DefaultHTTPPort
	cfg.HTTPModules = append(cfg.HTTPModules, "admin")
	// add these paramters to activate Websockets
	cfg.WSHost = node.DefaultWSHost
	cfg.WSPort = node.DefaultWSPort

	demo.ServiceNode, err = node.New(cfg)
	if err != nil {
		demo.Log.Crit("ServiceNode create fail", "err", err)
	}
	err = demo.ServiceNode.Start()
	if err != nil {
		demo.Log.Crit("ServiceNode start fail", "err", err)
	}

	go func() {
		var localnodeinfo p2p.NodeInfo

		// get the info from HTTP RPC
		rpcclient, err := rpc.Dial(fmt.Sprintf("http://%s:%d", cfg.HTTPHost, cfg.HTTPPort))
		if err != nil {
			demo.Log.Error("HTTP RPC connect failed", "err", err, "host", cfg.HTTPHost, "port", cfg.HTTPPort)
			quitC <- false
			return
		}
		err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
		if err != nil {
			demo.Log.Error("HTTP RPC call failed", "err", err)
			quitC <- false
			return
		}
		demo.Log.Info("node version from HTTP RPC", "enode", localnodeinfo.Enode)

		// get the info from Websocket RPC
		rpcclient, err = rpc.Dial(fmt.Sprintf("ws://%s:%d", cfg.WSHost, cfg.WSPort))
		if err != nil {
			demo.Log.Error("Websocket RPC connect failed", "err", err, "host", cfg.WSHost, "port", cfg.WSPort)
			quitC <- false
			return
		}
		err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
		if err == nil {
			demo.Log.Error("Websocket RPC call should have failed")
			quitC <- false
			return
		}
		demo.Log.Info("Websocket RPC fails as expected", "err", err)

		quitC <- true
	}()

	ok := <-quitC
	if !ok {
		demo.Log.Crit("Oh-ooh, Spaghettios")
	}
	demo.Log.Info("That's all, folks")
}
