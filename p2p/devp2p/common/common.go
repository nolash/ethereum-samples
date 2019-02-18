package common

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm"
	bzzapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/network"
	//	"github.com/ethereum/go-ethereum/swarm/pss"
)

const (
	BzzDefaultNetworkId = 4242
	WSDefaultPort       = 18543
	BzzDefaultPort      = 8542
)

var (
	// custom log, easily grep'able
	Log = log.New("demolog", "*")

	// our working directory
	BasePath string

	// out local port for p2p connections
	P2PPort int

	// self-explanatory command line arguments
	verbose      = flag.Bool("v", false, "more verbose logs")
	remoteport   = flag.Int("p", 0, "remote port (enables remote RPC lookup of enode)")
	remotehost   = flag.String("h", "127.0.0.1", "remote host (RPC, p2p)")
	enodeconnect = flag.String("e", "", "enode to connect to (overrides remote RPC lookup)")
	p2plocalport = flag.Int("l", P2pPort, "local port for p2p connections")
)

// setup logging
// set up remote node, if present
func init() {
	var err error

	flag.Parse()

	// get the working directory
	BasePath, err = os.Getwd()
	if err != nil {
		Log.Crit("Could not determine working directory", "err", err)
	}

	// ensure good log formats for terminal
	// handle verbosity flag
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	loglevel := log.LvlInfo
	if *verbose {
		loglevel = log.LvlTrace
	}
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)
}

func NewSwarmService(stack *node.Node, bzzport int) func(ctx *node.ServiceContext) (node.Service, error) {
	return NewSwarmServiceWithProtocol(stack, bzzport, nil, nil)
}

func NewSwarmServiceWithProtocol(stack *node.Node, bzzport int, specs []*protocols.Spec, protocols []*p2p.Protocol) func(ctx *node.ServiceContext) (node.Service, error) {
	return func(ctx *node.ServiceContext) (node.Service, error) {
		// get the encrypted private key file
		keyid := fmt.Sprintf("%s/D3_Pss/nodekey", stack.DataDir())

		// load the private key from the file content
		prvkey, err := crypto.LoadECDSA(keyid)
		if err != nil {
			return nil, fmt.Errorf("privkey fail: %v", prvkey)
		}

		// create the swarm overlay address
		//chbookaddr := crypto.PubkeyToAddress(prvkey.PublicKey)

		// configure and create a swarm instance
		bzzdir := stack.InstanceDir() // todo: what is the difference between this and datadir?

		bzzconfig := bzzapi.NewConfig()
		bzzconfig.Path = bzzdir
		bzzconfig.Init(prvkey)
		bzzconfig.Port = fmt.Sprintf("%s", bzzport)
		if err != nil {
			Log.Crit("unable to configure swarm", "err", err)
		}
		svc, err := swarm.NewSwarm(bzzconfig, nil)
		if err != nil {
			return nil, err
		}

		//		for i, s := range specs {
		//			_, err := svc.RegisterProtocol(s, protocols[i], &pss.ProtocolParams{true, true})
		//			if err != nil {
		//				return nil, err
		//			}
		//		}
		return svc, nil
	}
}

// set up the local service node
func NewServiceNode(port int, httpport int, wsport int, modules ...string) (*node.Node, error) {
	if port == 0 {
		port = P2pPort
	}
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", port)
	cfg.P2P.EnableMsgEvents = true
	cfg.P2P.NoDiscovery = true
	cfg.IPCPath = IPCName
	cfg.DataDir = fmt.Sprintf("%s%d", DatadirPrefix, port)
	if httpport > 0 {
		cfg.HTTPHost = node.DefaultHTTPHost
		cfg.HTTPPort = httpport
	}
	if wsport > 0 {
		cfg.WSHost = node.DefaultWSHost
		cfg.WSPort = wsport
		cfg.WSOrigins = []string{"*"}
		for i := 0; i < len(modules); i++ {
			cfg.WSModules = append(cfg.WSModules, modules[i])
		}
	}
	stack, err := node.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("ServiceNode create fail: %v", err)
	}
	return stack, nil
}

// create a server
func NewServer(privkey *ecdsa.PrivateKey, name string, version string, proto p2p.Protocol, port int) *p2p.Server {

	cfg := p2p.Config{
		PrivateKey:      privkey,
		Name:            common.MakeName(name, version),
		MaxPeers:        1,
		Protocols:       []p2p.Protocol{proto},
		EnableMsgEvents: true,
	}
	if port > 0 {
		cfg.ListenAddr = fmt.Sprintf(":%d", port)
	}
	srv := &p2p.Server{
		Config: cfg,
	}
	return srv
}

// utility functions
//
// connects to the RPC specified by the url string
// on successful connection it retrieves the enode string from the endpoint
// RPC url can be IPC (filepath), websockets (ws://) or HTTP (http://)
func getEnodeFromRPC(rawurl string) (string, error) {
	rpcclient, err := rpc.Dial(rawurl)
	if err != nil {
		return "", fmt.Errorf("cannot add remote host: %v", err)
	}

	var nodeinfo p2p.NodeInfo
	err = rpcclient.Call(&nodeinfo, "admin_nodeInfo")
	if err != nil {
		return "", fmt.Errorf("RPC nodeinfo call failed: %v", err)
	}
	return nodeinfo.Enode, nil
}

func WaitHealthy(ctx context.Context, minbinsize int, rpcs ...*rpc.Client) error {
	var ids []enode.ID
	var addrs [][]byte
	for _, r := range rpcs {
		var nodeinfo p2p.NodeInfo
		err := r.Call(&nodeinfo, "admin_nodeInfo")
		if err != nil {
			return err
		}
		Log.Debug("nodeinfo", "n", nodeinfo)
		//var id enode.ID
		var p2pnode enode.Node
		//p2pnode, err := enode.UnmarshalText(nodeinfo.Enode)
		//err = id.UnmarshalText([]byte(nodeinfo.Enode))
		err = p2pnode.UnmarshalText([]byte(nodeinfo.Enode))
		if err != nil {
			return err
		}
		//ids = append(ids, id)
		ids = append(ids, p2pnode.ID())
		var bzzaddr string
		err = r.Call(&bzzaddr, "pss_baseAddr")
		if err != nil {
			return err
		}
		addrs = append(addrs, common.FromHex(bzzaddr))
	}
	peerpot := network.NewPeerPotMap(minbinsize, addrs)
	for {
		healthycount := 0
		for i, r := range rpcs {
			var health network.Health
			err := r.Call(&health, "hive_getHealthInfo", peerpot)
			if err != nil {
				return err
			}
			Log.Debug("health", "i", i, "addr", common.ToHex(addrs[i]), "id", ids[i], "info", health)
			if health.KnowNN && health.ConnectNN {
				healthycount++
			}
		}
		if healthycount == len(rpcs) {
			break
		}
		time.Sleep(time.Millisecond * 250)
	}
	return nil
}
