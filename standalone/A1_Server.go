package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"os"
)

func init() {
	loglevel := log.LvlTrace
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)
}

func main() {

	// make a new private key
	privkey, err := crypto.GenerateKey()
	if err != nil {
		log.Crit("Generate private key failed", "err", err)
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
		log.Crit("Start p2p.Server failed", "err", err)
	}

	// inspect the resulting values
	nodeinfo := srv.NodeInfo()
	log.Info("server started", "enode", nodeinfo.Enode, "name", nodeinfo.Name, "ID", nodeinfo.ID, "IP", nodeinfo.IP)

	// bring down the server
	srv.Stop()
}
