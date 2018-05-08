package service

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/pss"
)

type DemoService struct {
	enodes   []string
	protocol *pss.Protocol
}

// implement node.Service
func (self *DemoService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "demo",
			Version: 1,
			Length:  1,
			Run:     self.Run,
		},
	}
}

func (self *DemoService) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "demo",
			Version:   "1.0",
			Service:   newDemoServiceAPI(self.protocol, self.Run),
			Public:    true,
		},
	}
}

func (self *DemoService) Start(p *p2p.Server) error {
	return nil
}

func (self *DemoService) Stop() error {
	return nil
}

// Implement rest of PssService
func (self *DemoService) Spec() *protocols.Spec {
	return &protocols.Spec{
		Name:       protoName,
		Version:    protoVersion,
		MaxMsgSize: protoMax,
		Messages: []interface{}{
			&Result{},
			&Status{},
			&Job{},
		},
	}
}

func (self *DemoService) Topic() *pss.Topic {
	return &protoTopic
}

func (self *DemoService) Init(ps *pss.Pss) error {
	protocol := self.Protocols()[0]
	psp, err := pss.RegisterProtocol(ps, self.Topic(), self.Spec(), &protocol, &pss.ProtocolParams{true, true})
	if err != nil {
		return err
	}
	ps.Register(self.Topic(), psp.Handle)
	self.protocol = psp
	return nil
}

func (self *DemoService) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	pp := protocols.NewPeer(p, rw, self.Spec())
	go func() {
		err := pp.Send(&Status{true})
		log.Error("send fail", "peer", pp, "err", err)
	}()

	err := pp.Run(self.handle)
	return err
}

func (self *DemoService) handle(msg interface{}) error {
	log.Info("have msg", "msg", msg)
	return nil
}

// api to interact with pss protocol
type DemoServiceAPI struct {
	protocol *pss.Protocol
	run      func(*p2p.Peer, p2p.MsgReadWriter) error
}

func newDemoServiceAPI(prot *pss.Protocol, run func(*p2p.Peer, p2p.MsgReadWriter) error) *DemoServiceAPI {
	return &DemoServiceAPI{
		protocol: prot,
		run:      run,
	}
}

func (self *DemoServiceAPI) AddPssPeer(key string) error {
	self.protocol.AddPeer(&p2p.Peer{}, self.run, protoTopic, true, key)
	log.Info(fmt.Sprintf("adding peer %x to demoservice protocol", key))
	return nil
}
