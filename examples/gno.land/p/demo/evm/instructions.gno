package evm

import (
	"gno.land/p/demo/uint256"
)

type instructions struct {
	rip          int
	instructions []byte
	length       int
	err 		 error
}

func (i *instructions) Reset(l []byte) {
	i.rip = 0
	i.instructions = l
	i.length = len(l)
	i.err = nil
}

func (i *instructions) Increment(more ...int) *instructions {
	if i.Ill() {
		return i
	}
	x := 1
	if len(more) > 0 {
		x = more[0]
	}
	if ((i.rip + x) >= i.length) || ((i.rip + x) < 0) {
		i.err = outOfBounds
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
	return i.err != nil
}

func (i instructions) Error() error {
	return i.err
}

// read n bytes and jump to rip + n
func (i *instructions) Read(n int) (*uint256.Uint, error) {
	if i.Ill() {
		return uint256.Zero(), i.err
	}
	if i.rip+n > i.length {
		return uint256.Zero(), outOfBounds
	}
	x := uint256.Zero().SetBytes(i.instructions[i.rip : i.rip+n])
	i.Increment(n-1)
	return x, nil
}

func (i instructions) At(n int) byte {
	if i.Ill() || n >= i.length || n < 0 {
		return 0
	}
	return i.instructions[n]
}
