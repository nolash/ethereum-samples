// pss RPC routed over swarm
package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/swarm"
	bzzapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/pss"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
	"time"
)

const (
	bzzNetworkId = 3
	bzzPort      = 30399
)

func newSwarmService(stack *node.Node, bzzport int) func(ctx *node.ServiceContext) (node.Service, error) {
	return func(ctx *node.ServiceContext) (node.Service, error) {
		// get the encrypted private key file
		keyid := fmt.Sprintf("%s/D3_Pss/nodekey", stack.DataDir())

		// load the private key from the file content
		prvkey, err := crypto.LoadECDSA(keyid)
		if err != nil {
			return nil, fmt.Errorf("privkey fail: %v", prvkey)
		}

		// create the swarm overlay address
		chbookaddr := crypto.PubkeyToAddress(prvkey.PublicKey)

		// configure and create a swarm instance
		bzzdir := stack.InstanceDir() // todo: what is the difference between this and datadir?

		swapEnabled := false
		syncEnabled := false
		pssEnabled := true
		cors := "*"

		bzzconfig, err := bzzapi.NewConfig(bzzdir, chbookaddr, prvkey, bzzNetworkId)
		bzzconfig.Port = fmt.Sprintf("%s", bzzport)
		if err != nil {
			demo.Log.Crit("unable to configure swarm", "err", err)
		}
		return swarm.NewSwarm(ctx, nil, bzzconfig, swapEnabled, syncEnabled, cors, pssEnabled)
	}
}

func main() {

	stack_left, err := demo.NewServiceNode(demo.P2PDefaultPort, 0, 0)
	stack_mid, err := demo.NewServiceNode(demo.P2PDefaultPort+1, 0, 0)
	stack_right, err := demo.NewServiceNode(demo.P2PDefaultPort+2, 0, 0)

	// pss is enclosed by swarm
	// the overlay of pss and swram will be the same

	err = stack_left.Register(newSwarmService(stack_left, 30399))
	if err != nil {
		demo.Log.Crit("bzz service 'left' register fail", "err", err)
	}
	err = stack_mid.Register(newSwarmService(stack_mid, 30400))
	if err != nil {
		demo.Log.Crit("bzz service 'mid' register fail", "err", err)
	}
	err = stack_right.Register(newSwarmService(stack_right, 30401))
	if err != nil {
		demo.Log.Crit("bzz service 'right' register fail", "err", err)
	}

	// start the node
	err = stack_left.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}
	err = stack_mid.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}
	err = stack_right.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}

	// set the shared value in service bar
	var baseaddr_left []byte
	rpcclient_left, err := stack_left.Attach()
	err = rpcclient_left.Call(&baseaddr_left, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("Could not get rpcclient 'left' via p2p.Server", "err", err)
	}

	var baseaddr_right []byte
	rpcclient_right, err := stack_right.Attach()
	err = rpcclient_right.Call(&baseaddr_right, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("Could not get rpcclient 'right' via p2p.Server", "err", err)
	}

	demo.Log.Debug("pss send", "left", baseaddr_left, "leftlen", len(baseaddr_left), "right", baseaddr_right, "rightlen", len(baseaddr_right))

	// connect the nodes
	p2pnode_mid := stack_mid.Server().Self()
	stack_left.Server().AddPeer(p2pnode_mid)
	stack_right.Server().AddPeer(p2pnode_mid)

	time.Sleep(time.Millisecond * 250)
	// create a subscription on the receiver node.
	topic := pss.NewTopic(demo.FooProtocolName, demo.FooProtocolVersion)
	msgC := make(chan pss.APIMsg)

	sub, err := rpcclient_right.Subscribe(context.Background(), "pss", msgC, "receive", topic)
	if err != nil {
		demo.Log.Crit("pss subscribe fail", "err", err)
	}

	// create and send a pss msg from 'left' to 'right'
	apimsg := &pss.APIMsg{
		Addr: baseaddr_right,
		Msg:  []byte("foobar"),
	}
	err = rpcclient_left.Call(nil, "pss_send", topic, apimsg)
	if err != nil {
		demo.Log.Error("pss send failed", "err", err)
	}

	// receive message on 'right'
	inmsg := <-msgC
	demo.Log.Info("pss received", "msg", string(inmsg.Msg), "from", fmt.Sprintf("%x", inmsg.Addr))

	// bring everything down
	sub.Unsubscribe()
	stack_left.Stop()
	stack_mid.Stop()
	stack_right.Stop()
}
