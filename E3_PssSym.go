// pss send-to-self hello world
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/contracts/chequebook"
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

		// create necessary swarm params
		var ensApi chequebook.Backend = nil
		swapEnabled := false
		syncEnabled := false
		pssEnabled := true
		cors := "*"
		checkbookaddr := crypto.PubkeyToAddress(privkey.PublicKey)
		bzzconfig, err := bzzapi.NewConfig(bzzdir, checkbookaddr, privkey, bzznetworkid)
		if err != nil {
			demo.Log.Crit("unable to configure swarm", "err", err)
		}
		bzzconfig.Port = fmt.Sprintf("%d", bzzport)

		return swarm.NewSwarm(ctx, ensApi, bzzconfig, swapEnabled, syncEnabled, cors, pssEnabled)
	}
}

func main() {

	// create two nodes
	l_stack, err := demo.NewServiceNode(demo.P2PDefaultPort, 0, 0)
	if err != nil {
		demo.Log.Crit(err.Error())
	}
	r_stack, err := demo.NewServiceNode(demo.P2PDefaultPort+1, 0, 0)
	if err != nil {
		demo.Log.Crit(err.Error())
	}

	// register the pss activated bzz services
	l_svc := newService(l_stack.InstanceDir(), demo.BzzDefaultPort, demo.BzzDefaultNetworkId)
	err = l_stack.Register(l_svc)
	if err != nil {
		demo.Log.Crit("servicenode 'left' pss register fail", "err", err)
	}
	r_svc := newService(r_stack.InstanceDir(), demo.BzzDefaultPort+1, demo.BzzDefaultNetworkId)
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

	// connect the nodes to the middle
	l_stack.Server().AddPeer(r_stack.Server().Self())

	// get the rpc clients
	l_rpcclient, err := l_stack.Attach()
	r_rpcclient, err := r_stack.Attach()

	// wait until the state of the swarm overlay network is ready
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = demo.WaitHealthy(ctx, 2, l_rpcclient, r_rpcclient)
	if err != nil {
		demo.Log.Crit("health check fail", "err", err)
	}
	time.Sleep(time.Second) // because the healthy does not work

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

	// get the recipient node's swarm overlay address
	var l_bzzaddr []byte
	err = r_rpcclient.Call(&l_bzzaddr, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("pss get baseaddr fail", "err", err)
	}
	var r_bzzaddr []byte
	err = r_rpcclient.Call(&r_bzzaddr, "pss_baseAddr")
	if err != nil {
		demo.Log.Crit("pss get baseaddr fail", "err", err)
	}

	symkey := make([]byte, 32)
	c, err := rand.Read(symkey)
	if err != nil {
		demo.Log.Crit("symkey gen fail", "err", err)
	} else if c < 32 {
		demo.Log.Crit("symkey size mismatch, expected 32", "size", c)
	}

	var l_symkeyid string
	err = l_rpcclient.Call(&l_symkeyid, "pss_setSymmetricKey", symkey, topic, r_bzzaddr, true)
	if err != nil {
		demo.Log.Crit("pss set symkey fail", "err", err)
	}

	var r_symkeyid string
	err = r_rpcclient.Call(&r_symkeyid, "pss_setSymmetricKey", symkey, topic, l_bzzaddr, true)
	if err != nil {
		demo.Log.Crit("pss set symkey fail", "err", err)
	}

	// send message using asymmetric encryption
	// since it's sent to ourselves, it will not go through pss forwarding
	err = l_rpcclient.Call(nil, "pss_sendSym", l_symkeyid, topic, []byte("bar"))
	if err != nil {
		demo.Log.Crit("pss send fail", "err", err)
	}

	// get the incoming message
	inmsg := <-msgC
	demo.Log.Info("pss received", "msg", string(inmsg.Msg), "from", fmt.Sprintf("%x", inmsg.Key))

	// bring down the servicenodes
	sub.Unsubscribe()
	r_rpcclient.Close()
	l_rpcclient.Close()
	r_stack.Stop()
	l_stack.Stop()
}
