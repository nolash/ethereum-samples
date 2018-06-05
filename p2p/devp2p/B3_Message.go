// trigger p2p message with RPC
package main

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	demo "./common"
)

var (
	protoW   = &sync.WaitGroup{}
	messageW = &sync.WaitGroup{}
	msgC     = make(chan string)
	ipcpath  = ".demo.ipc"
)

// create a protocol that can take care of message sending
// the Run function is invoked upon connection
// it gets passed:
// * an instance of p2p.Peer, which represents the remote peer
// * an instance of p2p.MsgReadWriter, which is the io between the node and its peer

type FooMsg struct {
	Content string
}

var (
	proto = p2p.Protocol{
		Name:    "foo",
		Version: 42,
		Length:  1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {

			// only one of the peers will send this
			content, ok := <-msgC
			if ok {
				outmsg := &FooMsg{
					Content: content,
				}

				// send the message
				err := p2p.Send(rw, 0, outmsg)
				if err != nil {
					return fmt.Errorf("Send p2p message fail: %v", err)
				}
				demo.Log.Info("sending message", "peer", p, "msg", outmsg)
			}

			// wait for the subscriptions to end
			messageW.Wait()
			protoW.Done()

			// terminate the protocol
			return nil
		},
	}
)

type FooAPI struct {
	sent bool
}

func (api *FooAPI) SendMsg(content string) error {
	if api.sent {
		return fmt.Errorf("Already sent")
	}
	msgC <- content
	close(msgC)
	api.sent = true
	return nil
}

// create a server
func newP2pServer(privkey *ecdsa.PrivateKey, name string, version string, port int) *p2p.Server {
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

func newRPCServer() (*rpc.Server, error) {
	// set up the RPC server
	rpcsrv := rpc.NewServer()
	err := rpcsrv.RegisterName("foo", &FooAPI{})
	if err != nil {
		return nil, fmt.Errorf("Register API method(s) fail: %v", err)
	}

	// create IPC endpoint
	ipclistener, err := net.Listen("unix", ipcpath)
	if err != nil {
		return nil, fmt.Errorf("IPC endpoint create fail: %v", err)
	}

	// mount RPC server on IPC endpoint
	// it will automatically detect and serve any valid methods
	go func() {
		err = rpcsrv.ServeListener(ipclistener)
		if err != nil {
			demo.Log.Crit("Mount RPC on IPC fail", "err", err)
		}
	}()

	return rpcsrv, nil
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
	srv_one := newP2pServer(privkey_one, "foo", "42", 0)
	err = srv_one.Start()
	if err != nil {
		demo.Log.Crit("Start p2p.Server #1 failed", "err", err)
	}

	srv_two := newP2pServer(privkey_two, "bar", "666", 31234)
	err = srv_two.Start()
	if err != nil {
		demo.Log.Crit("Start p2p.Server #2 failed", "err", err)
	}

	// set up the event subscriptions on both servers
	// the Err() on the Subscription object returns when subscription is closed
	eventOneC := make(chan *p2p.PeerEvent)
	sub_one := srv_one.SubscribeEvents(eventOneC)
	go func() {
		for {
			select {
			case peerevent := <-eventOneC:
				if peerevent.Type == "add" {
					demo.Log.Debug("Received peer add notification on node #1", "peer", peerevent.Peer)
				} else if peerevent.Type == "msgsend" {
					demo.Log.Info("Received message send notification on node #1", "event", peerevent)
					messageW.Done()
				}
			case <-sub_one.Err():
				return
			}
		}
	}()

	eventTwoC := make(chan *p2p.PeerEvent)
	sub_two := srv_two.SubscribeEvents(eventTwoC)
	go func() {
		for {
			select {
			case peerevent := <-eventTwoC:
				if peerevent.Type == "add" {
					demo.Log.Debug("Received peer add notification on node #2", "peer", peerevent.Peer)
				} else if peerevent.Type == "msgsend" {
					demo.Log.Info("Received message send notification on node #2", "event", peerevent)
					messageW.Done()
				}
			case <-sub_two.Err():
				return
			}
		}
	}()

	// create and start RPC server
	rpcsrv, err := newRPCServer()
	if err != nil {
		demo.Log.Crit(err.Error())
	}
	defer os.Remove(ipcpath)

	// get the node instance of the second server
	node_two := srv_two.Self()

	// add it as a peer to the first node
	// the connection and crypto handshake will be performed automatically
	srv_one.AddPeer(node_two)

	// create an IPC client
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		demo.Log.Crit("IPC dial fail", "err", err)
	}

	// wait for one message be sent, and both protocols to end
	messageW.Add(1)
	protoW.Add(2)

	// call the RPC method
	err = rpcclient.Call(nil, "foo_sendMsg", "foobar")
	if err != nil {
		demo.Log.Crit("RPC call fail", "err", err)
	}

	// wait for protocols to finish
	protoW.Wait()

	// terminate subscription loops and unsubscribe
	sub_one.Unsubscribe()
	sub_two.Unsubscribe()

	// stop the servers
	rpcsrv.Stop()
	srv_one.Stop()
	srv_two.Stop()
}
