package main

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

var (
	// compiled with evm cli, commit 43dd8e62

	// bytecode is generated from this assembly code
	// get input value
	// store it to first storage slot
	// load from storage slot to memory
	// return first 4 bytes from memory slot
	// (we will call this code with 0xdeadbeef = 4 bytes)
	//
	//	push 0
	//	calldataload
	//	push 0
	//	sstore
	//	push 0
	//	sload
	//	push 0
	//	mstore
	//	push 4
	//	push 0
	//	return
	//
	prochex    = "60003560005560005460005260046000f3"
	prochexlen = len(prochex) / 2

	// "deploy code":
	// copy bytes of code from offset 0x0c to memory and return
	//
	//	push PROCHEXLEN
	//	push 0
	//	push PROCHEXLEN
	//	push 0x0c
	//	push 0
	//	codecopy
	//	return
	//
	deployhex = fmt.Sprintf("60%x600060%x600c600039f3", prochexlen, prochexlen)

	input = []byte{0xde, 0xad, 0xbe, 0xef}

	bytecode = common.FromHex(strings.Join([]string{deployhex, prochex}, ""))
	balance  = big.NewInt(int64(1000000000000))

	blocknumber = big.NewInt(0)
)

type fakeBackend interface {
	bind.ContractBackend
	Commit()
}

func init() {
	//log.StreamHandler(os.Stderr, nil)
	//h := log.LvlFilterHandler(log.LvlTrace, log.StdoutHandler)
	//log.Root().SetHandler(h)
}

// increment block number when we "mine"
func commit(sim fakeBackend) {
	sim.Commit()
	blocknumber.Add(blocknumber, big.NewInt(1))
}

func main() {

	// create key and derive address
	privkeysend, err := crypto.GenerateKey()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	sendaddr := crypto.PubkeyToAddress(privkeysend.PublicKey)

	// set up backend and slip it some dineros
	auth := bind.NewKeyedTransactor(privkeysend)
	alloc := make(core.GenesisAlloc, 1)
	alloc[sendaddr] = core.GenesisAccount{
		PrivateKey: crypto.FromECDSA(privkeysend),
		Balance:    balance,
	}
	sim := backends.NewSimulatedBackend(alloc)

	// set some foo values for transactions
	nonce := uint64(0)
	gaslimit := big.NewInt(2000000)
	amount := big.NewInt(0)
	gasprice := big.NewInt(20)

	// Create the contract creation transaction
	// sign it and schedule it for execution
	rawTx := types.NewContractCreation(nonce, amount, gaslimit, gasprice, bytecode)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	signedTx, err := auth.Signer(types.HomesteadSigner{}, sendaddr, rawTx)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	txhash := signedTx.Hash()
	err = sim.SendTransaction(context.TODO(), signedTx)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Mine one block
	commit(sim)

	// get the receipt, which should contain the contract address
	rcpt, err := sim.TransactionReceipt(context.TODO(), txhash)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	contractaddress := rcpt.ContractAddress
	if len(contractaddress) == 0 {
		log.Error("empty contract address")
		os.Exit(1)
	}

	// retrieve the code from the contract address
	// verify that it matches the code we sent
	code, err := sim.CodeAt(context.TODO(), contractaddress, big.NewInt(1))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	} else if !bytes.Equal(code, common.FromHex(prochex)) {
		log.Error("Code at contract address does not match input bytecode", "code", fmt.Sprintf("%x", code))
		os.Exit(1)
	}

	// call the runtime bytecode (not as transaction)
	// should not alter the storage, but should return 0xdeadbeef
	msg := ethereum.CallMsg{
		From:     sendaddr,
		To:       &rcpt.ContractAddress,
		GasPrice: gasprice,
		Gas:      gaslimit,
		Value:    amount,
		Data:     input,
	}
	r, err := sim.CallContract(context.TODO(), msg, blocknumber)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	} else if !bytes.Equal(r, input) {
		log.Error("Expected return value deadbeef", "got", fmt.Sprintf("%x", r))
		os.Exit(1)
	}
	storageat, err := sim.StorageAt(context.TODO(), contractaddress, common.BigToHash(big.NewInt(0)), blocknumber)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	} else if !bytes.Equal(storageat, make([]byte, 32)) {
		log.Info("storageat should be empty", "contains", fmt.Sprintf("%x", storageat))
		os.Exit(1)
	}

	// Create the transaction call, sign it and schedule
	nonce, err = sim.NonceAt(context.TODO(), contractaddress, blocknumber)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	rawTx = types.NewTransaction(nonce, contractaddress, amount, gaslimit, gasprice, input)
	signedTx, err = auth.Signer(types.HomesteadSigner{}, sendaddr, rawTx)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	txhash = signedTx.Hash()
	err = sim.SendTransaction(context.TODO(), signedTx)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// mine another block
	commit(sim)

	// now storage should have been changed
	// index 0 should hold 0xdeadbeef
	storageat, err = sim.StorageAt(context.TODO(), contractaddress, common.BigToHash(big.NewInt(0)), blocknumber)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// seeing is believing
	fmt.Printf("%x\n", storageat[:4])
}
