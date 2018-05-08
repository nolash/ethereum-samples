package minipow

import (
	"crypto/sha1"
)

func Mine(data []byte, difficulty int, resultC chan<- []byte, quitC <-chan struct{}, debug func([]byte, []byte)) {
	h := sha1.New()

	datalen := len(data)
	hashsizeminusone := h.Size() - 1
	diffbytes := make([]byte, h.Size())
	result := make([]byte, h.Size())

	// generate the difficulty mask
	var register byte
	c := 0
OUTER_ONE:
	for i := hashsizeminusone; i > 0; i-- {
		register = 0x01
		for j := 1; j < 256; j *= 2 {
			diffbytes[i] |= register
			register <<= 1
			c++
			if c == difficulty {
				break OUTER_ONE
			}
		}
	}

	diffthreshold := hashsizeminusone - int(difficulty/8)

	// 256 bit number is pretty close to eternity
OUTER_TWO:
	for {
		// timeout handling
		select {
		case <-quitC:
			break OUTER_TWO
		default:
		}

		// increment the nonce
		for i := datalen - 1; i >= 0; i-- {
			data[i]++
			if data[i]|0x0 != 0 {
				break
			}
		}

		// hashhhhh
		h.Reset()
		h.Write(data)
		sum := h.Sum(nil)
		if debug != nil {
			debug(data, sum)
		}
		copy(result[:], sum)

		// and byte for byte with the difficulty mask
		// we only check the bytes we have to, though, until diffthreshold
		// if not 0, we failed miserably
		for i := hashsizeminusone; i >= diffthreshold; i-- {
			sum[i] &= diffbytes[i]
			if sum[i] != 0x0 {
				continue OUTER_TWO
			}
		}

		// there was rejoicing
		resultC <- result
		return
	}
	resultC <- nil
}
