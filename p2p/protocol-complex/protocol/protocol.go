package protocol

import (
	"errors"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
)

// enumeration used in status messages
const (
	StatusThanksABunch = iota
	StatusBusy
	StatusAreYouKidding
	StatusGaveup
)

// which hashes a hasher node offers
const (
	HashSHA1 = iota
)

// variables shared between p2p.Protocol and protocols.Spec
const (
	protoName    = "demo"
	protoVersion = 1
	protoMax     = 2048
)

type ID [8]byte

// Skills is a protocol message type
//
// It is an asynchronous handshake message, signaling the state the node is in.
// The message may be issued at any time, and receiving peers should behave accordingly
// towards the node:
//
// Difficulty > 0 means it's open for hashing, and what the max difficulty is.
//
// MaxSize tells how many bytes can accompany one data submission
type Skills struct {
	Difficulty uint8
	MaxSize    uint16
}

// Status is a protocol message type
//
// It is mostly used to signal errors states.
//
// A node also uses Status to signal successful reception of a Result message
type Status struct {
	Id   ID
	Code uint8
}

// Request is a protocol message type
//
// It is used by nodes to request a hashing job
type Request struct {
	Id         ID
	Data       []byte
	Difficulty uint8
}

// Result is a protocol message type
//
// It is used by nodes to transmit the results of a hashing job
type Result struct {
	Id    ID
	Nonce []byte
	Hash  []byte
}

var (
	Messages = []interface{}{
		&Skills{},
		&Status{},
		&Request{},
		&Result{},
	}

	Spec = &protocols.Spec{
		Name:       protoName,
		Version:    protoVersion,
		MaxMsgSize: protoMax,
		Messages:   Messages,
	}
)

// The protocol object wraps the code that starts a protocol on a peer upon connection
//
// This implementation holds a callback function thats called upon a successful connection
// Any logic needed to be performed in the context of the protocol's service should be put there
type DemoProtocol struct {
	Protocol       p2p.Protocol
	SkillsHandler  func(*Skills, *protocols.Peer) error
	StatusHandler  func(*Status, *protocols.Peer) error
	RequestHandler func(*Request, *protocols.Peer) error
	ResultHandler  func(*Result, *protocols.Peer) error
	handler        func(interface{}) error
	runHook        func(*protocols.Peer) error
}

func NewDemoProtocol(runHook func(*protocols.Peer) error) (*DemoProtocol, error) {
	proto := &DemoProtocol{
		Protocol: p2p.Protocol{
			Name:    protoName,
			Version: protoVersion,
			Length:  4,
		},
		runHook: runHook,
	}

	return proto, nil
}

// TODO: double-check if we need the Init detached
func (self *DemoProtocol) Init() error {
	if self.SkillsHandler == nil {
		return errors.New("missing skills handler")
	}
	if self.StatusHandler == nil {
		return errors.New("missing status handler")
	}
	if self.RequestHandler == nil {
		return errors.New("missing request handler")
	}
	if self.ResultHandler == nil {
		return errors.New("missing response handler")
	}
	self.Protocol.Run = self.Run
	return nil
}

// This method is run on every new peer connection
//
// It enters a loop that takes care of dispatching and receiving messages
func (self *DemoProtocol) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	pp := protocols.NewPeer(p, rw, Spec)
	log.Info("running demo protocol on peer", "peer", pp, "self", self)
	go self.runHook(pp)
	dp := &DemoPeer{
		Peer:           pp,
		skillsHandler:  self.SkillsHandler,
		statusHandler:  self.StatusHandler,
		requestHandler: self.RequestHandler,
		resultHandler:  self.ResultHandler,
	}
	return pp.Run(dp.Handle)
}
