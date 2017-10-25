// Previous "reply" example using p2p.protocols abstraction
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/chequebook"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/swarm"
	bzzapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/pss"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
	"sync"
)

var (
	messageW  = &sync.WaitGroup{}
	pssprotos []*pss.Protocol
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
	topic = pss.ProtocolTopic(&fooProtocol)
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
			log.Warn("running", "peer", p)
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

func newService(bzzdir string, bzzport int, bzznetworkid uint64, specs []*protocols.Spec, protocols []*p2p.Protocol) func(ctx *node.ServiceContext) (node.Service, error) {
	return func(ctx *node.ServiceContext) (node.Service, error) {
		// generate a new private key
		privkey, err := crypto.GenerateKey()
		if err != nil {
			demo.Log.Crit("private key generate servicenode 'left' fail: %v")
		}

		// create necessary swarm params
		var ensApi chequebook.Backend = nil
		swapEnabled := false
		syncEnabled := false
		pssEnabled := true
		cors := "*"
		checkbookaddr := crypto.PubkeyToAddress(privkey.PublicKey)
		bzzconfig, err := bzzapi.NewConfig(bzzdir, checkbookaddr, privkey, bzznetworkid)
		if err != nil {
			demo.Log.Crit("unable to configure swarm", "err", err)
		}
		bzzconfig.Port = fmt.Sprintf("%d", bzzport)

		svc, err := swarm.NewSwarm(ctx, ensApi, bzzconfig, swapEnabled, syncEnabled, cors, pssEnabled)
		if err != nil {
			return nil, err
		}

		for i, s := range specs {
			p, err := svc.RegisterPssProtocol(s, protocols[i], &pss.ProtocolParams{true, true})
			if err != nil {
				return nil, err
			}
			p.Pss.Register(&topic, p.Handle)
			pssprotos = append(pssprotos, p)
		}
		return svc, nil
	}
}
func main() {

	// create two nodes
	l_stack, err := demo.NewServiceNode(demo.P2PDefaultPort, 0, 0)
	if err != nil {
		demo.Log.Crit(err.Error())
	}
	r_stack, err := demo.NewServiceNode(demo.P2PDefaultPort+1, 0, 0)
	if err != nil {
		demo.Log.Crit(err.Error())
	}

	// register the pss activated bzz services
	l_svc := newService(l_stack.InstanceDir(), demo.BzzDefaultPort, demo.BzzDefaultNetworkId, []*protocols.Spec{&fooProtocol}, []*p2p.Protocol{&proto})
	err = l_stack.Register(l_svc)
	if err != nil {
		demo.Log.Crit("servicenode 'left' pss register fail", "err", err)
	}
	r_svc := newService(r_stack.InstanceDir(), demo.BzzDefaultPort, demo.BzzDefaultNetworkId, []*protocols.Spec{&fooProtocol}, []*p2p.Protocol{&proto})
	err = r_stack.Register(r_svc)
	if err != nil {
		demo.Log.Crit("servicenode 'right' pss register fail", "err", err)
	}

	// start the nodes
	err = l_stack.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}
	err = r_stack.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}

	// connect the nodes
	l_stack.Server().AddPeer(r_stack.Server().Self())

	// get the rpc clients
	l_rpcclient, err := l_stack.Attach()
	r_rpcclient, err := r_stack.Attach()

	// wait until the state of the swarm overlay network is ready
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = demo.WaitHealthy(ctx, 2, l_rpcclient, r_rpcclient)
	if err != nil {
		demo.Log.Crit("health check fail", "err", err)
	}
	time.Sleep(time.Second) // because the healthy does not work

	// get the overlay addresses
	var l_bzzaddr []byte
	err = r_rpcclient.Call(&l_bzzaddr, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}
	var r_bzzaddr []byte
	err = r_rpcclient.Call(&r_bzzaddr, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}

	// get the publickeys
	var l_pubkey []byte
	err = l_rpcclient.Call(&l_pubkey, "pss_getPublicKey")
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}
	var r_pubkey []byte
	err = r_rpcclient.Call(&r_pubkey, "pss_getPublicKey")
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}

	// set the peers' publickeys
	err = l_rpcclient.Call(nil, "pss_setPeerPublicKey", r_pubkey, topic, r_bzzaddr)
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}
	err = r_rpcclient.Call(nil, "pss_setPeerPublicKey", l_pubkey, topic, l_bzzaddr)
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}

	// set up the event subscriptions on both nodes
	eventOneC := make(chan *p2p.PeerEvent)
	sub_one := l_stack.Server().SubscribeEvents(eventOneC)
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
	sub_two := r_stack.Server().SubscribeEvents(eventTwoC)
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

	// needed for sending
	r_pubkeyid := common.ToHex(r_pubkey)

	// addpeer
	nid, _ := discover.HexID("0x00") // this hack is needed to satisfy the p2p method
	p := p2p.NewPeer(nid, fmt.Sprintf("%x", l_bzzaddr), []p2p.Cap{})
	pssprotos[0].AddPeer(p, proto.Run, topic, true, r_pubkeyid)

	// wait for each respective message to be delivered on both sides
	messageW.Wait()

	// terminate subscription loops and unsubscribe
	sub_one.Unsubscribe()
	sub_two.Unsubscribe()
	r_rpcclient.Close()
	l_rpcclient.Close()
	r_stack.Stop()
	l_stack.Stop()
}
