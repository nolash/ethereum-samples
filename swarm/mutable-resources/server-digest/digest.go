//TODO: add signing
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/storage/mru"
	"github.com/ethereum/go-ethereum/swarm/storage/mru/lookup"
)

var (
	topicFlag   *string = flag.String("topic", "", "topic (optional)")
	nameFlag    *string = flag.String("name", "", "(optional, not implemented")
	levelFlag   *int    = flag.Int("level", 25, "epoch level (default 25)")
	timeFlag    *int    = flag.Int("time", 0, "time (default 0")
	userFlag    *string = flag.String("user", "", "user address (in hex)")
	signFlag    *bool   = flag.Bool("s", false, "sign digest (optional, not implemented")
	hexFlag     *bool   = flag.Bool("h", false, "input is hex encoded")
	verboseFlag *bool   = flag.Bool("v", false, "print debug output")

	user       common.Address
	topic      mru.Topic
	epochLevel uint8
	epochTime  uint64
	inFile     *os.File
)

func croak(r int, s string) {
	fmt.Fprintln(os.Stderr, "Error: ", s)
	if r == 1 {
		fmt.Println("\nUsage: digest.go --user <user> [options] [file]\n\tIf file is omitted, data will be read from stdin\n")
		flag.PrintDefaults()
	}
	os.Exit(r)
}

func init() {
	var err error
	flag.Parse()

	if *verboseFlag {
		log.Root().SetHandler(log.CallerFileHandler(log.LvlFilterHandler(6, log.StreamHandler(os.Stderr, log.TerminalFormat(false)))))
	}

	userBytes, err := hexutil.Decode(*userFlag)
	if err != nil {
		croak(1, fmt.Sprintf("Invalid user address '%s'", *userFlag))
	}
	user = common.BytesToAddress(userBytes)
	if *topicFlag != "" {
		topicBytes, err := hexutil.Decode(*topicFlag)
		if err != nil {
			croak(1, fmt.Sprintf("Invalid topic '%s'", *topicFlag))
		} else if len(topic) > mru.TopicLength {
			croak(1, fmt.Sprintf("Topic too long, max %d", mru.TopicLength))
		}
		topic, err = mru.NewTopic("", topicBytes)
		if err != nil {
			croak(1, fmt.Sprintf("Topic encode fail: %v", err))
		}
	}

	// without arg read from stdin
	if len(flag.Args()) == 0 {
		inFile = os.Stdin
	} else {
		inFile, err = os.Open(flag.Args()[0])
		if err != nil {
			croak(2, fmt.Sprintf("Could not open file '%s': %v", flag.Args()[0], err))
		}

	}
	epochLevel = uint8(*levelFlag)
	epochTime = uint64(*timeFlag)
}

func main() {
	r := mru.NewFirstRequest(topic)
	r.User = user
	//	r.ResourceUpdate.UpdateLookup.Epoch = lookup.Epoch{
	//		Level: epochLevel,
	//		Time:  epochTime,
	//	}
	r.Epoch = lookup.Epoch{
		Level: epochLevel,
		Time:  epochTime,
	}

	// TODO: input length check
	data := make([]byte, 8192)
	c, err := inFile.Read(data)
	if err != nil {
		croak(3, fmt.Sprintf("Data read fail: %v", err))
	}

	if *hexFlag {
		realData, err := hexutil.Decode(string(data[:c]))
		if err != nil {
			croak(3, fmt.Sprintf("Data hex decode fail '%v': %v", string(data), err))
		}
		data = realData
		c = len(realData)
	}
	r.SetData(data[:c])

	digest, err := r.GetDigest()
	if err != nil {
		croak(4, fmt.Sprintf("Digest generation fail: %v", err))
	}
	fmt.Println(hexutil.Encode(digest.Bytes()))
}
