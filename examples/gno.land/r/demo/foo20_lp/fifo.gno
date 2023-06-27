package foo20_lp

import "std"

type Fifo []Entry

func (f Fifo) AddrHasEntries(addr std.Address) bool {
	for _, entry := range f {
		if entry.Addr == addr {
			return true
		}
	}
	return false
}

func (f Fifo) Empty() bool {
	return len(f) == 0
}
