// send, receive, get notified about a message
package main

import (
	"crypto/ecdsa"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"

	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

var (
	messageW = &sync.WaitGroup{}
)

type FooMsg struct {
	V uint
}

// create a protocol that can take care of message sending
// the Run function is invoked upon connection
// it gets passed:
// * an instance of p2p.Peer, which represents the remote peer
// * an instance of p2p.MsgReadWriter, which is the io between the node and its peer
var (
	proto = p2p.Protocol{
		Name:    "foo",
		Version: 42,
		Length:  1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {

			// simplest payload possible; a byte slice
			outmsg := "foobar"

			// send the message
			err := p2p.Send(rw, 0, outmsg)
			if err != nil {
				return fmt.Errorf("Send p2p message fail: %v", err)
			}
			demo.Log.Info("sending message", "peer", p, "msg", outmsg)

			// wait for the message to come in from the other side
			// note that receive message event doesn't get emitted until we ReadMsg()
			inmsg, err := rw.ReadMsg()
			if err != nil {
				return fmt.Errorf("Receive p2p message fail: %v", err)
			}
			demo.Log.Info("received message", "peer", p, "msg", inmsg)

			// terminate the protocol
			return nil
		},
	}
)

// create a server
func newServer(privkey *ecdsa.PrivateKey, name string, version string, port int) *p2p.Server {

	// we need to explicitly allow at least one peer, otherwise the connection attempt will be refused
	// we also need to explicitly tell the server to generate events for messages
	cfg := p2p.Config{
		PrivateKey:      privkey,
		Name:            common.MakeName(name, version),
		MaxPeers:        1,
		Protocols:       []p2p.Protocol{proto},
		EnableMsgEvents: true,
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

	// set up the event subscriptions on both servers
	// the Err() on the Subscription object returns when subscription is closed
	eventOneC := make(chan *p2p.PeerEvent)
	sub_one := srv_one.SubscribeEvents(eventOneC)
	messageW.Add(1)
	go func() {
		for {
			select {
			case peerevent := <-eventOneC:
				if peerevent.Type == "add" {
					demo.Log.Debug("Received peer add notification on node #1", "peer", peerevent.Peer)
				} else if peerevent.Type == "msgrecv" {
					demo.Log.Info("Received message nofification on node #1", "event", peerevent)
					messageW.Done()
				}
			case <-sub_one.Err():
				return
			}
		}
	}()

	eventTwoC := make(chan *p2p.PeerEvent)
	sub_two := srv_two.SubscribeEvents(eventTwoC)
	messageW.Add(1)
	go func() {
		for {
			select {
			case peerevent := <-eventTwoC:
				if peerevent.Type == "add" {
					demo.Log.Debug("Received peer add notification on node #2", "peer", peerevent.Peer)
				} else if peerevent.Type == "msgrecv" {
					demo.Log.Info("Received message nofification on node #2", "event", peerevent)
					messageW.Done()
				}
			case <-sub_two.Err():
				return
			}
		}
	}()

	// get the node instance of the second server
	node_two := srv_two.Self()

	// add it as a peer to the first node
	// the connection and crypto handshake will be performed automatically
	srv_one.AddPeer(node_two)

	// wait for each respective message to be delivered on both sides
	messageW.Wait()

	// terminate subscription loops and unsubscribe
	sub_one.Unsubscribe()
	sub_two.Unsubscribe()

	// stop the servers
	srv_one.Stop()
	srv_two.Stop()
}
