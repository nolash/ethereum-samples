// pss RPC routed over swarm
package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/pot"
	pss "github.com/ethereum/go-ethereum/swarm/pss/client"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
	"time"
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

func main() {
	// pssclient uses context for session control
	ctx, cancel := context.WithCancel(context.Background())
	_, _, stopfunc, err := demo.NewPssPool()
	if err != nil {
		demo.Log.Crit("create pss pool fail: %v", err)
	}

	// give the kademlia overlay a bit of time to get ready
	time.Sleep(time.Millisecond * 250)

	// configure and start up pss client RPCs
	cfg_left := pss.NewClientConfig()
	psclient_left, err := pss.NewClient(ctx, cancel, cfg_left)
	if err != nil {
		demo.Log.Crit("pssclient 'left' create fail", "err", err)
	}
	cfg_right := pss.NewClientConfig()
	cfg_right.RemotePort += 2
	psclient_right, err := pss.NewClient(ctx, cancel, cfg_right)
	if err != nil {
		demo.Log.Crit("pssclient 'right' create fail", "err", err)
	}

	// connect to the RPCs
	err = psclient_left.Start()
	if err != nil {
		demo.Log.Crit("pssclient 'left' start fail", "err", err)
	}
	psclient_right.Start()
	if err != nil {
		demo.Log.Crit("pssclient 'right' start fail", "err", err)
	}

	// run the pssclient protocols
	// this registers the protocol handler in pss with topic generated from the protocol name and version
	psclient_left.RunProtocol(&proto)
	psclient_right.RunProtocol(&proto)

	// add the 'right' peer
	// this will automatically make the 'left' ping the 'right'
	var potaddr pot.Address
	copy(potaddr[:], psclient_right.BaseAddr)
	psclient_left.AddPssPeer(potaddr, &demo.FooProtocol)

	// the 'right' will receive the ping and send on the quit channel
	<-quitC
	stopfunc()
}
