package service

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rpc"

	"../protocol"
)

// TODO: Change the id to sha1(peerid|data|submits.lastid), so moocher can find it in resource updates later
// Demo implements the node.Service interface
type Demo struct {

	// a unique identifier used to track a request across messages
	id []byte

	// worker mode params
	maxJobs       int           // maximum number of simultaneous hashing jobs the node will accept
	currentJobs   int           // how many jobs currently executing
	maxDifficulty uint8         // the maximum difficulty of jobs this node will handle
	maxTimePerJob time.Duration // maximum time one hashing job will run

	// moocher mode params
	workers map[*protocols.Peer]uint8 // an address book of hasher peers for nodes that send requests

	submits *submitStore
	results *resultStore
	save    SaveFunc

	// internal stuff
	protocol *p2p.Protocol
	mu       sync.RWMutex
	ctx      context.Context
	cancel   func()
}

type SaveFunc func(nid []byte, mid protocol.ID, difficulty uint8, data []byte, nonce []byte, hash []byte)

type DemoParams struct {
	Id            []byte
	MaxDifficulty uint8
	MaxJobs       int
	MaxTimePerJob time.Duration
	ResultSink    ResultSinkFunc
	Save          SaveFunc
}

func NewDemoParams(sinkFunc ResultSinkFunc, saveFunc SaveFunc) *DemoParams {
	return &DemoParams{
		ResultSink: sinkFunc,
		Save:       saveFunc,
	}
}

func NewDemo(params *DemoParams) (*Demo, error) {
	ctx, cancel := context.WithCancel(context.Background())
	d := &Demo{
		id:            params.Id,
		maxJobs:       params.MaxJobs,
		maxDifficulty: params.MaxDifficulty,
		maxTimePerJob: params.MaxTimePerJob,
		workers:       make(map[*protocols.Peer]uint8),
		submits:       newSubmitStore(),
		results:       newResultStore(ctx, params.ResultSink),
		save:          params.Save,
		ctx:           ctx,
		cancel:        cancel,
	}
	if err := d.initProtocol(); err != nil {
		return nil, err
	}
	return d, nil
}

func (self *Demo) IsWorker() bool {
	return self.maxDifficulty > 0
}

func (self *Demo) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "demo",
			Version:   "1.0",
			Service:   newDemoAPI(self),
			Public:    true,
		},
	}
}

func (self *Demo) initProtocol() error {
	proto, err := protocol.NewDemoProtocol(self.Run)
	if err != nil {
		return fmt.Errorf("cant't create demo protocol")
	}
	proto.SkillsHandler = self.skillsHandlerLocked
	proto.StatusHandler = self.statusHandlerLocked
	proto.RequestHandler = self.requestHandlerLocked
	proto.ResultHandler = self.resultHandlerLocked
	if err := proto.Init(); err != nil {
		return fmt.Errorf("can't init demo protocol")
	}
	self.protocol = &proto.Protocol
	return nil
}

func (self *Demo) Protocol() *p2p.Protocol {
	return self.protocol
}

func (self *Demo) Spec() *protocols.Spec {
	return protocol.Spec
}

func (self *Demo) Protocols() (protos []p2p.Protocol) {
	return []p2p.Protocol{*self.protocol}
}

func (self *Demo) Start(srv *p2p.Server) error {
	self.results.Start()
	return nil
}

func (self *Demo) Stop() error {
	self.cancel()
	return nil
}

// The protocol code provides Hook to run when protocol starts on a peer
func (self *Demo) Run(p *protocols.Peer) error {
	self.mu.RLock()
	defer self.mu.RUnlock()
	log.Info("run protocol hook", "peer", p, "difficulty", self.maxDifficulty)
	go p.Send(&protocol.Skills{
		Difficulty: self.maxDifficulty,
	})
	return nil
}

func (self *Demo) getNextWorker(difficulty uint8) *protocols.Peer {
	for p, d := range self.workers {
		if d >= difficulty {
			return p
		}
	}
	return nil
}

func (self *Demo) SubmitRequest(data []byte, difficulty uint8) (protocol.ID, error) {
	self.mu.Lock()
	p := self.getNextWorker(difficulty)
	if p == nil {
		return protocol.ID{}, fmt.Errorf("Couldn't find any workers for difficulty %d", difficulty)
	}
	id := newID(data, self.submits.IncSerial())
	self.mu.Unlock()
	go func(id protocol.ID) {
		req := &protocol.Request{
			Id:         id,
			Data:       data,
			Difficulty: difficulty,
		}
		if err := p.Send(req); err == nil {
			if err := self.submits.Put(req, id); err != nil {
				log.Error("submits put fail", "err", err)
			}
		}
	}(id)
	return id, nil
}

func (self *Demo) skillsHandlerLocked(msg *protocol.Skills, p *protocols.Peer) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	log.Trace("have skills type", "msg", msg, "peer", p)
	self.workers[p] = msg.Difficulty
	return nil
}

func (self *Demo) statusHandlerLocked(msg *protocol.Status, p *protocols.Peer) error {
	log.Trace("have status type", "msg", msg, "peer", p)

	self.mu.Lock()
	defer self.mu.Unlock()

	switch msg.Code {
	case protocol.StatusThanksABunch:
		if self.IsWorker() {
			log.Debug("got thanks, how polite!", "msg", msg.Id)
			self.results.Del(msg.Id)
		}
	case protocol.StatusBusy:
		if self.IsWorker() {
			return nil
		}
		log.Debug("peer is busy. please implement throttling")
	case protocol.StatusAreYouKidding:
		if self.IsWorker() {
			return nil
		}
		log.Debug("we sent wrong difficulty or it changed. please implement adjusting it")
	case protocol.StatusGaveup:
		if self.IsWorker() {
			return nil
		}
		log.Debug("peer gave up on the job. please implement how to select someone else for the job")
	}

	return nil
}

func (self *Demo) requestHandlerLocked(msg *protocol.Request, p *protocols.Peer) error {

	self.mu.Lock()
	defer self.mu.Unlock()

	log.Trace("have request type", "msg", msg, "currentjobs", self.currentJobs, "ourdifficulty", self.maxDifficulty, "peer", p)

	if self.currentJobs >= self.maxJobs || self.results.IsFull() {
		go p.Send(&protocol.Status{
			Id:   msg.Id,
			Code: protocol.StatusBusy,
		})
		log.Error("Too busy!")
		return nil
	}

	if self.maxDifficulty < msg.Difficulty {
		go p.Send(&protocol.Status{
			Id:   msg.Id,
			Code: protocol.StatusAreYouKidding,
		})
		return fmt.Errorf("too hard!")
	}
	self.currentJobs++

	go func(msg *protocol.Request) {
		ctx, cancel := context.WithTimeout(self.ctx, self.maxTimePerJob)
		defer cancel()
		j, err := doJob(ctx, msg.Data, msg.Difficulty)

		if err != nil {
			go p.Send(&protocol.Status{
				Id:   msg.Id,
				Code: protocol.StatusGaveup,
			})
			log.Debug("too long!")
			return
		}

		res := &protocol.Result{
			Id:    msg.Id,
			Nonce: j.Nonce,
			Hash:  j.Hash,
		}

		self.results.Put(msg.Id, res)
		self.mu.Lock()
		self.currentJobs--
		self.mu.Unlock()

		go p.Send(res)

		log.Debug("finished job", "id", msg.Id, "nonce", j.Nonce, "hash", j.Hash)
	}(msg)

	return nil
}

func (self *Demo) resultHandlerLocked(msg *protocol.Result, p *protocols.Peer) error {
	self.mu.RLock()
	defer self.mu.RUnlock()
	if self.maxDifficulty > 0 {
		log.Trace("ignored result type", "msg", msg)
	}
	log.Trace("got result type", "msg", msg, "peer", p)

	if !self.submits.Have(msg.Id) {
		log.Debug("stale or fake request id")
		return nil // in case it's stale not fake don't punish the peer
	}
	if !checkJob(msg.Hash, self.submits.GetData(msg.Id), msg.Nonce) {
		return fmt.Errorf("Got incorrect result", "p", p)
	}
	go p.Send(&protocol.Status{
		Id:   msg.Id,
		Code: protocol.StatusThanksABunch,
	})
	self.save(self.id, msg.Id, self.submits.GetDifficulty(msg.Id), self.submits.GetData(msg.Id), msg.Nonce, msg.Hash)
	return nil
}

func newID(data []byte, nonce uint64) (id protocol.ID) {
	c := make([]byte, 8)
	binary.LittleEndian.PutUint64(c, nonce)
	h := sha1.New()
	h.Write(data)
	h.Sum(c)
	copy(id[:], h.Sum(nil)[:8])
	return id
}
