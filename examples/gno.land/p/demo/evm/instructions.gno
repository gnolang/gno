package evm

import (
	"fmt"

	"gno.land/p/demo/evm/uint256"
)

type instructions struct {
	rip          int
	instructions []byte
	length       int
}

func (i *instructions) Reset(l []byte) {
	i.rip = 0
	i.instructions = l
	i.length = len(l)
}

func (i *instructions) Increment(more ...int) *instructions {
	x := 1
	if len(more) > 0 {
		x = more[0]
	}
	if (i.rip + x) >= i.length {
		i.rip = -1 // out of bounds
		return i
	}
	i.rip += x
	return i
}

func (i instructions) Current() byte {
	if i.Ill() {
		return 0
	}
	return i.instructions[i.rip]
}

func (i instructions) Ill() bool {
	return i.rip < 0
}

// read n bytes and jump to rip + n
// TODO: make it uint256
func (i *instructions) Read(n int) *uint256.Uint {
	if i.rip < 0 {
		return uint256.New()
	}
	x := uint256.NewFromBytes(i.instructions[i.rip:i.rip+n])
	i.Increment(n)
	return x
}
