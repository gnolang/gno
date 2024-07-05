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

func (f Fifo) Size() int {
	return len(f)
}

func (f Fifo) TotalAmount() int64 {
	var total int64
	for _, entry := range f {
		total += entry.Amount
	}
	return total
}
