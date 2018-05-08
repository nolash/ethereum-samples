package bzz

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rpc"
	swarmapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/network/stream"
	"github.com/ethereum/go-ethereum/swarm/pss"
	"github.com/ethereum/go-ethereum/swarm/state"
	"github.com/ethereum/go-ethereum/swarm/storage"

	"../service"
)

type SubService interface {
	node.Service
	Spec() *protocols.Spec
	Protocol() *p2p.Protocol
}

type pssDemoService struct {
	SubService
	protocol *pss.Protocol
}

type BzzService struct {
	bzz        *network.Bzz
	lstore     *storage.LocalStore
	ps         *pss.Pss
	pssService map[pss.Topic]*pssDemoService
	//pssProtocol *pss.Protocol
	//Topic       *pss.Topic
	streamer *stream.Registry
	demo     *service.Demo
	rh       *storage.ResourceHandler
}

func NewBzzService(cfg *swarmapi.Config) (*BzzService, error) {
	var err error

	// master parameters
	self := &BzzService{}
	privkey := cfg.ShiftPrivateKey()
	kp := network.NewKadParams()
	to := network.NewKademlia(
		common.FromHex(cfg.BzzKey),
		kp,
	)

	nodeID, err := discover.HexID(cfg.NodeID)
	addr := &network.BzzAddr{
		OAddr: common.FromHex(cfg.BzzKey),
		UAddr: []byte(discover.NewNode(nodeID, net.IP{127, 0, 0, 1}, 30303, 30303).String()),
	}

	// storage
	lstoreparams := storage.NewDefaultLocalStoreParams()
	lstoreparams.Init(filepath.Join(cfg.Path, "chunk"))
	self.lstore, err = storage.NewLocalStore(lstoreparams, nil)
	if err != nil {
		return nil, fmt.Errorf("lstore fail: %v", err)
	}

	// resource handler
	rhparams := &storage.ResourceHandlerParams{
		QueryMaxPeriods: &storage.ResourceLookupParams{
			Limit: false,
		},
		Signer: &storage.GenericResourceSigner{
			PrivKey: privkey,
		},
		EthClient: storage.NewBlockEstimator(),
	}
	self.rh, err = storage.NewResourceHandler(rhparams)
	if err != nil {
		return nil, fmt.Errorf("resource fail: %v", err)
	}
	self.lstore.Validators = []storage.ChunkValidator{self.rh}

	// sync/stream
	stateStore, err := state.NewDBStore(filepath.Join(cfg.Path, "state-store.db"))
	if err != nil {
		return nil, fmt.Errorf("statestore fail: %v", err)
	}
	db := storage.NewDBAPI(self.lstore)
	delivery := stream.NewDelivery(to, db)

	self.streamer = stream.NewRegistry(addr, delivery, db, stateStore, &stream.RegistryOptions{
		DoSync:     false,
		DoRetrieve: true,
	})

	// pss
	pssparams := pss.NewPssParams().WithPrivateKey(privkey)
	self.ps, err = pss.NewPss(to, pssparams)
	if err != nil {
		return nil, err
	}
	self.pssService = make(map[pss.Topic]*pssDemoService)

	// bzz protocol
	bzzconfig := &network.BzzConfig{
		OverlayAddr:  addr.OAddr,
		UnderlayAddr: addr.UAddr,
		HiveParams:   cfg.HiveParams,
	}
	self.bzz = network.NewBzz(bzzconfig, to, stateStore, stream.Spec, self.streamer.Run)

	return self, nil
}

func (self *BzzService) RegisterPssProtocol(psssvc SubService) error {
	spec := psssvc.Spec()
	topic := pss.BytesToTopic([]byte(fmt.Sprintf("%s:%d", spec.Name, spec.Version)))
	psp, err := pss.RegisterProtocol(self.ps, &topic, spec, psssvc.Protocol(), &pss.ProtocolParams{true, true})
	if err != nil {
		return fmt.Errorf("register pss protocol fail: %v", err)
	}
	self.pssService[topic] = &pssDemoService{
		SubService: psssvc,
		protocol:   psp,
	}
	self.ps.Register(&topic, psp.Handle)
	return nil
}

func (self *BzzService) Protocols() (protos []p2p.Protocol) {
	protos = append(protos, self.bzz.Protocols()[0])
	protos = append(protos, self.bzz.Protocols()[1])
	protos = append(protos, self.ps.Protocols()...)
	return
}

func (self *BzzService) APIs() []rpc.API {
	apis := []rpc.API{
		{
			Namespace: "pss",
			Version:   "1.0",
			Service:   newBzzServiceAPI(self),
			Public:    true,
		},
	}
	apis = append(apis, self.bzz.APIs()...)
	apis = append(apis, self.ps.APIs()...)
	for _, a := range self.pssService {
		apis = append(apis, a.APIs()...)
	}
	return apis
}

func (self *BzzService) Start(srv *p2p.Server) error {
	newaddr := self.bzz.UpdateLocalAddr([]byte(srv.Self().String()))
	log.Warn("Updated bzz local addr", "oaddr", fmt.Sprintf("%x", newaddr.OAddr), "uaddr", fmt.Sprintf("%s", newaddr.UAddr))
	err := self.bzz.Start(srv)
	if err != nil {
		return err
	}
	//self.streamer.Start(srv)
	self.ps.Start(srv)
	for _, psssvc := range self.pssService {
		psssvc.Start(srv)
	}
	return nil
}

func (self *BzzService) Stop() error {
	for _, psssvc := range self.pssService {
		psssvc.Stop()
	}
	self.ps.Stop()
	//self.streamer.Stop()
	self.lstore.Close()
	return nil
}

// api to interact with pss protocol
// TODO: change protocol methods so we only have to use pss api here and remove this structure
type BzzServiceAPI struct {
	service *BzzService
	api     *pss.API
}

func newBzzServiceAPI(svc *BzzService) *BzzServiceAPI {
	return &BzzServiceAPI{
		service: svc,
		api:     pss.NewAPI(svc.ps),
	}
}

func (self *BzzServiceAPI) AddPeer(topic pss.Topic, pubKey hexutil.Bytes, addr pss.PssAddress) error {

	psssvc, ok := self.service.pssService[topic]
	if !ok {
		return fmt.Errorf("pss protocol not registered")
	}

	// add the public key to the pss address book
	err := self.api.SetPeerPublicKey(pubKey, topic, addr)
	if err != nil {
		return err
	}

	// register the underlying protocol as pss protocol
	// and start running it on the peer
	var nid discover.NodeID
	copy(nid[:], addr)
	p2pp := p2p.NewPeer(nid, string(pubKey), []p2p.Cap{})
	log.Info(fmt.Sprintf("adding peer %s to demoservice protocol %d, %p %s", pubKey, topic, p2pp, common.ToHex(pubKey)))
	psssvc.protocol.AddPeer(p2pp, topic, true, common.ToHex(pubKey))
	return nil
}
