package common

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm"
	bzzapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/network"
)

const (
	P2PDefaultPort = 30100
	IPCName        = "demo.ipc"
	bzzNetworkId   = 3
)

var (
	// custom log, easily grep'able
	Log = log.New("demolog", "*")

	// our working directory
	BasePath string

	// out local port for p2p connections
	P2PPort int

	// predictable datadir name prefix
	DatadirPrefix = ".data_"

	// self-explanatory command line arguments
	verbose      = flag.Bool("v", false, "more verbose logs")
	remoteport   = flag.Int("p", 0, "remote port (enables remote RPC lookup of enode)")
	remotehost   = flag.String("h", "127.0.0.1", "remote host (RPC, p2p)")
	enode        = flag.String("e", "", "enode to connect to (overrides remote RPC lookup)")
	p2plocalport = flag.Int("l", P2PDefaultPort, "local port for p2p connections")
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

// set up a line of three pss-enabled hosts offering websocket interface
//func NewPssPool() (ipc_left *rpc.Client, ipc_right *rpc.Client, stopfunc func(), err error) {
func NewPssPool() (port_left int, port_right int, stopfunc func(), err error) {

	var stacks []*node.Node
	for i := 0; i < 3; i++ {
		stack, err := NewServiceNode(P2PDefaultPort+i, 0, node.DefaultWSPort+i, "pss")
		if err != nil {
			return 0, 0, nil, fmt.Errorf("bzz service #%d create fail: %v", i+1, err)
		}
		err = stack.Register(NewSwarmService(stack, 30399+i))
		if err != nil {
			return 0, 0, nil, fmt.Errorf("bzz service #%d register fail: %v", i+1, err)
		}
		err = stack.Start()
		if err != nil {
			return 0, 0, nil, fmt.Errorf("servicenode #%d start failed: %v", i+1, err)
		}
		stacks = append(stacks, stack)
	}

	// connect the nodes
	p2pnode_mid := stacks[1].Server().Self()
	stacks[0].Server().AddPeer(p2pnode_mid)
	stacks[2].Server().AddPeer(p2pnode_mid)

	stop := func() {
		for i := 0; i < len(stacks); i++ {
			stacks[i].Stop()
		}
	}

	return node.DefaultWSPort, node.DefaultWSPort + 2, stop, nil
	//return ipc_left, ipc_right, stop, nil
}

func NewSwarmService(stack *node.Node, bzzport int) func(ctx *node.ServiceContext) (node.Service, error) {
	return func(ctx *node.ServiceContext) (node.Service, error) {
		// get the encrypted private key file
		keyid := fmt.Sprintf("%s/D3_Pss/nodekey", stack.DataDir())

		// load the private key from the file content
		prvkey, err := crypto.LoadECDSA(keyid)
		if err != nil {
			return nil, fmt.Errorf("privkey fail: %v", prvkey)
		}

		// create the swarm overlay address
		chbookaddr := crypto.PubkeyToAddress(prvkey.PublicKey)

		// configure and create a swarm instance
		bzzdir := stack.InstanceDir() // todo: what is the difference between this and datadir?

		swapEnabled := false
		syncEnabled := false
		pssEnabled := true
		cors := "*"

		bzzconfig, err := bzzapi.NewConfig(bzzdir, chbookaddr, prvkey, bzzNetworkId)
		bzzconfig.Port = fmt.Sprintf("%s", bzzport)
		if err != nil {
			Log.Crit("unable to configure swarm", "err", err)
		}
		//return swarm.NewSwarm(ctx, nil, nil, bzzconfig, swapEnabled, syncEnabled, cors, pssEnabled)
		return swarm.NewSwarm(ctx, nil, bzzconfig, swapEnabled, syncEnabled, cors, pssEnabled)
	}
}

// set up the local service node
func NewServiceNode(port int, httpport int, wsport int, modules ...string) (*node.Node, error) {
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", port)
	cfg.IPCPath = IPCName
	cfg.DataDir = Datadir(port)
	if httpport > 0 {
		cfg.HTTPHost = node.DefaultHTTPHost
		cfg.HTTPPort = httpport
		// cfg.HTTPModules = append(cfg.HTTPModules, "foo")
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

// return a uniform datadir name
func Datadir(port int) string {
	return fmt.Sprintf("%s/%s%d", BasePath, DatadirPrefix, port)
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
	var ids []discover.NodeID
	var addrs [][]byte
	for _, r := range rpcs {
		var nodeinfo p2p.NodeInfo
		err := r.Call(&nodeinfo, "admin_nodeInfo")
		if err != nil {
			return err
		}
		p2pnode, err := discover.ParseNode(nodeinfo.Enode)
		if err != nil {
			return err
		}
		ids = append(ids, p2pnode.ID)
		var bzzaddr []byte
		err = r.Call(&bzzaddr, "pss_baseAddr")
		if err != nil {
			return err
		}
		addrs = append(addrs, bzzaddr)
	}
	peerpot := network.NewPeerPot(minbinsize, ids, addrs)
	for {
		healthycount := 0
		for i, r := range rpcs {
			var health network.Health
			err := r.Call(&health, "hive_healthy", peerpot)
			if err != nil {
				return err
			}
			Log.Debug("health", "i", i, "addr", common.ToHex(addrs[i]), "id", ids[i], "info", health)
			if health.KnowNN && health.GotNN && health.Full {
				healthycount++
			}
		}
		if healthycount == len(rpcs) {
			break
		}
	}
	return nil
}
