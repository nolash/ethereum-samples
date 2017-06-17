package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/pot"
	"github.com/ethereum/go-ethereum/rpc"
	pss "github.com/ethereum/go-ethereum/swarm/pss/client"
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

var (
	quitC = make(chan struct{})

	demolog log.Logger

	overlayaddress []byte

	psshost              = flag.String("h", "127.0.0.1", "host of pss websocket rpc")
	pssport              = flag.Int("p", 8546, "port of pss websocket rpc")
	sendtooverlayaddress = flag.String("r", "", "swarm overlay address of recipient to messages")
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
	cfg := pss.NewClientConfig()
	cfg.RemoteHost = *psshost
	cfg.RemotePort = *pssport

	// maybe use rpc.Client instead
	psc, err := pss.NewClient(context.Background(), nil, cfg)
	if err != nil {
		log.Crit("no pss shortcuts for you", "err", err)
	}
	err = psc.Start()
	if err != nil {
		log.Crit("can't start pss client")
		os.Exit(1)
	}

	demolog.Debug("connected to pss node", "bzz addr", psc.BaseAddr)

	foosvc, _ := newFooService()
	fooprotos := foosvc.Protocols()
	err = psc.RunProtocol(&fooprotos[0])
	if err != nil {
		demolog.Crit("can't start protocol on pss websocket")
	}

	if *sendtooverlayaddress != "" {
		byteaddr, err := hex.DecodeString(*sendtooverlayaddress)
		if err != nil {
			demolog.Crit("failed to decode overlay address")
		}
		var potaddr pot.Address
		copy(potaddr[:], byteaddr[:])

		psc.AddPssPeer(potaddr, &fooProtocol)
	}

	<-quitC
	psc.Stop()
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
						serial++
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
