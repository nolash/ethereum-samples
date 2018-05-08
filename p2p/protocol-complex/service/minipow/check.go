package minipow

import (
	"bytes"
	"crypto/sha1"
)

// data does NOT include nonce here
func Check(hash []byte, data []byte, nonce []byte) bool {
	h := sha1.New()
	h.Write(data)
	h.Write(nonce)
	return bytes.Equal(hash, h.Sum(nil))
}
