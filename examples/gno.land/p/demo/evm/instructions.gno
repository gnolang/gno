package evm

type Instructions struct {
	rip int
	instructions []byte
	length int
}

func (i *Instructions) Reset(l []byte) {
	i.rip = 0
	i.instructions = l
	i.length = len(l)
}

func (i *Instructions) Increment(more ...int) {
	x := 1
	if len(more) > 0 {
		i.rip = more[0]
	}
	if (i.rip + x) >= i.length {
		i.rip = -1 // out of bounds
		return
	}
	i.rip += x
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