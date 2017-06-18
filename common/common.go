package common

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
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
	OurDir string

	// out local port for p2p connections
	P2PPort int

	// predictable datadir name prefix
	DatadirPrefix = ".data_"

	// ServiceNode is the base service stack of the go-ethereum codebase
	// it starts the tcp socket (p2p server) with protocols, and handles the APIs
	ServiceNode *node.Node

	// RemoteNode is an abstract representation of the other end of a tcp-connection
	// We use this object to set up a p2p connection,
	// and to retrieve data necessary to perform p2p communications
	RemoteNode *discover.Node

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
	OurDir, err = os.Getwd()
	Log.Debug("htewyt")
	if err != nil {
		Log.Crit("Could not determine working directory", "err", err)
	}

	// ensure good log formats for terminal
	// handle verbosity flag
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	loglevel := log.LvlDebug
	if *verbose {
		loglevel = log.LvlTrace
	}
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)

	// remote node
	//
	// if the enode argument is empty and we have RPC argument, try to fetch the enode from the RPC
	if *enode == "" && *remoteport > 0 {
		*enode, err = getEnodeFromRPC(fmt.Sprintf("%s/%s%d/%s", OurDir, DatadirPrefix, *remoteport, IPCName))
		if err != nil {
			Log.Warn("Can't connect to remote RPC", "err", err)
		}
	}

	// if we have an enode string now, use it to get the p2p node representation
	if *enode != "" {
		remotenodeptr, err := discover.ParseNode(*enode)
		if err != nil {
			Log.Warn("Can't create pointer for remote node", "err", err, "enode", *enode)
		}
		RemoteNode = remotenodeptr
	}
}

// set up the local service node
func SetupNode() (err error) {
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", *p2plocalport)
	cfg.IPCPath = IPCName
	cfg.DataDir = fmt.Sprintf("%s/%s%d", OurDir, DatadirPrefix, *p2plocalport)
	ServiceNode, err = node.New(cfg)
	return
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
