package main

import (
	"net"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/pss"

	demo "./common"
)

func main() {

	// generate a new private key
	privkey, err := crypto.GenerateKey()
	if err != nil {
		demo.Log.Crit("private key generate servicenode 'left' fail", "err", err)
	}

	// constructor configuration for the bzz service bundle
	bzzaddr := network.PrivateKeyToBzzKey(privkey)
	hiveconfig := network.NewHiveParams()
	enod := enode.NewV4(&privkey.PublicKey, net.IPv4(127, 0, 0, 1), 0, 0) // tmp
	bzzconfig := &network.BzzConfig{
		OverlayAddr:  bzzaddr,
		UnderlayAddr: []byte(enod.String()),
		NetworkID:    666,
		HiveParams:   hiveconfig,
	}

	// kademlia object controls the node connection tables
	kadparams := network.NewKadParams()
	kad := network.NewKademlia(
		bzzaddr,
		kadparams,
	)

	// bzz provides connectivity between swarm nodes (handshake)
	bz := network.NewBzz(bzzconfig, kad, nil, nil, nil)

	// set up pss with the same kademlia as the swarm instance
	pp := pss.NewPssParams().WithPrivateKey(privkey)
	ps, err := pss.NewPss(kad, pp)
	if err != nil {
		demo.Log.Crit("PSS create fail", "err", err)
	}

	demo.Log.Info("done", "pss", ps, "bzz", bz)
}
