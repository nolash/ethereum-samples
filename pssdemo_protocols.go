package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rpc"
	"os"
	"reflect"
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

// using p2p.protocols abstraction we can register the message structs we use for our protocol more conveniently
// it enables us to use an external handler function for incoming messages
// only messages of the types given in the "Messages" member will make it through
var (
	fooProtocol = protocols.Spec{
		Name:       c_fooprotocolname,
		Version:    c_fooprotocolversion,
		MaxMsgSize: c_fooprotocolmaxmsgsize,
		Messages: []interface{}{
			&fooMsg{}, &fooOtherMsg{}, &fooUnhandledMsg{},
		},
	}
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
		demolog.Crit("no fooservice for you", "err", err)
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
		adminclient, err := stack.Attach()
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
			Name:    fooProtocol.Name,
			Version: fooProtocol.Version,
			Length:  uint64(len(fooProtocol.Messages)),
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				pp := protocols.NewPeer(p, rw, &fooProtocol)
				go func() {
					var serial uint = 0
					var err error
					for serial < c_msgcount {
						time.Sleep(time.Second)
						err = pp.Send(&fooMsg{Serial: serial})
						if err != nil {
							demolog.Error("can't send to peer", "peer", pp, "err", err)
							quitC <- struct{}{}
						}
						err = pp.Send(&fooOtherMsg{Created: time.Now()})
						if err != nil {
							demolog.Error("can't send to peer", "peer", pp, "err", err)
							quitC <- struct{}{}
						}

						serial++
					}
					err = pp.Send(&fooUnhandledMsg{Content: []byte{0x64, 0x6f, 0x6f}})
					if err != nil {
						demolog.Debug("can't send to peer", "err", err)
						quitC <- struct{}{}
					}

				}()
				pp.Run(fooHandler)
				quitC <- struct{}{}
				return nil
			},
		},
	}
}

// this is the handler of the incoming message
func fooHandler(msg interface{}) error {
	foomsg, ok := msg.(*fooMsg)
	if ok {
		demolog.Debug("foomsg!", "msg", foomsg)
		return nil
	}

	fooothermsg, ok := msg.(*fooOtherMsg)
	if ok {
		demolog.Debug("fooothermsg!", "msg", fooothermsg)
		return nil
	}
	return fmt.Errorf("If you reach here, you forgot to make a handler for the message type %v", reflect.TypeOf(msg))
}

// we have two types of messages we want to send
type fooMsg struct {
	Serial uint
}

type fooOtherMsg struct {
	Created time.Time
}

type fooUnhandledMsg struct {
	Content []byte
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
