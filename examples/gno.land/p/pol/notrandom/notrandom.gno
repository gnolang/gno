package notrandom

type Machine struct {
	seed uint64
}

func NewMachine(seed int64) *Machine {
	return &Machine{
		seed: uint64(seed),
	}
}

func (rm *Machine) xorshift64star() uint64 {
	rm.seed ^= rm.seed >> 12
	rm.seed ^= rm.seed << 25
	rm.seed ^= rm.seed >> 27
	return rm.seed * 2685821657736338717

	// const a = 1664525
    // const c = 1013904223
    // const m = 4294967296
    // rm.seed = (a*rm.seed + c) % m
	// return 1
}

func (rm *Machine) RandFloat() float64 {
	return float64(rm.xorshift64star()>>11) / (1 << 53)
}

func iAbs(x int) int {
	if (x < 0) {
		return -x
	} else {
		return x
	}
}

func (r *Machine) ChangeSeed(seed int64) {
	r.seed = uint64(seed)
}

func (r *Machine) RandInt(min, max int) int {
	return (iAbs(r.GenerateInt()) % ((max - min) + 1)) + min
}

func (rm *Machine) GenerateInt() int {
	return int(rm.xorshift64star())
}

func Hello() int {
	return 42
}
