package minipow

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"os"
	"testing"
	"time"
)

const (
	showlimit = 32 + 8 + 3
)

var (
	diff       = flag.Int("d", 8, "difficulty")
	verbose    = flag.Bool("vv", false, "verbose")
	datalen    = flag.Int("l", 128, "data length (will be randomized")
	timeout    = flag.Int("t", 10, "timeout in seconds")
	showoffset = 0
	showprefix = ""
)

func init() {
	flag.Parse()
	if *datalen > showlimit {
		showoffset = *datalen - showlimit
		showprefix = "..."
	}
}

func debug(data []byte, sum []byte) {
	fmt.Printf("Trying %s%x ==> %x\n", showprefix, data[showoffset:], sum)
}

func TestMine(t *testing.T) {

	var debugFunc func([]byte, []byte)
	if *verbose {
		debugFunc = debug
	}

	fmt.Printf("Datalength %d, difficulty 2^%d = %d\n", *datalen, *diff, int(math.Pow(2, float64(*diff))))

	// make up some data
	data := make([]byte, *datalen+8)
	rand.Read(data[:*datalen])

	// start mining
	resultC := make(chan []byte)
	quitC := make(chan struct{})
	go Mine(data, *diff, resultC, quitC, debugFunc)

	// set timeout, wait for return or cancel if it runs too long
	timeoutduration, err := time.ParseDuration(fmt.Sprintf("%ds", *timeout))
	if err != nil {
		t.Fatalf("couldn't parse timeout '%d'", *timeout)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeoutduration)
	defer cancel()

	var result []byte
	select {
	case <-ctx.Done():
		close(quitC)
	case result = <-resultC:
	}

	// output results
	rounds := binary.BigEndian.Uint64(data[len(data)-8:])
	if result == nil {
		t.Fatalf("Timeout after %d rounds\n", rounds)
		os.Exit(1)
	}
	t.Logf("Found %x after %d rounds\n", result, rounds)
}
