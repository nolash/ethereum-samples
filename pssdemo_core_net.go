// +build pssdemocorenet
package main

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/pss"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"io/ioutil"
	"os"
)

var (
	righttopic = pss.NewTopic("foo", 4)
	wrongtopic = pss.NewTopic("bar", 2)
)

func init() {
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	hf := log.LvlFilterHandler(log.LvlTrace, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)
}

// Pss.Handler type
//
// this is a handler for a particular PssMsg topic
// there can be many for the same topic
func handler(msg []byte, p *p2p.Peer, from []byte) error {
	log.Debug("received", "msg", msg, "from", from, "forwarder", p.ID())
	return nil
}

func main() {

	// bogus addresses for illustration purposes
	meaddr := network.RandomAddr()
	toaddr := network.RandomAddr()
	fwdaddr := network.RandomAddr()

	// new kademlia for routing
	kp := network.NewKadParams()
	to := network.NewKademlia(meaddr.Over(), kp)

	// new (local) storage for cache
	cachedir, err := ioutil.TempDir("", "pss-cache")
	if err != nil {
		panic("overlay")
	}
	dpa, err := storage.NewLocalDPA(cachedir)
	if err != nil {
		panic("storage")
	}

	// setup pss
	psp := pss.NewPssParams(false)
	ps := pss.NewPss(to, dpa, psp)

	// does nothing but please include it
	ps.Start(nil)

	dereg := ps.Register(&righttopic, handler)

	// in its simplest form a message is just a byteslice
	payload := []byte("foobar")

	// send a raw message
	err = ps.SendRaw(toaddr.Over(), righttopic, payload)
	log.Error("Fails. Not connect, so nothing in kademlia. But it illustrates the point.", "err", err)

	// forward a full message
	envfwd := pss.NewEnvelope(fwdaddr.Over(), righttopic, payload)
	msgfwd := &pss.PssMsg{
		To:      toaddr.Over(),
		Payload: envfwd,
	}
	err = ps.Forward(msgfwd)
	log.Error("Also fails, same reason. I wish, I wish, I wish there was somebody out there.", "err", err)

	// process an incoming message
	// (this is the first step after the devp2p PssMsg message handler)
	envme := pss.NewEnvelope(toaddr.Over(), righttopic, payload)
	msgme := &pss.PssMsg{
		To:      meaddr.Over(),
		Payload: envme,
	}
	err = ps.Process(msgme)
	if err == nil {
		log.Info("this works :)")
	}

	// if we don't have a registered topic it fails
	dereg() // remove the previously registered topic-handler link
	ps.Process(msgme)
	log.Error("It fails as we expected", "err", err)

	// does nothing but please include it
	ps.Stop()
}
