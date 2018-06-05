// multiple services in same node
package main

import (
	// "fmt"
	"os"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	demo "./common"
)

// the fooservice retrieves the shared value
type fooService struct {
	v *int
}

func newFooService(v *int) *fooService {
	return &fooService{
		v: v,
	}
}

func (self *fooService) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "foo",
			Version:   "0.42",
			Service:   &FooAPI{self.v},
			Public:    true,
		},
	}
}

func (self *fooService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

func (self *fooService) Start(srv *p2p.Server) error {
	return nil
}

func (self *fooService) Stop() error {
	return nil
}

type FooAPI struct {
	v *int
}

func (api *FooAPI) Get() (int, error) {
	return *api.v, nil
}

// the barservice sets the shared value
type barService struct {
	v *int
}

func newBarService(v *int) *barService {
	return &barService{
		v: v,
	}
}

func (self *barService) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "bar",
			Version:   "0.42",
			Service:   &BarAPI{self.v},
			Public:    true,
		},
	}
}

func (self *barService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

func (self *barService) Start(srv *p2p.Server) error {
	return nil
}

func (self *barService) Stop() error {
	return nil
}

type BarAPI struct {
	v *int
}

func (api *BarAPI) Set(n int) error {
	*api.v = n
	return nil
}

func main() {

	var sharedvalue int

	stack, err := demo.NewServiceNode(demo.P2pPort, 0, 0)

	// register two separate services
	foosvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return newFooService(&sharedvalue), nil
	}
	err = stack.Register(foosvc)
	if err != nil {
		demo.Log.Crit("Register fooservice in servicenode failed", "err", err)
	}

	barsvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return newBarService(&sharedvalue), nil
	}
	err = stack.Register(barsvc)
	if err != nil {
		demo.Log.Crit("Register barservice in servicenode failed", "err", err)
	}
	defer os.RemoveAll(stack.DataDir())

	// start the node
	err = stack.Start()
	if err != nil {
		demo.Log.Crit("servicenode start failed", "err", err)
	}

	// set the shared value in service bar
	rpcclient, err := stack.Attach()
	err = rpcclient.Call(nil, "bar_set", 42)
	if err != nil {
		demo.Log.Crit("Could not get rpcclient via p2p.Server", "err", err)

	}

	// get the shared value in service foo
	var result int
	err = rpcclient.Call(&result, "foo_get")
	if err != nil {
		demo.Log.Crit("Could not get rpcclient via p2p.Server", "err", err)

	}
	demo.Log.Info("get", "result", result, "sharedvalue", sharedvalue)

	// bring down the servicenode
	stack.Stop()
}
