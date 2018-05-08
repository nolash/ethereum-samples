package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/contracts/ens"
)

func main() {
	if len(os.Args) == 1 {
		return
	}
	fmt.Println(ens.EnsNode(os.Args[1]).Hex())
}
