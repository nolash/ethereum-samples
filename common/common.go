package common

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"os"
)

const (
	P2PDefaultPort = 30100
	IPCName        = "demo.ipc"
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

	//	// remote node
	//	//
	//	// if the enode argument is empty and we have RPC argument, try to fetch the enode from the RPC
	//	if *enode == "" && *remoteport > 0 {
	//		*enode, err = getEnodeFromRPC(fmt.Sprintf("%s/%s%d/%s", BasePath, DatadirPrefix, *remoteport, IPCName))
	//		if err != nil {
	//			Log.Warn("Can't connect to remote RPC", "err", err)
	//		}
	//	}
	//
	//	// if we have an enode string now, use it to get the p2p node representation
	//	if *enode != "" {
	//		remotenodeptr, err := discover.ParseNode(*enode)
	//		if err != nil {
	//			Log.Warn("Can't create pointer for remote node", "err", err, "enode", *enode)
	//		}
	//		RemoteNode = remotenodeptr
	//	}
}

// set up the local service node
func NewServiceNode(port int, httpport int, wsport int) (*node.Node, error) {
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
