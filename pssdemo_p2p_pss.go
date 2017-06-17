package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/pss"
	"os"
	"time"
)

const (
	c_msgcount              = 5
	c_fooprotocolname       = "foo"
	c_fooprotocolversion    = 42
	c_fooprotocolmaxmsgsize = 1024
)

var (
	quitC = make(chan struct{})

	demolog log.Logger

	overlayaddress = network.RandomAddr().Over()

	cheatps *pss.Pss

	localport   = flag.Int("p", 30303, "local port to open")
	remoteenode = flag.String("c", "", "enode to connect to")
)

func init() {
	flag.Parse()

	demolog = log.New("demolog", "*")
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	hf := log.LvlFilterHandler(log.LvlTrace, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)
}

func main() {

	var err error

	// configure the node
	// we want the datadir in current dir
	ourdir, err := os.Getwd()
	if err != nil {
		panic("getwd")
	}

	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", *localport)
	cfg.IPCPath = "food.ipc"
	cfg.DataDir = fmt.Sprintf("%s/.data_%d", ourdir, *localport)
	stack, err := node.New(cfg)
	if err != nil {
		demolog.Crit("no soup for you", "err", err)
		os.Exit(1)
	}

	// add the services we want to run
	// we will be serving the protocols specified in foosvc.Protocols()
	psssvc := func(ctx *node.ServiceContext) (node.Service, error) {
		overlayparams := network.NewKadParams()
		overlay := network.NewKademlia(overlayaddress, overlayparams)
		psparams := pss.NewPssParams(true)
		ps := pss.NewPss(overlay, nil, psparams)
		if ps == nil {
			return nil, fmt.Errorf("pss new fail: %v", err)
		}
		cheatps = ps
		return ps, nil
	}
	err = stack.Register(psssvc)
	if err != nil {
		demolog.Crit("no pss-service for you", "err", err)
		os.Exit(1)
	}

	// fire up the node
	// this starts the tcp/rlpx server and activate protocols on it
	err = stack.Start()
	if err != nil {
		demolog.Crit("no stack for you", "err", err)
	}
	nodeinfo := stack.Server().NodeInfo()
	demolog.Info("soup for you after all :)", "id", nodeinfo.ID, "enode", nodeinfo.Enode, "ip", nodeinfo.IP)

	// this is how we should retrieve the pss service, in case we need to use it for something
	// we are using it wrong apparently, so we've cheated above
	//	ps := &pss.Pss{}
	//	err = stack.Service(ps)
	//	if err != nil {
	//		demolog.Crit("we shouldnt be here", "err", err)
	//	}

	psaddr := cheatps.BaseAddr()
	demolog.Info("pss is initialized on the totally made-up swarm overlay address", "addr", fmt.Sprintf("%x", psaddr))

	//fooproto := fooService{}.Protocols()[0]
	//fooprototopic := cheatps.NewTopic(fooproto.Name, fooproto.Version)
	topic := pss.NewTopic("foo", 42)
	cheatps.Register(&topic, func(msg []byte, p *p2p.Peer, from []byte) error {
		demolog.Debug("psshandler", "msg", msg, "peer", p, "from", from, "topic", topic)
		return nil
	})

	// if we have a connect flag from the invocation
	// connect to the node with the specified enode
	if *remoteenode != "" {
		adminclient, err := stack.Attach()
		if err != nil {
			demolog.Crit("no rpc for you", "err", err)
		}
		err = adminclient.Call(nil, "admin_addPeer", *remoteenode)
		if err != nil {
			demolog.Crit("no connect for you", "err", err)
		}

		// use the pss rpc api to send a message
		// it's to ourselves, cos we dont actually have routing now
		// send-to-self enabling is a debug setting enabled in pssconfig
		// ... it's not normally possible
		err = adminclient.Call(nil, "pss_send", topic, pss.APIMsg{
			Msg:  []byte{0x64, 0x6f, 0x64},
			Addr: psaddr,
		})
		if err != nil {
			demolog.Error("no bytes thru pssmsg for you", "err", err)
		}

		protomsg, err := pss.NewProtocolMsg(0, &fooMsg{
			Serial: 1,
		})
		err = adminclient.Call(nil, "pss_send", topic, pss.APIMsg{
			Msg:  protomsg,
			Addr: psaddr,
		})
		if err != nil {
			demolog.Error("no foomsg thru pssmsg for you", "err", err)
		}
	}

	<-quitC
	demolog.Info("that's all folks")

}

// we need to abstract this to get access to the p2p.MsgReadWriter
// it is not exported in p2p.Peer
type fooPeer struct {
	Peer *p2p.Peer
	rw   p2p.MsgReadWriter
}

// this is the service we want to run
type fooService struct {
}

func newFooService() (fooService, error) {
	return fooService{}, nil
}

func (self fooService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		p2p.Protocol{
			Name:    c_fooprotocolname,
			Version: c_fooprotocolversion,
			Length:  1,
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				var serial uint = 0
				fp := fooPeer{
					Peer: p,
					rw:   rw,
				}
				fmt.Printf("%v", fp)
				go func() {
					for {
						msg, err := fp.rw.ReadMsg()
						if err != nil {
							demolog.Error("readmsg failed", "err", err)
							quitC <- struct{}{}
						} else {
							demolog.Debug("got", "msg", msg)
						}
					}
				}()
				for serial < c_msgcount {
					err := p2p.Send(rw, 0, &fooMsg{Serial: serial})
					if err != nil {
						return err
					}
					serial++
					time.Sleep(time.Second)
				}
				return nil
			},
		},
	}
}

// this is the structure of the protocol message we want to send around
type fooMsg struct {
	Serial uint
}

// we really only care about the protocol part
// everything else is minimal
func (self fooService) APIs() []rpc.API {
	return []rpc.API{}
}

func (self fooService) Start(srv *p2p.Server) error {
	return nil
}

func (self fooService) Stop() error {
	return nil
}
