package common

import (
	"github.com/ethereum/go-ethereum/p2p/protocols"
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
