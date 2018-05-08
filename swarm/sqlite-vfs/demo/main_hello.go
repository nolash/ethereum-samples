package main

import (
	"crypto/rand"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	sys "golang.org/x/sys/unix"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
	sqlite3 "github.com/mattn/go-sqlite3"

	"./hello"
)

const (
	SQL_DRIVER         = "sqlite" // if "sqlite3" is supplied, it fails complaining name is already taken. Conflict with cgo backend?
	DEFAULT_DATACOUNT  = 100
	DEFAULT_DATASIZE   = 1024
	SQL_EXTRA_FACTOR   = 1.5 // allow for extra bytes of data per row for calculating disk size
	CHUNK_EXTRA_FACTOR = 1.1
	CHUNKDIR_NAME      = "chunks"
	DB_NAME            = "hello.db"
)

var (
	dataDir   string
	dataFile  string
	newChunks bool
	keep      bool
	dataSize  uint64
	dataCount uint64
	dbSize    uint64
	chunkDir  string
)

func init() {

	// flags flags flags
	flag.BoolVar(&keep, "k", false, "don't delete datadir after running")
	flag.StringVar(&dataDir, "d", os.TempDir(), "dir to use for datadir")
	flag.StringVar(&dataFile, "f", "", "existing db file to open (must exist, Implies -k)")
	flag.Uint64Var(&dataSize, "s", DEFAULT_DATASIZE, "blob value size per row")
	flag.Uint64Var(&dataCount, "c", DEFAULT_DATACOUNT, "number of rows to generate")
	flag.BoolVar(&newChunks, "u", true, "if exists, replace chunkstore in datadir. Implies -d")
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "verbose debug output")
	var veryverbose bool
	flag.BoolVar(&veryverbose, "vv", false, "VERY verbose debug output")
	var cverbose bool
	flag.BoolVar(&cverbose, "vc", false, "include debug output from c backend")
	var help bool
	flag.BoolVar(&help, "h", false, "show this usage information")
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}
	// calculate capacities we need
	disksize := uint64(float64(dataSize)) * dataCount
	dbSize = uint64(float64(disksize) * 1.1 * SQL_EXTRA_FACTOR)

	// chunk specifics
	chunkDir = filepath.Join(dataDir, CHUNKDIR_NAME)

	// debugging
	loglevel := log.LvlInfo
	if veryverbose {
		loglevel = log.LvlTrace
	} else if verbose {
		loglevel = log.LvlDebug
	}
	log.Root().SetHandler(log.CallerFileHandler(log.LvlFilterHandler(loglevel, log.StreamHandler(os.Stderr, log.TerminalFormat(true)))))

	if cverbose {
		hello.Debug(true)
	}
}

func main() {

	var err error

	// check if we have arg, and if it's file or dir

	var dataurl string
	var stat sys.Statfs_t
	var actualdbsize int64
	var diskreq uint64
	checkavail := true
	// if we are using existing db
	if dataFile != "" {
		// must exist
		dataurl = filepath.Join(dataDir, dataFile)
		fi, err := os.Stat(dataurl)
		if err != nil {
			log.Crit(err.Error())
		}
		actualdbsize = fi.Size()
		_, err = os.Stat(chunkDir)
		if err == nil {
			if newChunks {
				os.RemoveAll(chunkDir)
				checkavail = false
			}
		} else {
			tmpdir, err := ioutil.TempDir(dataDir, fmt.Sprintf("swarmvfs-%dx%d-", dataCount, dataSize))
			if err != nil {
				log.Crit("failed to create datadir in '%s'", "err", err)
			}
			dataDir = tmpdir
			req := float64(dbSize) * CHUNK_EXTRA_FACTOR
			diskreq = uint64(req)
		}
	} else {
		_, err = os.Stat(filepath.Join(dataDir, DB_NAME))
		if err == nil {
			log.Crit("db file '%s' already exists", filepath.Join(dataDir, DB_NAME))
		}
		tmpdir, err := ioutil.TempDir(dataDir, fmt.Sprintf("swarmvfs-%dx%d-", dataCount, dataSize))
		if err != nil {
			log.Crit("failed to create datadir in '%s'", "err", err)
		}
		dataDir = tmpdir
		dataurl = filepath.Join(dataDir, DB_NAME)
		req := float64(dbSize) * (CHUNK_EXTRA_FACTOR * SQL_EXTRA_FACTOR)
		diskreq = uint64(req)
	}
	if checkavail {
		sys.Statfs(dataDir, &stat)
		diskavail := stat.Bavail * uint64(stat.Bsize)
		if diskavail < diskreq {
			log.Crit("insufficient disk space", "have", diskavail, "need", diskreq)
		}
	}

	// don't delete it was an existing datafile
	if !keep {
		if dataFile != "" {
			log.Warn("existing datafile specified, implying -k")
		} else {
			defer os.RemoveAll(dataDir)
		}
	}

	log.Debug("Hello!", "datadir", dataDir, "dbfile", dataFile, "keep", keep, "heuristic dbsize", dbSize)
	if dataFile == "" {
		// create a database and add some values
		sql.Register(SQL_DRIVER, &sqlite3.SQLiteDriver{})
		db, err := sql.Open(SQL_DRIVER, dataurl)
		if err != nil {
			log.Crit(err.Error())
		}

		_, err = db.Exec(`CREATE TABLE hello (
			id INT UNSIGNED NOT NULL,
			val BLOB NOT NULL

		);
		CREATE INDEX hello_idx ON hello(id);
		`)
		if err != nil {
			log.Crit(err.Error())
		}

		for i := uint64(0); i < dataCount; i++ {
			data := make([]byte, dataSize)
			c, err := rand.Read(data)
			if err != nil {
				log.Crit(err.Error())
			} else if uint64(c) < dataSize {
				log.Crit("shortread")
			}
			_, err = db.Exec(`INSERT INTO hello (id, val) VALUES 
				($1, $2)`,
				i,
				data)
			if err != nil {
				log.Crit(err.Error())
			}
			log.Trace("insert", "id", i, "data", fmt.Sprintf("%x", data[:8]))
		}
		err = db.Close()
		if err != nil {
			log.Crit(err.Error())
		}
		fi, err := os.Stat(dataurl)
		if err != nil {
			log.Crit(err.Error())
		}
		actualdbsize = fi.Size()
	}

	chunkcount := float64(actualdbsize/storage.CHUNKSIZE) * CHUNK_EXTRA_FACTOR
	log.Debug("Data ok, passing on to swarm", "actual dbsize", actualdbsize, "chunkcount", uint64(chunkcount))

	// create the chunkstore and start dpa on it
	dpa, err := storage.NewLocalDPA(filepath.Join(dataDir, CHUNKDIR_NAME), "SHA3", uint64(chunkcount), 5000)
	if err != nil {
		log.Crit(err.Error())
	}
	dpa.Start()
	defer dpa.Stop()

	err = hello.Init(dpa)
	if err != nil {
		log.Crit("init fail", "err", err)
	}

	// stick the database in the chunker, muahahaa
	r, err := os.Open(dataurl)
	if err != nil {
		log.Crit(err.Error())
	}
	fi, err := r.Stat()
	if err != nil {
		log.Crit(err.Error())
	}
	swg := &sync.WaitGroup{}
	wwg := &sync.WaitGroup{}
	key, err := dpa.Store(r, fi.Size(), swg, wwg)
	if err != nil {
		log.Crit(err.Error())
	}
	log.Debug("store", "key", key)

	// test the sqlite_vfs bzz backend:

	// open the database
	err = hello.Open(key)
	if err != nil {
		log.Crit("open fail", "err", err)
	}

	// execute query
	err = hello.Exec(fmt.Sprintf("SELECT * FROM hello WHERE id = %d or id = %d", 1, dataCount-1))
	if err != nil {
		log.Crit("exec fail", "err", err)
	}
}
