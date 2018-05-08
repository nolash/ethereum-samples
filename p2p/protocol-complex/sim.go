package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"

	colorable "github.com/mattn/go-colorable"

	"./protocol"
	"./resource"
	"./service"
)

const (
	defaultMaxDifficulty   = 24
	defaultMinDifficulty   = 8
	defaultMaxTime         = time.Second * 10
	defaultMaxJobs         = 100
	defaultResourceApiHost = "http://localhost:8500"
)

var (
	loglevel      = flag.Bool("v", false, "loglevel")
	useResource   = flag.Bool("r", false, "use resource sink")
	ensAddr       = flag.String("e", "", "ens name to post resource update")
	maxDifficulty uint8
	minDifficulty uint8
	maxTime       time.Duration
	maxJobs       int
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

	adapters.RegisterServices(newServices())
}

func main() {
	a := adapters.NewSimAdapter(newServices())
	//	a, err := adapters.NewDockerAdapter()
	//	if err != nil {
	//		log.Crit(err.Error())
	//	}

	n := simulations.NewNetwork(a, &simulations.NetworkConfig{
		ID:             "protocol-demo",
		DefaultService: "demo",
	})
	defer n.Shutdown()

	var nids []discover.NodeID
	for i := 0; i < 5; i++ {
		c := adapters.RandomNodeConfig()
		nod, err := n.NewNodeWithConfig(c)
		if err != nil {
			log.Error(err.Error())
			return
		}
		nids = append(nids, nod.ID())
	}

	go http.ListenAndServe(":8888", simulations.NewServer(n))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := simulations.Up(ctx, n, nids, simulations.UpModeStar); err != nil {
		log.Error(err.Error())
		return
	}

	quitC := make(chan struct{})
	trigger := make(chan discover.NodeID)
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
				go func(nid discover.NodeID) {
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

			go func(nid discover.NodeID) {
				sendlimit := 20
				tick := time.NewTicker(time.Millisecond * 100)
				defer tick.Stop()
				c := 0
				for {
					select {
					case <-events:
						continue
					case <-quitC:
						return
					case <-ctx.Done():
						return
					case <-tick.C:
					}
					if sendlimit < c {
						log.Debug("stop sending", "node", nid)
						trigger <- nid
						tick.Stop()
						continue
					}
					c++
					data := make([]byte, 16)
					rand.Read(data)
					difficulty := rand.Intn(int(maxDifficulty-minDifficulty)) + int(minDifficulty)

					var id protocol.ID
					err := client.Call(&id, "demo_submit", data, difficulty)
					if err != nil {
						log.Warn("job not accepted", "err", err)
					} else {
						log.Info("job submitted", "id", id)
					}
				}
			}(nid)
		}
		return nil
	}
	check := func(ctx context.Context, nid discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
		default:
		}
		log.Warn("ok", "nid", nid)
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
	for i, nid := range nids {
		if i == 0 {
			continue
		}
		log.Debug("stopping node", "nid", nid)
		n.Stop(nid)

	}
	sigC := make(chan os.Signal)
	signal.Notify(sigC, syscall.SIGINT)

	<-sigC

	return
}

func newServices() adapters.Services {
	return adapters.Services{
		"demo": func(node *adapters.ServiceContext) (node.Service, error) {
			var resourceEnsName string
			if *ensAddr != "" {
				resourceEnsName = *ensAddr
			} else {
				resourceEnsName = fmt.Sprintf("%x.mutable.test", node.Config.ID[:])
			}
			resourceapi := resource.NewClient(defaultResourceApiHost, resourceEnsName)
			var sinkFunc service.ResultSinkFunc
			if *useResource {
				sinkFunc = resourceapi.ResourceSinkFunc()
			}
			params := service.NewDemoParams(sinkFunc, saveFunc)
			params.MaxJobs = maxJobs
			params.MaxTimePerJob = maxTime
			params.MaxDifficulty = maxDifficulty

			params.Id = node.Config.ID[:]
			return service.NewDemo(params)
		},
	}
}

func saveFunc(nid []byte, id protocol.ID, difficulty uint8, data []byte, nonce []byte, hash []byte) {
	fmt.Fprintf(os.Stdout, "RESULT >> %x/%x : %x@%d|%x => %x\n", nid[:8], id, data, difficulty, nonce, hash)
}
