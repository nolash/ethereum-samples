// pss RPC routed over swarm
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/chequebook"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/swarm"
	bzzapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/pss"
	pssclient "github.com/ethereum/go-ethereum/swarm/pss/client"

	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

// simple ping and receive protocol
var (
	proto = p2p.Protocol{
		Name:    demo.FooProtocolName,
		Version: demo.FooProtocolVersion,
		Length:  uint64(len(demo.FooMessages)),
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			pp := protocols.NewPeer(p, rw, &demo.FooProtocol)
			go func() {
				err := pp.Send(&demo.FooPingMsg{
					Created: time.Now(),
				})
				if err != nil {
					demo.Log.Error("protocol send fail", "err", err)
				}
			}()
			return pp.Run(run)
		},
	}
	topic = pss.PingTopic
	quitC = make(chan struct{})
)

// receive message and quit
func run(msg interface{}) error {
	foomsg, ok := msg.(*demo.FooPingMsg)
	if !ok {
		return fmt.Errorf("invalid msg: %v", msg)
	}
	demo.Log.Info("received message", "time", foomsg.Created)
	quitC <- struct{}{}
	return nil
}

func newService(bzzdir string, bzzport int, bzznetworkid uint64) func(ctx *node.ServiceContext) (node.Service, error) {
	return func(ctx *node.ServiceContext) (node.Service, error) {

		// generate a new private key
		privkey, err := crypto.GenerateKey()
		if err != nil {
			demo.Log.Crit("private key generate servicenode 'left' fail: %v")
		}

		// create necessary swarm params
		var ensApi chequebook.Backend = nil
		bzzconfig := bzzapi.NewConfig()
		bzzconfig.Path = bzzdir
		bzzconfig.Init(privkey)
		if err != nil {
			demo.Log.Crit("unable to configure swarm", "err", err)
		}
		bzzconfig.Port = fmt.Sprintf("%d", bzzport)

		// shortcut to setting up a swarm node
		return swarm.NewSwarm(ctx, ensApi, bzzconfig, nil)

	}
}

func main() {

	// create two nodes
	l_stack, err := demo.NewServiceNode(demo.P2pPort, 0, demo.WSDefaultPort, "pss")
	if err != nil {
		demo.Log.Crit(err.Error())
	}
	r_stack, err := demo.NewServiceNode(demo.P2pPort+1, 0, demo.WSDefaultPort+1, "pss")
	if err != nil {
		demo.Log.Crit(err.Error())
	}

	// register the pss activated bzz services
	l_svc := newService(l_stack.InstanceDir(), demo.BzzDefaultPort, demo.BzzDefaultNetworkId)
	err = l_stack.Register(l_svc)
	if err != nil {
		demo.Log.Crit("servicenode 'left' pss register fail", "err", err)
	}

	r_svc := newService(r_stack.InstanceDir(), demo.BzzDefaultPort+1, demo.BzzDefaultNetworkId)
	err = r_stack.Register(r_svc)
	if err != nil {
		demo.Log.Crit("servicenode 'right' pss register fail", "err", err)
	}

	// start the nodes
	err = l_stack.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}
	defer os.RemoveAll(l_stack.DataDir())
	err = r_stack.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}
	defer os.RemoveAll(r_stack.DataDir())

	// connect the nodes to the middle
	l_stack.Server().AddPeer(r_stack.Server().Self())

	// get the rpc clients
	l_rpcclient, err := l_stack.Attach()
	r_rpcclient, err := r_stack.Attach()

	// get the public keys
	var l_pubkey string
	err = l_rpcclient.Call(&l_pubkey, "pss_getPublicKey")
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}
	var r_pubkey string
	err = r_rpcclient.Call(&r_pubkey, "pss_getPublicKey")
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}

	// get the overlay addresses
	var l_bzzaddr string
	err = l_rpcclient.Call(&l_bzzaddr, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("pss get baseaddr fail", "err", err)
	}
	var r_bzzaddr string
	err = r_rpcclient.Call(&r_bzzaddr, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("pss get baseaddr fail", "err", err)
	}

	// make the nodes aware of each others' public keys
	err = l_rpcclient.Call(nil, "pss_setPeerPublicKey", r_pubkey, topic, r_bzzaddr)
	if err != nil {
		demo.Log.Crit("pss set pubkey fail", "err", err)
	}
	err = r_rpcclient.Call(nil, "pss_setPeerPublicKey", l_pubkey, topic, l_bzzaddr)
	if err != nil {
		demo.Log.Crit("pss set pubkey fail", "err", err)
	}

	// wait until the state of the swarm overlay network is ready
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = demo.WaitHealthy(ctx, 2, l_rpcclient, r_rpcclient)
	if err != nil {
		demo.Log.Crit("health check fail", "err", err)
	}
	time.Sleep(time.Second) // because the healthy does not work

	// configure and start up pss client RPCs
	c_left, err := pssclient.NewClient(fmt.Sprintf("ws://localhost:%d", demo.WSDefaultPort))
	if err != nil {
		demo.Log.Crit("pssclient 'left' create fail", "err", err)
	}
	c_right, err := pssclient.NewClient(fmt.Sprintf("ws://localhost:%d", demo.WSDefaultPort+1))
	if err != nil {
		demo.Log.Crit("pssclient 'right' create fail", "err", err)
	}

	// set up generic ping protocol
	l_ping := pss.Ping{
		Pong: false,
		OutC: make(chan bool),
		InC:  make(chan bool),
	}
	l_proto := pss.NewPingProtocol(&l_ping)
	r_ping := pss.Ping{
		Pong: true,
		OutC: make(chan bool),
		InC:  make(chan bool),
	}
	r_proto := pss.NewPingProtocol(&r_ping)

	// run the pssclient protocols
	// this registers the protocol handler in pss with topic generated from the protocol name and version
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c_left.RunProtocol(ctx, l_proto)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c_right.RunProtocol(ctx, r_proto)

	// add the 'right' peer
	c_left.AddPssPeer(r_pubkey, common.FromHex(r_bzzaddr), pss.PingProtocol)

	time.Sleep(time.Second)

	// send ping
	l_ping.OutC <- false

	// get ping
	<-r_ping.InC
	demo.Log.Info("got ping")

	// get pong
	<-l_ping.InC
	demo.Log.Info("got pong")

	// the 'right' will receive the ping and send on the quit channel
	c_left.Close()
	c_right.Close()
}
