// get notified when the peer connection has been completed
package main

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"

	demo "./common"
)

var (
	quitC = make(chan bool)
)

// create a server
func newServer(privkey *ecdsa.PrivateKey, name string, version string, port int) *p2p.Server {

	// we need to explicitly allow at least one peer, otherwise the connection attempt will be refused
	cfg := p2p.Config{
		PrivateKey: privkey,
		Name:       common.MakeName(name, version),
		MaxPeers:   1,
	}
	if port > 0 {
		cfg.ListenAddr = fmt.Sprintf(":%d", port)
	}
	srv := &p2p.Server{
		Config: cfg,
	}
	return srv
}

func main() {

	// we need private keys for both servers
	privkey_one, err := crypto.GenerateKey()
	if err != nil {
		demo.Log.Crit("Generate private key #1 failed", "err", err)
	}
	privkey_two, err := crypto.GenerateKey()
	if err != nil {
		demo.Log.Crit("Generate private key #2 failed", "err", err)
	}

	// set up the two servers
	srv_one := newServer(privkey_one, "foo", "42", 0)
	err = srv_one.Start()
	if err != nil {
		demo.Log.Crit("Start p2p.Server #1 failed", "err", err)
	}

	srv_two := newServer(privkey_two, "bar", "666", 31234)
	err = srv_two.Start()
	if err != nil {
		demo.Log.Crit("Start p2p.Server #2 failed", "err", err)
	}

	// set up the event subscription on the first server
	eventC := make(chan *p2p.PeerEvent)
	sub_one := srv_one.SubscribeEvents(eventC)

	// listen for events
	go func() {
		peerevent := <-eventC
		demo.Log.Info("received peerevent", "type", peerevent.Type, "peer", peerevent.Peer)
		quitC <- true
	}()

	// get the node instance of the second server
	node_two := srv_two.Self()

	// add it as a peer to the first node
	// the connection and crypto handshake will be performed automatically
	srv_one.AddPeer(node_two)

	// receives when the event is received
	<-quitC

	// inspect the results
	demo.Log.Info("after add", "node one peers", srv_one.Peers(), "node two peers", srv_two.Peers())

	// terminate subscription
	sub_one.Unsubscribe()

	// stop the servers
	srv_one.Stop()
	srv_two.Stop()
}
