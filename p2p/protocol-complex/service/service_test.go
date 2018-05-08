package service

import (
	"bytes"
	"context"
	"crypto/sha1"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/protocols"

	"../protocol"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StderrHandler))
}

type testPeer struct {
	*protocols.Peer
	rw p2p.MsgReader
}

func newPeer(s *protocols.Spec) *testPeer {
	var nid [discover.NodeIDBits / 8]byte
	lrw, rwr := p2p.MsgPipe()
	p := protocols.NewPeer(
		p2p.NewPeer(discover.NodeID(nid), "testpeer", []p2p.Cap{}),
		lrw,
		s,
	)
	return &testPeer{
		Peer: p,
		rw:   rwr,
	}
}

func TestRequestHandler(t *testing.T) {

	// make service and peer
	s := NewDemoService(8, 3, time.Millisecond*500)
	p := newPeer(protocol.Spec)

	// generate data for work
	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err.Error())
	}

	// inject easy request, should complete well within a second
	s.requestHandler(&protocol.Request{
		Data:       data,
		Difficulty: 2,
	}, p.Peer)

	// get the response
	rlpmsg, _ := p.rw.ReadMsg()
	resultmsg := &protocol.Result{}
	if err := rlpmsg.Decode(resultmsg); err != nil {
		t.Fatal(err.Error())
	}

	// inject too high difficulty
	s.requestHandler(&protocol.Request{
		Data:       data,
		Difficulty: 9,
	}, p.Peer)

	// get the response
	rlpmsg, _ = p.rw.ReadMsg()
	statusmsg := &protocol.Status{}
	if err := rlpmsg.Decode(statusmsg); err != nil {
		t.Fatal(err.Error())
	} else if statusmsg.Code != protocol.StatusAreYouKidding {
		t.Fatalf("Expected StatusGaveup (%d), got %d", protocol.StatusAreYouKidding, statusmsg.Code)
	}

	// change the difficulty so we can time out
	s.maxDifficulty = 128

	// start three jobs (maxjobs)
	for i := 0; i < 4; i++ {
		go s.requestHandler(&protocol.Request{
			Data:       data,
			Difficulty: 128,
		}, p.Peer)
	}

	rlpmsg, _ = p.rw.ReadMsg()
	if err := rlpmsg.Decode(statusmsg); err != nil {
		t.Fatal(err.Error())
	} else if statusmsg.Code != protocol.StatusBusy {
		t.Fatalf("Expected StatusBusy (%d), got %d", protocol.StatusBusy, statusmsg.Code)
	}

	rlpmsg, _ = p.rw.ReadMsg()
	if err := rlpmsg.Decode(statusmsg); err != nil {
		t.Fatal(err.Error())
	} else if statusmsg.Code != protocol.StatusGaveup {
		t.Fatalf("Expected StatusGaveup (%d), got %d", protocol.StatusGaveup, statusmsg.Code)
	}

	_ = statusmsg

}

func TestJob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// prepare data arrays
	data := make([]byte, 128)
	c, err := rand.Read(data)
	if err != nil {
		t.Fatal(err.Error())
	} else if c != len(data) {
		t.Fatal("short read")
	}

	// mine
	j, err := doJob(ctx, data, 8)
	if err != nil {
		t.Fatal(err)
	}

	// check
	if !bytes.Equal(j.Data, data) {
		t.Fatalf("data mismatch, expected %x, got %x", j.Data, data)
	}
	checkData := make([]byte, len(data)+8)
	copy(checkData, data)
	copy(checkData[len(data):], j.Nonce)
	h := sha1.New()
	h.Write(checkData)
	result := h.Sum(nil)
	if !bytes.Equal(result, j.Hash) {
		t.Fatalf("hash mismatch, expected %x, got %x (check data %x)", result, j.Hash, checkData)
	}
}
