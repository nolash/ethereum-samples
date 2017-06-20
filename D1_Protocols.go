// Previous "reply" example using p2p.protocols abstraction
package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
	"sync"
)

var (
	messageW = &sync.WaitGroup{}
)

type FooMsg struct {
	V uint
}

// using the protocols abstraction, message structures are registered and their message codes handled automatically
var (
	fooProtocol = protocols.Spec{
		Name:       demo.FooProtocolName,
		Version:    demo.FooProtocolVersion,
		MaxMsgSize: demo.FooProtocolMaxMsgSize,
		Messages: []interface{}{
			&FooMsg{},
		},
	}
)

// the protocols abstraction enables use of an external handler function
type fooHandler struct {
	peer *p2p.Peer
}

func (self *fooHandler) handle(msg interface{}) error {
	foomsg, ok := msg.(*FooMsg)
	if !ok {
		return fmt.Errorf("invalid message", "msg", msg, "peer", self.peer)
	}
	demo.Log.Info("received message", "foomsg", foomsg, "peer", self.peer)
	return nil
}

// create the protocol with the protocols extension
var (
	proto = p2p.Protocol{
		Name:    "foo",
		Version: 42,
		Length:  1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {

			// create the enhanced peer
			pp := protocols.NewPeer(p, rw, &fooProtocol)

			// send the message
			go func() {
				outmsg := &FooMsg{
					V: 42,
				}
				err := pp.Send(outmsg)
				if err != nil {
					demo.Log.Error("Send p2p message fail", "err", err)
				}
				demo.Log.Info("sending message", "peer", p, "msg", outmsg)
			}()

			// protocols abstraction provides a separate blocking run loop for the peer
			// when this returns, the protocol will be terminated
			run := &fooHandler{
				peer: p,
			}
			err := pp.Run(run.handle)
			return err
		},
	}
)

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
	srv_one := demo.NewServer(privkey_one, "foo", "42", proto, 0)
	err = srv_one.Start()
	if err != nil {
		demo.Log.Crit("Start p2p.Server #1 failed", "err", err)
	}

	srv_two := demo.NewServer(privkey_two, "bar", "666", proto, 31234)
	err = srv_two.Start()
	if err != nil {
		demo.Log.Crit("Start p2p.Server #2 failed", "err", err)
	}

	// set up the event subscriptions on both servers
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
