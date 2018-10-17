package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

const (
	defaultGasLimit = 21000
	defaultGasPrice = 5 * 1000000000 // 40 gwei
	ethSzabo        = float64(1000000)
)

var (
	g_amount   uint64
	g_gasPrice uint64
	g_gasLimit uint64
	g_dir      string
	g_nonce    uint64
	g_from     string
	g_to       string
)

func init() {

	var debug bool

	flag.StringVar(&g_from, "f", "", "Account to send transaction from")
	flag.StringVar(&g_to, "t", "", "Account to send transaction to")
	flag.StringVar(&g_dir, "d", "", "datadir")
	flag.Uint64Var(&g_nonce, "n", 0, "nonce")
	flag.BoolVar(&debug, "debug", false, "Output debug information")
	flag.Parse()

	lvl := log.LvlError
	if debug {
		lvl = log.LvlDebug
	}
	h := log.LvlFilterHandler(lvl, log.StderrHandler)
	log.Root().SetHandler(h)

	if g_from == "" {
		log.Error("Need from address")
		os.Exit(1)
	}
	if g_to == "" {
		log.Error("Need to-address")
		os.Exit(1)
	}
	amount, err := strconv.ParseFloat(flag.Arg(0), 32)
	if err != nil {
		log.Error("Invalid amount")
		os.Exit(1)
	}
	g_amount = uint64(amount * ethSzabo)

	if g_dir == "" {
		usr, err := user.Current()
		if err != nil {
			log.Error("Could not get user info to resolve homedir")
			os.Exit(1)
		}
		g_dir = fmt.Sprintf("%s/.ethereum/keystore", usr.HomeDir)
	}
	fi, err := os.Stat(g_dir)
	if err != nil {
		log.Error("datadir invalid", "reason", err)
		os.Exit(1)
	} else if !fi.IsDir() {
		log.Error("not a directory", "path", g_dir)
		os.Exit(1)
	}
	g_gasPrice = defaultGasPrice
	g_gasLimit = defaultGasLimit
}

func main() {

	store := keystore.NewKeyStore(g_dir, keystore.StandardScryptN, keystore.StandardScryptP)

	var from accounts.Account
	var wallet accounts.Wallet
	for _, w := range store.Wallets() {
		for _, a := range w.Accounts() {
			if a.Address == common.HexToAddress(g_from) {
				from = a
				wallet = w
			}
		}
	}

	zeroaccount := accounts.Account{}
	if from == zeroaccount {
		log.Error("From address not valid", "address", g_from)
		os.Exit(1)
	}

	to := common.HexToAddress(g_to)
	var wei big.Int
	szaboWei := big.NewInt(1000000000000)
	wei.Mul(big.NewInt(int64(g_amount)), szaboWei)
	//wei.Set(big.NewInt(1))
	log.Debug("Creating transaction", "from", from, "to", to.Hex(), "amount", wei.Text(10), "nonce", g_nonce, "gaslimit", int64(g_gasLimit), "gasprice", int64(g_gasPrice))
	tx := types.NewTransaction(g_nonce, to, &wei, big.NewInt(int64(g_gasLimit)), big.NewInt(int64(g_gasPrice)), []byte{})

	fmt.Printf("pass: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	passphrase := string(bytePassword)

	signedTx, err := wallet.SignTxWithPassphrase(from, passphrase, tx, nil)
	if err != nil {
		log.Error("Transaction hash sign fail", "reason", err)
		os.Exit(1)
	}
	rawTx := bytes.NewBuffer(nil)
	err = signedTx.EncodeRLP(rawTx)
	if err != nil {
		log.Error("Transaction RLP encode fail", "reason", err)
		os.Exit(1)
	}
	fmt.Printf("Tx: %x\nraw: %x\n", signedTx.Hash(), rawTx)
}
