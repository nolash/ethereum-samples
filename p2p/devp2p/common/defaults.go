package common

import (
	"os"
)

const (
	P2pPort       = 30100
	IPCName       = "demo.ipc"
	DatadirPrefix = ".data_"
)

var (
	basePath, _ = os.Getwd()
)
