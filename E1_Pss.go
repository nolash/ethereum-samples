// pss send-to-self hello world
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/swarm"
	bzzapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/pss"
	demo "github.com/nolash/go-ethereum-p2p-demo/common"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlDebug, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
}

func newService(bzzdir string, bzzport int, bzznetworkid uint64) func(ctx *node.ServiceContext) (node.Service, error) {
	return func(ctx *node.ServiceContext) (node.Service, error) {

		// generate a new private key
		privkey, err := crypto.GenerateKey()
		if err != nil {
			demo.Log.Crit("private key generate servicenode 'left' fail: %v")
		}

		// create the swarm overlay address
		chbookaddr := crypto.PubkeyToAddress(privkey.PublicKey)

		// configure and create a swarm instance
		//bzzdir := stack.InstanceDir() // todo: what is the difference between this and datadir?

		swapEnabled := false
		syncEnabled := false
		pssEnabled := true
		cors := "*"

		bzzconfig, err := bzzapi.NewConfig(bzzdir, chbookaddr, privkey, bzznetworkid)
		bzzconfig.Port = fmt.Sprintf("%s", bzzport)
		if err != nil {
			demo.Log.Crit("unable to configure swarm", "err", err)
		}
		return swarm.NewSwarm(ctx, nil, bzzconfig, swapEnabled, syncEnabled, cors, pssEnabled)
	}
}

func main() {

	// get the input params
	l_port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		demo.Log.Crit(err.Error())
	}
	r_port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		demo.Log.Crit(err.Error())
	}
	networkid, err := strconv.Atoi(os.Args[3])
	if err != nil {
		demo.Log.Crit(err.Error())
	}

	// create two nodes
	l_stack, err := demo.NewServiceNode(demo.P2PDefaultPort, 0, 0)
	if err != nil {
		demo.Log.Crit(err.Error())
	}
	r_stack, err := demo.NewServiceNode(demo.P2PDefaultPort+1, 0, 0)
	if err != nil {
		demo.Log.Crit(err.Error())
	}

	// load the private key from the file content
	//keyfile := filepath.Join(stack.DataDir(), "nodekey")
	//prvkey, err := crypto.LoadECDSA(keyid)
	//if err != nil {
	//	return nil, fmt.Errorf("privkey fail: %v", prvkey)
	//}

	// register the service bundles
	l_svc := newService(l_stack.InstanceDir(), l_port, uint64(networkid))
	err = l_stack.Register(l_svc)
	if err != nil {
		demo.Log.Crit("servicenode 'left' pss register fail", "err", err)
	}
	r_svc := newService(r_stack.InstanceDir(), r_port, uint64(networkid))
	err = r_stack.Register(r_svc)
	if err != nil {
		demo.Log.Crit("servicenode 'right' pss register fail", "err", err)
	}

	// start the nodes
	err = l_stack.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}
	err = r_stack.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}

	// connect the nodes
	l_stack.Server().AddPeer(r_stack.Server().Self())

	// get the rpc clients
	l_rpcclient, err := l_stack.Attach()
	r_rpcclient, err := r_stack.Attach()

	// wait until the state of the swarm overlay network is ready
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() {
		err = demo.WaitHealthy(ctx, 2, l_rpcclient, r_rpcclient)
		if err != nil {
			demo.Log.Crit("health check fail", "err", err)
		}
	}()
	//time.Sleep(time.Second)

	// get a valid topic byte
	var topic pss.Topic
	err = l_rpcclient.Call(&topic, "pss_stringToTopic", "foo")
	if err != nil {
		demo.Log.Crit("pss string to topic fail", "err", err)
	}

	// subscribe to incoming messages on the receiving sevicenode
	// this will register a message handler on the specified topic
	msgC := make(chan pss.APIMsg)
	sub, err := r_rpcclient.Subscribe(context.Background(), "pss", msgC, "receive", topic)

	// get the receiver's swarm overlay address
	var r_bzzaddr []byte
	err = r_rpcclient.Call(&r_bzzaddr, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("pss get addr fail", "err", err)
	}

	// get the receiver's public key
	var r_pubkey []byte
	err = r_rpcclient.Call(&r_pubkey, "pss_getPublicKey")
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}

	// make the sender aware of the receiver's public key
	err = l_rpcclient.Call(&r_pubkey, "pss_setPeerPublicKey", r_pubkey, topic, r_bzzaddr)
	if err != nil {
		demo.Log.Crit("pss get pubkey fail", "err", err)
	}

	// convert the pubkey to hex string
	// we need this for the send api call
	pubkeyid := common.ToHex(r_pubkey)

	// send message using asymmetric encryption
	// since it's sent to ourselves, it will not go through pss forwarding
	err = l_rpcclient.Call(nil, "pss_sendAsym", pubkeyid, topic, []byte("bar"))
	if err != nil {
		demo.Log.Crit("pss send fail", "err", err)
	}

	// get the incoming message
	inmsg := <-msgC
	demo.Log.Info("pss received", "msg", string(inmsg.Msg), "from", fmt.Sprintf("%x", inmsg.Key))

	// bring down the servicenode
	sub.Unsubscribe()
	r_rpcclient.Close()
	l_rpcclient.Close()
	r_stack.Stop()
	l_stack.Stop()
}
