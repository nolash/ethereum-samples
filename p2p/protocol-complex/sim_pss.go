package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	swarmapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/pss"

	colorable "github.com/mattn/go-colorable"

	"./bzz"
	"./protocol"
	"./resource"
	"./service"
)

const (
	defaultMaxDifficulty = 24
	defaultMinDifficulty = 8
	defaultSubmitDelay   = time.Millisecond * 100
	defaultDataSize      = 32
	defaultMaxTime       = time.Second * 15
	defaultSimDuration   = time.Second * 1
	defaultMaxJobs       = 100
	//defaultResourceApiHost = "http://localhost:8500"
)

var (
	loglevel = flag.Bool("v", false, "loglevel")
	//useResource   = flag.Bool("r", false, "use resource sink")
	ensAddr       = flag.String("e", "", "ens name to post resource updates")
	maxDifficulty uint8
	minDifficulty uint8
	maxTime       time.Duration
	maxJobs       int
	privateKeys   map[enode.ID]*ecdsa.PrivateKey
)

func init() {
	flag.Parse()
	if *loglevel {
		log.PrintOrigins(true)
		log.Root().SetHandler(log.LvlFilterHandler(log.LvlDebug, log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
	}

	maxDifficulty = defaultMaxDifficulty
	minDifficulty = defaultMinDifficulty
	maxTime = defaultMaxTime
	maxJobs = defaultMaxJobs

	privateKeys = make(map[enode.ID]*ecdsa.PrivateKey)
	adapters.RegisterServices(newServices())
}

func main() {
	a := adapters.NewSimAdapter(newServices())

	n := simulations.NewNetwork(a, &simulations.NetworkConfig{
		ID:             "protocol-demo",
		DefaultService: "bzz",
	})
	defer n.Shutdown()

	var nids []enode.ID
	for i := 0; i < 5; i++ {
		c := adapters.RandomNodeConfig()
		nod, err := n.NewNodeWithConfig(c)
		if err != nil {
			log.Error(err.Error())
			return
		}
		nids = append(nids, nod.ID())
		privateKeys[nod.ID()], err = crypto.GenerateKey()
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	// TODO: need better assertion for network readiness
	n.StartAll()
	for i, nid := range nids {
		if i == 0 {
			continue
		}
		n.Connect(nids[0], nid)
	}

	go http.ListenAndServe(":8888", simulations.NewServer(n))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err := connectPssPeers(n, nids)
	if err != nil {
		log.Error(err.Error())
		return
	}

	// the fucking healthy stuff
	time.Sleep(time.Second * 1)

	quitC := make(chan struct{})
	trigger := make(chan enode.ID)
	events := make(chan *simulations.Event)
	sub := n.Events().Subscribe(events)
	// event sink on quit
	defer func() {
		sub.Unsubscribe()
		close(quitC)
		select {
		case <-events:
		default:
		}
		return
	}()

	action := func(ctx context.Context) error {
		for i, nid := range nids {
			if i == 0 {
				log.Info("appointed worker node", "node", nid.String())
				go func(nid enode.ID) {
					trigger <- nid
				}(nid)
				continue
			}
			client, err := n.GetNode(nid).Client()
			if err != nil {
				return err
			}

			err = client.Call(nil, "demo_setDifficulty", 0)
			if err != nil {
				return err
			}

			go func(nid enode.ID) {
				timer := time.NewTimer(defaultSimDuration)
				for {
					select {
					case <-events:
						continue
					case <-quitC:
						return
					case <-ctx.Done():
						return
					case <-timer.C:
					}
					log.Debug("stop sending", "node", nid)
					trigger <- nid
					continue
				}
			}(nid)
		}
		return nil
	}
	check := func(ctx context.Context, nid enode.ID) (bool, error) {
		select {
		case <-ctx.Done():
		default:
		}
		log.Warn("sim loop terminated", "nid", nid)
		return true, nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	sim := simulations.NewSimulation(n)
	step := sim.Run(ctx, &simulations.Step{
		Action:  action,
		Trigger: trigger,
		Expect: &simulations.Expectation{
			Nodes: nids,
			Check: check,
		},
	})
	if step.Error != nil {
		log.Error(step.Error.Error())
	}
	var pivotPubKeyHex string
	for i, nid := range nids {
		if i == 0 {
			pubkey := privateKeys[nid].PublicKey
			pubkeybytes := crypto.FromECDSAPub(&pubkey)
			pivotPubKeyHex = common.ToHex(pubkeybytes)
			continue
		}
		log.Debug("stopping node", "nid", nid)
		client, err := n.GetNode(nid).Client()
		if err != nil {
			log.Error("can't get node rpc client", "nid", nid.TerminalString(), "err", err)
			return
		}
		topic := pss.BytesToTopic([]byte(fmt.Sprintf("%s:%d", protocol.Spec.Name, protocol.Spec.Version)))
		err = client.Call(nil, "pss_removePeer", topic, pivotPubKeyHex)
		if err != nil {
			log.Error("can't remove pss protocol peer", "nid", nid.TerminalString(), "err", err)
			return
		}
	}
	for i, nid := range nids {
		if i > 0 {
			n.Stop(nid)
		}
	}
	sigC := make(chan os.Signal)
	signal.Notify(sigC, syscall.SIGINT)

	<-sigC

	return
}

func connectPssPeers(n *simulations.Network, nids []enode.ID) error {
	var pivotBaseAddr string
	var pivotPubKeyHex string
	var pivotClient *rpc.Client
	topic := pss.BytesToTopic([]byte(fmt.Sprintf("%s:%d", protocol.Spec.Name, protocol.Spec.Version)))
	for i, nid := range nids {
		client, err := n.GetNode(nid).Client()
		if err != nil {
			return err
		}
		var baseAddr string
		err = client.Call(&baseAddr, "pss_baseAddr")
		if err != nil {
			return err
		}
		pubkey := privateKeys[nid].PublicKey
		pubkeybytes := crypto.FromECDSAPub(&pubkey)
		pubkeyhex := common.ToHex(pubkeybytes)
		if i == 0 {
			pivotBaseAddr = baseAddr
			pivotPubKeyHex = pubkeyhex
			pivotClient = client
		} else {
			err = client.Call(nil, "pss_setPeerPublicKey", pivotPubKeyHex, common.ToHex(topic[:]), pivotBaseAddr)
			if err != nil {
				return err
			}
			err = pivotClient.Call(nil, "pss_setPeerPublicKey", pubkeyhex, common.ToHex(topic[:]), baseAddr)
			if err != nil {
				return err
			}
			err = client.Call(nil, "pss_addPeer", topic, pivotPubKeyHex, pivotBaseAddr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func newServices() adapters.Services {
	haveWorker := false
	return adapters.Services{
		"bzz": func(node *adapters.ServiceContext) (node.Service, error) {
			//			var resourceEnsName string
			//			if *ensAddr != "" {
			//				resourceEnsName = *ensAddr
			//			} else {
			//				resourceEnsName = fmt.Sprintf("%x.mutable.test", node.Config.ID[:])
			//			}
			//			resourceapi := resource.NewClient(defaultResourceApiHost, resourceEnsName)
			//			var sinkFunc service.ResultSinkFunc
			//			if *useResource {
			//				sinkFunc = resourceapi.ResourceSinkFunc()
			//			}
			params := service.NewDemoParams(sinkFunc, saveFunc)
			params.MaxJobs = maxJobs
			params.MaxTimePerJob = maxTime
			if !haveWorker {
				params.MaxDifficulty = maxDifficulty
				haveWorker = true
			}
			params.SubmitDelay = defaultSubmitDelay
			params.SubmitDataSize = defaultDataSize
			params.MaxSubmitDifficulty = defaultMaxDifficulty
			params.MinSubmitDifficulty = defaultMinDifficulty

			//params.MaxDifficulty = maxDifficulty
			params.Id = node.Config.ID[:]

			// create the pss service that wraps the demo protocol
			svc, err := service.NewDemo(params)
			if err != nil {
				return nil, err
			}
			bzzCfg := swarmapi.NewConfig()
			bzzCfg.SyncEnabled = false
			//bzzCfg.Port = *bzzport
			//bzzCfg.Path = node.ServiceContext.
			bzzCfg.HiveParams.Discovery = true
			bzzCfg.Init(privateKeys[node.Config.ID])

			bzzSvc, err := bzz.NewBzzService(bzzCfg)
			if err != nil {
				return nil, err
			}
			bzzSvc.RegisterPssProtocol(svc)
			return bzzSvc, nil
		},
	}
}

func saveFunc(nid []byte, id protocol.ID, difficulty uint8, data []byte, nonce []byte, hash []byte) {
	fmt.Fprintf(os.Stdout, "RESULT >> %x/%x : %x@%d|%x => %x\n", nid[:8], id, data, difficulty, nonce, hash)
}
