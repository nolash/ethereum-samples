package escrow

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

var (
	ownerPrivKey *ecdsa.PrivateKey
	ownerAddress common.Address
	transactors  []*bind.TransactOpts
	privKeys     []*ecdsa.PrivateKey
	addresses    []common.Address
	debugflag    = flag.Bool("vv", false, "verbose output")
)

func init() {
	var err error
	flag.Parse()
	loglevel := log.LvlInfo
	if *debugflag {
		loglevel = log.LvlTrace
	}
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)

	for i := 0; i < 4; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			panic(fmt.Sprintf("keygen fail: %v", err))
		}
		privKeys = append(privKeys, key)
		addresses = append(addresses, crypto.PubkeyToAddress(key.PublicKey))
		transactors = append(transactors, bind.NewKeyedTransactor(key))
		log.Info("Generate key", "#", i, "privKey", fmt.Sprintf("%x", crypto.FromECDSA(privKeys[i])), "address", fmt.Sprintf("%x", addresses[i]))
	}

	ownerPrivKey, err = crypto.GenerateKey()
	if err != nil {
		panic(fmt.Sprintf("keygen fail: %v", err))
	}
	ownerAddress = crypto.PubkeyToAddress(ownerPrivKey.PublicKey)

}

func newTestBackend() *backends.SimulatedBackend {
	return backends.NewSimulatedBackend(core.GenesisAlloc{
		addresses[0]: {Balance: big.NewInt(10000000000)},
		addresses[1]: {Balance: big.NewInt(10000000000)},
		addresses[2]: {Balance: big.NewInt(10000000000)},
		addresses[3]: {Balance: big.NewInt(10000000000)},
		ownerAddress: {Balance: big.NewInt(10000000000)},
	})
}

func TestEscrow(t *testing.T) {
	backend := newTestBackend()
	deployTransactor := bind.NewKeyedTransactor(ownerPrivKey)
	deployTransactor.Value = big.NewInt(0)

	// deploy escrow contract
	escrow_addr, escrow_trans, escrow, err := DeployEscrow(deployTransactor, backend)
	if err != nil {
		t.Fatal(err)
	}
	backend.Commit()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	createC := make(chan *EscrowCreated)
	getC := make(chan interface{})
	createFilterOpts := &bind.WatchOpts{
		Start:   nil,
		Context: ctx,
	}

	escrow.EscrowFilterer.WatchCreated(createFilterOpts, createC)
	go func() {
		select {
		case ev := <-createC:
			getC <- ev
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}()

	// create a new escrow
	v, err := escrow.CreateEscrow(transactors[0], [8]byte{0x66, 0x6f, 0x64})
	if err != nil {
		t.Fatal(err)
	}
	backend.Commit()
	log.Info("createescrow return", "v", v)
	ev := <-getC
	log.Warn("event", "ev", ev)

	// owner add a participant
	v, err = escrow.AddParticipant(transactors[0], ev.(*EscrowCreated).Seq, big.NewInt(4200), addresses[1], addresses[0])
	if err != nil {
		t.Fatal(err)
	}
	backend.Commit()

	// owner add previously added participant -> FAIL
	v, err = escrow.AddParticipant(transactors[0], ev.(*EscrowCreated).Seq, big.NewInt(4200), addresses[1], addresses[0])
	if err == nil {
		t.Fatalf("should have failed: %v", v)
	}
	backend.Commit()

	// non-owner add a partiticant -> FAIL
	v, err = escrow.AddParticipant(transactors[1], ev.(*EscrowCreated).Seq, big.NewInt(4200), addresses[1], addresses[0])
	if err == nil {
		t.Fatalf("should have failed: %v", v)
	}
	backend.Commit()

	_ = escrow_addr
	_ = escrow_trans
}
