package common

import (
	"os"
)

const (
	p2pPort       = 30100
	ipcName       = "demo.ipc"
	datadirPrefix = ".data_"
)

var (
	basePath, _ = os.Getwd()
)
