package evm

type Instructions struct {
	rip          int
	instructions []byte
	length       int
}

func (i *Instructions) Reset(l []byte) {
	i.rip = 0
	i.instructions = l
	i.length = len(l)
}

func (i *Instructions) Increment(more ...int) *Instructions {
	x := 1
	if len(more) > 0 {
		i.rip = more[0]
	}
	if (i.rip + x) >= i.length {
		i.rip = -1 // out of bounds
		return i
	}
	i.rip += x
	return i
}

func (i Instructions) Current() byte {
	if i.Ill() {
		return 0
	}
	return i.instructions[i.rip]
}

func (i Instructions) Ill() bool {
	return i.rip < 0
}

// read n bytes and jump to rip + n
// TODO: make it uint256
func (i *Instructions) Read(n int) uint64 {
	if i.rip < 0 {
		return 0
	}

	var res uint64 = 0
	for n > 0 {
		res <<= 8
		res |= uint64(i.Current())
		i.rip++
		if i.rip >= i.length {
			i.rip = -1 // out of bounds
			return 0
		}
		n--
	}
	return res
}
