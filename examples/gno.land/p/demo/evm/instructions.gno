package evm

import (
	"gno.land/p/onbloc/uint256"
)

type instructions struct {
	rip          int
	instructions []byte
	length       int
	err          error
	isStopped    bool
}

func (i *instructions) reset(l []byte) {
	i.rip = 0
	i.instructions = l
	i.length = len(l)
	i.err = nil
}

func (i *instructions) increment(more ...int) *instructions {
	if i.ill() {
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

func (i instructions) current() byte {
	if i.ill() {
		return 0
	}
	return i.instructions[i.rip]
}

func (i instructions) ill() bool {
	return i.err != nil
}

func (i instructions) error() error {
	return i.err
}

func (i instructions) isDone() bool {
	return i.ill() || i.isStopped
}

func (i *instructions) stop() {
	i.isStopped = true
}

// read n bytes and jump to rip + n
func (i *instructions) read(n int) (*uint256.Uint, error) {
	if i.ill() {
		return uint256.Zero(), i.err
	}
	if i.rip+n > i.length {
		return uint256.Zero(), outOfBounds
	}
	x := uint256.Zero().SetBytes(i.instructions[i.rip : i.rip+n])
	i.increment(n - 1)
	return x, nil
}

func (i instructions) at(n int) byte {
	if i.ill() || n >= i.length || n < 0 {
		return 0
	}
	return i.instructions[n]
}

func (i instructions) endOfCode() bool {
	return i.rip >= i.length
}
