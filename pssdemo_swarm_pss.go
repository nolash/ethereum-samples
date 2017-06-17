package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm"
	bzzapi "github.com/ethereum/go-ethereum/swarm/api"
	//"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/pss"
	"os"
	"reflect"
	"time"
)

const (
	c_msgcount              = 5
	c_fooprotocolname       = "foo"
	c_fooprotocolversion    = 42
	c_fooprotocolmaxmsgsize = 1024
	c_swarmnetworkid        = 3
)

var (
	quitC = make(chan struct{})

	demolog log.Logger

	overlayaddress []byte

	cheatps *pss.Pss

	wsport               = flag.Int("ps", 8546, "websocket port to listen on")
	localport            = flag.Int("p", 30303, "local port to open")
	remoteenode          = flag.String("c", "", "enode to connect to")
	sendtooverlayaddress = flag.String("r", "", "swarm overlay address of recipient to messages")
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

	// make examples of rpc with ipc, http
	cfg.WSHost = "127.0.0.1"
	cfg.WSPort = *wsport
	cfg.WSModules = []string{"pss"}
	cfg.WSOrigins = []string{"*"}

	stack, err := node.New(cfg)
	if err != nil {
		demolog.Crit("no soup for you", "err", err)
		os.Exit(1)
	}

	// pss is enclosed by swarm
	// the overlay of pss and swram will be the same
	bzzsvc := func(ctx *node.ServiceContext) (node.Service, error) {
		//keyid := "e673972555c363fb58a2e007356a408687374ef992757206b8028e64fe9bf4cf"
		// this is fragile, get the file path in more stable way
		keyid := fmt.Sprintf("%s/pssdemo_swarm_pss/nodekey", cfg.DataDir)
		prvkey, err := crypto.LoadECDSA(keyid)
		if err != nil {
			demolog.Crit("no swarm for you", "keyid", keyid, "err", err)
		}

		chbookaddr := crypto.PubkeyToAddress(prvkey.PublicKey)
		bzzdir := stack.InstanceDir()

		bzzconfig, err := bzzapi.NewConfig(bzzdir, chbookaddr, prvkey, c_swarmnetworkid)
		if err != nil {
			demolog.Crit("unable to configure swarm", "err", err)
		}
		bzzport := "30399"
		if len(bzzport) > 0 {
			bzzconfig.Port = bzzport
		}

		swapEnabled := false
		syncEnabled := false
		pssEnabled := true
		cors := "*"

		return swarm.NewSwarm(ctx, nil, bzzconfig, swapEnabled, syncEnabled, cors, pssEnabled)
	}
	err = stack.Register(bzzsvc)
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

	topic := pss.NewTopic(c_fooprotocolname, c_fooprotocolversion)

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

		// hang on a bit till the host is added in the kademlia table
		time.Sleep(time.Second)

		toaddrbytes, err := hex.DecodeString(*sendtooverlayaddress)
		if err == nil {
			// the pss api can tell us our own overlay address
			// we want to get this so we can tell the other party who the msg came from
			err = adminclient.Call(&overlayaddress, "pss_baseAddr")
			if err != nil {
				demolog.Crit("no overlayaddress for you", "err", err)
			}
			demolog.Info("pss is initialized on a not-so-made-up swarm overlay address", "addr", overlayaddress)

			// use the pss rpc api to send a message
			// it's to ourselves, cos we dont actually have routing now
			// send-to-self enabling is a debug setting enabled in pssconfig
			// ... it's not normally possible

			err = adminclient.Call(nil, "pss_send", topic, pss.APIMsg{
				Msg:  []byte{0x64, 0x6f, 0x64},
				Addr: toaddrbytes,
			})
			if err != nil {
				demolog.Error("no pssmsg for you", "err", err)
			}
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
