package evm

type Instructions struct {
	pc int
	instructions []byte
	length int
}

func (i *Instructions) Reset(l []byte) {
	i.pc = 0
	i.instructions = l
	i.length = len(l)
}

func (i *Instructions) Increment(more ...int) {
	x := 1
	if len(more) > 0 {
		i.pc = more[0]
	}
	if (i.pc + x) >= i.length {
		i.pc = -1 // out of bounds
		return
	}
	i.pc += x
}

func (i Instructions) Current() byte {
	if i.Ill() {
		return 0
	}
	return i.instructions[i.pc]
}

func (i Instructions) Ill() bool {
	return i.pc < 0
}