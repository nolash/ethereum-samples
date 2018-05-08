package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/storage"

	colorable "github.com/mattn/go-colorable"
)

var (
	dir     = flag.String("d", "", "datadir")
	verbose = flag.Bool("v", false, "verbose")
	pass    = flag.String("p", "", "account password")
	show    = flag.Int("n", 10, "amount to show")
)

func init() {
	flag.Parse()
	if *dir == "" {
		var ok bool
		home, ok := os.LookupEnv("HOME")
		if !ok {
			panic("unknown datadir")
		}
		*dir = fmt.Sprintf("%s/.ethereum", home)
	}

	if *verbose {
		log.PrintOrigins(true)
		log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))

	}
}

func main() {

	name := flag.Arg(0)

	backends := []accounts.Backend{
		keystore.NewKeyStore(fmt.Sprintf("%s/keystore", *dir), keystore.StandardScryptN, keystore.StandardScryptP),
	}
	accman := accounts.NewManager(backends...)
	var bzzdir string
	var baseaddr common.Hash
	//var bzzkey string
	var resourceSigner storage.ResourceSigner
OUTER:
	for _, w := range accman.Wallets() {
		for _, a := range w.Accounts() {
			f, err := os.Open(a.URL.Path)
			if err != nil {
				log.Warn("cant open keyfile", "url", a.URL.Path)
				continue
			}
			defer f.Close()
			keyjson, err := ioutil.ReadAll(f)
			if err != nil {
				log.Warn("cant read keyfile", "url", a.URL.Path)
				continue
			}
			storekey, err := keystore.DecryptKey(keyjson, *pass)
			if err != nil {
				log.Warn("passphrase did not match account", "addr", a.Address)
				continue
			}
			pubkeybytes := crypto.FromECDSAPub(&storekey.PrivateKey.PublicKey)
			baseaddr = crypto.Keccak256Hash(pubkeybytes)
			trydir := fmt.Sprintf("%s/swarm/bzz-%x", *dir, a.Address)
			if f, err := os.Open(trydir); err == nil {
				f.Close()
				bzzdir = trydir
				//bzzkey = a.Address.Hex()
				resourceSigner = &signer{
					keystore: backends[0].(*keystore.KeyStore),
					account:  a,
				}
				break OUTER
			}

		}
	}
	if bzzdir == "" {
		log.Error("no chunkdir found")
		return
	}
	lstoreparams := storage.NewDefaultLocalStoreParams()
	log.Info("using keyhex", "key", baseaddr.Bytes())
	lstoreparams.BaseKey = baseaddr.Bytes()
	lstoreparams.Init(bzzdir)
	log.Info("actual chunkdir", "dir", lstoreparams.ChunkDbPath)

	// don't try this at home, kids
	// its safe here cos we're only writing
	linkdir, err := ioutil.TempDir("", "exec-")
	if err != nil {
		log.Error("can't create link dir", "err", err.Error())
		return
	}
	//defer os.RemoveAll(linkdir)
	chunkdir, err := ioutil.ReadDir(lstoreparams.ChunkDbPath)
	if err != nil {
		log.Error("can't open chunk dir", "err", err.Error())
		return
	}
	for _, d := range chunkdir {
		if d.Name() != "LOCK" {
			if err := os.Symlink(filepath.Join(lstoreparams.ChunkDbPath, d.Name()), filepath.Join(linkdir, d.Name())); err != nil {
				log.Error("link error", "err", err.Error())
				return
			}
		}
	}
	// (for now) same as Init() but won't exec again if already set
	lstoreparams.ChunkDbPath = linkdir
	log.Info("aliased chunkdir", "dir", lstoreparams.ChunkDbPath)

	lstore, err := storage.NewLocalStore(lstoreparams, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer lstore.Close()

	rhparams := &storage.ResourceHandlerParams{}
	rhparams.Signer = resourceSigner
	rhparams.EthClient = storage.NewBlockEstimator()
	rh, err := storage.NewResourceHandler(rhparams)
	if err != nil {
		log.Error(err.Error())
		return
	}
	rh.SetStore(lstore)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rsrc, err := rh.LookupLatestByName(ctx, name, true, &storage.ResourceLookupParams{})
	if err != nil {
		log.Error(err.Error())
		return
	}
	output(rh, name)
	var i int
	for ; ; i++ {
		rsrc, err = rh.LookupPreviousByName(ctx, name, &storage.ResourceLookupParams{})
		if err != nil {
			log.Warn(err.Error())
			break
		}
		if i < *show {
			output(rh, name)
		}
	}
	if i >= *show {
		fmt.Printf("... and %d more\n", i-*show)
	}
	_ = rsrc
}

func output(rh *storage.ResourceHandler, name string) {
	period, _ := rh.GetLastPeriod(name)
	version, _ := rh.GetVersion(name)
	key, content, _ := rh.GetContent(name)
	fmt.Printf("v%d.%d [%s]: ", period, version, key)
	if len(content) < 32 {
		fmt.Printf("%x\n", content)
	} else {
		fmt.Printf("%x ...\n", content[:32])
	}
}

// implements storage.ResourceSigner
type signer struct {
	keystore *keystore.KeyStore
	account  accounts.Account
}

func (s *signer) Sign(h common.Hash) (signature storage.Signature, err error) {
	signaturebytes, err := s.keystore.SignHash(s.account, h.Bytes())
	if err != nil {
		return signature, err
	}
	copy(signature[:], signaturebytes)
	return signature, nil
}
