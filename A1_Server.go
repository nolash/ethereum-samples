package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

func main() {

	// make a new private key
	privkey, err := crypto.GenerateKey()
	if err != nil {
		demo.Log.Crit("Generate private key failed", "err", err)
	}

	// set up server
	cfg := p2p.Config{
		PrivateKey: privkey,
		Name:       common.MakeName("foo", "42"),
	}
	srv := p2p.Server{
		Config: cfg,
	}

	// attempt to start the server
	err = srv.Start()
	if err != nil {
		demo.Log.Crit("Start p2p.Server failed", "err", err)
	}

	// inspect the resulting values
	nodeinfo := srv.NodeInfo()
	demo.Log.Info("server started", "enode", nodeinfo.Enode, "name", nodeinfo.Name, "ID", nodeinfo.ID, "IP", nodeinfo.IP)

	// bring down the server
	srv.Stop()
}
