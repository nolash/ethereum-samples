// A simple bidirectional exchange of messages with incremental serial numbers
package main

import (
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
	"os"
	"time"
)

const (
	msgCount = 5
)

var (
	quitC = make(chan bool)
)

func init() {
	err := demo.SetupNode()
	if err != nil {
		demo.Log.Crit("Couldn't set up node", "err", err)
	}
	if demo.RemoteNode == nil {
		demo.Log.Warn("Could not resolve a remote node. This node will be listening until someone connects")
	}
}

func main() {

	var err error

	// add the services we want to run
	// we will be serving the protocols specified in foosvc.Protocols()
	// the run-argument passed to the foo-service constructor is the function that handles the protocol logic. See below.
	foosvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return demo.NewFooService(run)
	}
	err = demo.ServiceNode.Register(foosvc)
	if err != nil {
		demo.Log.Crit("Register service in ServiceNode failed", "err", err)
		os.Exit(1)
	}

	// fire up the node
	// this starts the tcp/rlpx server and activate protocols on it
	err = demo.ServiceNode.Start()
	if err != nil {
		demo.Log.Crit("ServiceNode start failed", "err", err)
	}

	if demo.RemoteNode != nil {
		demo.ServiceNode.Server().AddPeer(demo.RemoteNode)
	}

	ok := <-quitC
	if !ok {
		demo.Log.Crit("Oh-ooh, spaghettios")
	}
	demo.Log.Info("That's all, folks")
}

// this is this example's protocol implementation
// it sends 5 "FooMsg" with incrementing serial numbers to its peer
// upon receiving a message it attempts to decode it as a FooMsg
func run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	var serial uint = 0
	go func() {
		for {
			// listen for incoming messages
			msg, err := rw.ReadMsg()
			if err != nil {
				demo.Log.Error("Protocol readmsg failed", "err", err)
				quitC <- false
				return
			}

			// we expect a "FooMsg" type
			// so we try to infer it into a FooMsg pointer
			foomsg := &demo.FooMsg{}
			err = msg.Decode(foomsg)

			if err != nil {
				demo.Log.Error("Unexpected message type", "err", err)
				quitC <- false
				return
			}
			demo.FooHandler(foomsg)
		}
	}()
	for serial < msgCount {
		err := p2p.Send(rw, 0, &demo.FooMsg{Serial: serial})
		if err != nil {
			return err
		}
		serial++
		time.Sleep(time.Second)
	}

	quitC <- true
	return nil
}
