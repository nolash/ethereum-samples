package service

import (
	"context"
	"sync"

	"./minipow"
)

var (
	mu = sync.Mutex{}
)

type job struct {
	Data  []byte
	Hash  []byte
	Nonce []byte
}

func doJob(ctx context.Context, rawData []byte, difficulty uint8) (*job, error) {
	resultC := make(chan []byte)
	quitC := make(chan struct{})

	workData := make([]byte, len(rawData)+8)
	copy(workData, rawData)

	go minipow.Mine(workData, int(difficulty), resultC, quitC, nil)

	var r []byte
	select {
	case <-ctx.Done():
		quitC <- struct{}{}
		return nil, ctx.Err()
	case r = <-resultC:
	}

	j := &job{
		Data:  rawData,
		Nonce: workData[len(workData)-8:],
		Hash:  r,
	}
	return j, nil
}

func checkJob(hash []byte, data []byte, nonce []byte) bool {
	if hash == nil || data == nil || nonce == nil {
		return false
	}
	return minipow.Check(hash, data, nonce)
}
