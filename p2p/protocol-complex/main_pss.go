package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	swarmapi "github.com/ethereum/go-ethereum/swarm/api"

	"./bzz"
	"./service"
)

const (
	ipcName              = "pssdemo.ipc"
	defaultMaxDifficulty = 23
	defaultMaxJobs       = 3
	defaultMaxTime       = time.Second
)

var (
	loglevel = flag.Int("l", 3, "loglevel")
	port     = flag.Int("p", 30499, "p2p port")
	bzzport  = flag.String("b", "8555", "bzz port")
	enode    = flag.String("e", "", "enode to connect to")
	httpapi  = flag.String("a", "localhost:8545", "http api")
)

func init() {
	flag.Parse()
	log.Root().SetHandler(log.CallerFileHandler(log.LvlFilterHandler(log.Lvl(*loglevel), (log.StreamHandler(os.Stderr, log.TerminalFormat(true))))))
}

func main() {

	datadir, err := ioutil.TempDir("", "pssmailboxdemo-")
	if err != nil {
		log.Error("dir create fail", "err", err)
		return
	}
	defer os.RemoveAll(datadir)

	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", *port)
	cfg.P2P.EnableMsgEvents = true
	cfg.IPCPath = ipcName

	httpspec := strings.Split(*httpapi, ":")
	httpport, err := strconv.ParseInt(httpspec[1], 10, 0)
	if err != nil {
		log.Error("node create fail", "err", err)
		return
	}

	if *httpapi != "" {
		cfg.HTTPHost = httpspec[0]
		cfg.HTTPPort = int(httpport)
		cfg.HTTPModules = []string{"demo", "admin", "pss"}
	}
	cfg.DataDir = datadir

	stack, err := node.New(cfg)
	if err != nil {
		log.Error("node create fail", "err", err)
		return
	}

	// create the demo service, but now we don't register it directly
	// so we avoid the protocol running on the direct connected peers
	params := service.NewDemoParams(nil)
	params.MaxJobs = defaultMaxJobs
	params.MaxTimePerJob = defaultMaxTime
	params.MaxDifficulty = defaultMaxDifficulty
	svc, err := service.NewDemo(params)
	if err != nil {
		log.Error(err.Error())
		return
	}

	// create the pss service that wraps the demo protocol
	privkey, err := crypto.GenerateKey()
	if err != nil {
		log.Error(err.Error())
		return
	}

	bzzCfg := swarmapi.NewConfig()
	bzzCfg.SyncEnabled = false
	bzzCfg.Port = *bzzport
	bzzCfg.Path = datadir
	bzzCfg.HiveParams.Discovery = true
	bzzCfg.Init(privkey)

	bzzSvc, err := bzz.NewBzzService(bzzCfg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = bzzSvc.RegisterPssProtocol(svc)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return bzzSvc, nil
	}); err != nil {
		log.Error(err.Error())
		return
	}

	if err := stack.Start(); err != nil {
		log.Error(err.Error())
		return
	}
	defer stack.Stop()
	sigC := make(chan os.Signal)
	signal.Notify(sigC, syscall.SIGINT)
	<-sigC
}
