package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
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
	foosvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return newFooService()
	}
	err = stack.Register(foosvc)
	if err != nil {
		demolog.Crit("no service for you", "err", err)
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

	// if we have a connect flag from the invocation
	// connect to the node with the specified enode
	if *remoteenode != "" {
		adminclient, err := stack.Attach() // stack.Server.AddPeer()
		if err != nil {
			demolog.Crit("no rpc for you", "err", err)
		}
		err = adminclient.Call(nil, "admin_addPeer", *remoteenode)
		if err != nil {
			demolog.Crit("no connect for you", "err", err)
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
				go func() {
					for {
						msg, err := rw.ReadMsg()
						if err != nil {
							demolog.Error("readmsg failed", "err", err)
							quitC <- struct{}{}
						} else {
							demolog.Debug("got", "msg", msg)
						}
					}
				}()
				//  sorted alphabetically protoclols on adding
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

// explain caps, rlpx handshake
// ints not possible in shipment
// senditems too
