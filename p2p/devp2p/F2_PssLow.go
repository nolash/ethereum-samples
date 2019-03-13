package main

import (
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/pss"

	demo "./common"
)

// SwarmAndPss is a wrapper for the pss and bzz service combo
// Makes necessary objects for pss comms available to caller
type SwarmAndPss struct {
	pss       *pss.Pss
	bzz       *network.Bzz
	networkid uint64
	node      *node.Node
	host      net.TCPAddr
}

// ServiceConstructor implements node.ServiceConstructor
// it is passed to a service node and combines the bzz and pss services
func (s *SwarmAndPss) ServiceConstructor() node.ServiceConstructor {
	return func(ctx *node.ServiceContext) (node.Service, error) {

		// generate a new private key
		privkey, err := crypto.GenerateKey()
		if err != nil {
			demo.Log.Crit("private key generate servicenode 'left' fail: %v")
		}

		// constructor configuration for the bzz service bundle
		bzzaddr := network.PrivateKeyToBzzKey(privkey)
		hiveconfig := network.NewHiveParams()
		enod := enode.NewV4(&privkey.PublicKey, s.host.IP, s.host.Port, s.host.Port) // tmp
		bzzconfig := &network.BzzConfig{
			OverlayAddr:  bzzaddr,
			UnderlayAddr: []byte(enod.String()),
			NetworkID:    s.networkid,
			HiveParams:   hiveconfig,
		}

		// kademlia object controls the node connection tables
		kadconfig := network.NewKadParams()
		kad := network.NewKademlia(
			bzzaddr,
			kadconfig,
		)
		// bzz provides connectivity between swarm nodes (handshake)
		s.bzz = network.NewBzz(bzzconfig, kad, nil, nil, nil)

		// set up pss with the same kademlia as the swarm instance
		pssconfig := pss.NewPssParams().WithPrivateKey(privkey)
		s.pss, err = pss.NewPss(kad, pssconfig)
		if err != nil {
			return nil, fmt.Errorf("PSS create fail: %v", err)
		}

		return s, nil
	}
}

// Pss returns the pss.Pss instance running on the node
func (s *SwarmAndPss) Pss() *pss.Pss {
	return s.pss
}

// Node returns the node.Node instance running the services
func (s *SwarmAndPss) Node() *node.Node {
	return s.node
}

// Protocols implements node.Service
func (s *SwarmAndPss) Protocols() []p2p.Protocol {
	p := s.bzz.Protocols()
	p = append(p, s.pss.Protocols()...)
	return p
}

// APIs implements node.Service
func (s *SwarmAndPss) APIs() []rpc.API {
	return []rpc.API{}
}

// Start implements node.Service
func (s *SwarmAndPss) Start(srv *p2p.Server) error {
	s.bzz.Start(srv)
	s.pss.Start(srv)
	return nil
}

// Stop implements node.Service
func (s *SwarmAndPss) Stop() error {
	return nil
}

// creates a minimal service node (without http and ws)
func newServiceController(datadir string, port int, bzznetworkid uint64) (*SwarmAndPss, error) {
	var err error

	svcWrapper := &SwarmAndPss{
		networkid: bzznetworkid,
	}
	svcWrapper.host = net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: port,
	}
	nodconfig := &node.DefaultConfig
	nodconfig.P2P.ListenAddr = fmt.Sprintf("%v:%d", svcWrapper.host.IP, svcWrapper.host.Port)
	nodconfig.P2P.NoDiscovery = true
	nodconfig.IPCPath = demo.IPCName
	nodconfig.DataDir = fmt.Sprintf("%s%d", datadir, svcWrapper.host.Port)
	svcWrapper.node, err = node.New(nodconfig)
	if err != nil {
		return nil, fmt.Errorf("ServiceNode create fail: %v", err)
	}
	err = svcWrapper.node.Register(svcWrapper.ServiceConstructor())
	if err != nil {
		return nil, fmt.Errorf("ServiceNode register service fail: %v", err)
	}

	return svcWrapper, nil
}

// a wrapper for notifying main thread of the received message
type pssMsgNotification struct {
	keyid string
	msg   []byte
}

// object providing the handler function for message in pss
// includes a notification channel for received messages
type pssMsgHandler struct {
	notifyC chan pssMsgNotification
}

// Implements pss.HandlerFunc
func (h *pssMsgHandler) handler(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
	demo.Log.Debug("Received msg", "msg", msg, "keyid", keyid)
	h.notifyC <- pssMsgNotification{
		keyid: keyid,
		msg:   msg,
	}
	return nil
}

func main() {
	var err error

	// create two nodes and start them
	bundle_l, err := newServiceController(".data", 30399, 666)
	if err != nil {
		demo.Log.Crit("Service create fail", "err", err)
	}
	node_l := bundle_l.Node()
	err = node_l.Start()
	if err != nil {
		demo.Log.Crit("ServiceNode start fail", "err", err)
	}
	defer node_l.Stop()

	bundle_r, err := newServiceController(".data", 30340, 666)
	if err != nil {
		demo.Log.Crit("Service create fail", "err", err)
	}
	node_r := bundle_r.Node()
	err = node_r.Start()
	if err != nil {
		demo.Log.Crit("ServiceNode start fail", "err", err)
	}
	defer node_r.Stop()

	// get the connection handler service instance (p2p.Server)
	//  and initiate connection
	srv_l := node_l.Server()
	enod_r := node_r.Server().Self()
	srv_l.AddPeer(enod_r)

	// create a pss message handler in the second node
	// the received message will appear on the channel
	topic := pss.BytesToTopic([]byte("foo"))
	notifyC := make(chan pssMsgNotification)
	handler_r := pssMsgHandler{
		notifyC: notifyC,
	}

	// register the handler in the pss instance
	h := pss.NewHandler(handler_r.handler)
	pss_r := bundle_r.Pss()
	pss_r.Register(&topic, h)

	// add the second node to the address book of the first node
	pss_l := bundle_l.Pss()
	pss_l.SetPeerPublicKey(pss_r.PublicKey(), topic, pss_r.BaseAddr())

	// send the message using the address book entry
	pss_l.SendAsym(hexutil.Encode(crypto.FromECDSAPub(pss_r.PublicKey())), topic, []byte("hey"))

	// that's all folks
	demo.Log.Info("done", "notification", <-notifyC)
}
