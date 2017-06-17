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

func main() {

	var err error

	// add the services we want to run
	// we will be serving the protocols specified in foosvc.Protocols()
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
			demo.Log.Debug("Message received!", "payload", foomsg, "code", msg.Code, "size", msg.Size)
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
