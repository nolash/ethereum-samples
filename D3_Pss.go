// pss send-to-self hello world
package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/pss"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

func main() {

	// create servicenode
	stack, err := demo.NewServiceNode(demo.P2PDefaultPort, 0, 0)
	if err != nil {
		demo.Log.Crit("Create servicenode #1 fail", "err", err)
	}

	// add the pss service
	psssvc := func(ctx *node.ServiceContext) (node.Service, error) {

		// we will use a made-up address
		// swarm overlay address generation can be arbitrary
		addr := network.RandomAddr()

		// set up kademlia overlay
		// it needs to exist, but it will not be operational in this example
		overlayparams := network.NewKadParams()
		overlay := network.NewKademlia(addr.Over(), overlayparams)

		// the true param enables send-to-self test feature
		psparams := pss.NewPssParams(true)

		// set up pss
		// we don't include DPA for now
		ps := pss.NewPss(overlay, nil, psparams)
		if ps == nil {
			return nil, fmt.Errorf("pss create fail: %v", err)
		}
		return ps, nil
	}

	// register the pss service
	err = stack.Register(psssvc)
	if err != nil {
		demo.Log.Crit("servicenode pss register fail", "err", err)
	}

	// start the node
	err = stack.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}

	// topic for our pss context
	// this will be used to match messages to a message handler
	topic := pss.NewTopic(demo.FooProtocolName, demo.FooProtocolVersion)

	// subscribe to incoming messages
	// this will register a message handler on the specified topic
	msgC := make(chan pss.APIMsg)
	rpcclient, err := stack.Attach()
	sub, err := rpcclient.Subscribe(context.Background(), "pss", msgC, "receive", topic)

	// get our swarm overlay address
	var bzzaddr []byte
	err = rpcclient.Call(&bzzaddr, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("pss get addr fail", "err", err)
	}

	// send message
	// since it's sent to ourselves, it will not go through pss forwarding
	apimsg := &pss.APIMsg{
		Addr: bzzaddr,
		Msg:  []byte("foobar"),
	}
	err = rpcclient.Call(nil, "pss_send", topic, apimsg)
	if err != nil {
		demo.Log.Crit("pss send fail", "err", err)
	}

	// get the incoming message
	inmsg := <-msgC
	demo.Log.Info("pss received", "msg", string(inmsg.Msg), "from", fmt.Sprintf("%x", inmsg.Addr))

	// bring down the servicenode
	sub.Unsubscribe()
	stack.Stop()
}
