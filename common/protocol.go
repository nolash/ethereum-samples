package common

import (
	"github.com/ethereum/go-ethereum/p2p/protocols"
)

const (
	FooProtocolName       = "foo"
	FooProtocolVersion    = 42
	FooProtocolMaxMsgSize = 1024
)

// needed to use the p2p/protocols abstractions
var (
	FooProtocol = protocols.Spec{
		Name:       FooProtocolName,
		Version:    FooProtocolVersion,
		MaxMsgSize: FooProtocolMaxMsgSize,
		Messages:   fooprotocolmessages,
	}
)
