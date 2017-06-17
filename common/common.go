package common

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
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
	RemoteNode *discover.Node

	// self-explanatory command line arguments
	verbose      = flag.Bool("v", false, "more verbose logs")
	remoteport   = flag.Int("c", 0, "remote port (enables remote RPC lookup of enode")
	enode        = flag.String("e", "", "enode to connect to (overrides remote RPC lookup)")
	p2plocalport = flag.Int("p", P2PDefaultPort, "local port for p2p connections")
)

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

	// if the enode argument is empty and we have RPC argument, try to fetch the enode from the RPC
	if *enode == "" && *remoteport > 0 {
		*enode, err = getEnodeFromRPC(fmt.Sprintf("%s/%s%d/%s", OurDir, DatadirPrefix, *remoteport, IPCName))
		if err != nil {
			Log.Warn("Can't connect to remove RPC", "err", err)
		}
	}

	// turn the enode string into an abstract p2p node representation
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
