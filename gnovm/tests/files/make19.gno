package main

// An oversized make([]T, n) must produce a Gno-recoverable panic when the
// allocation-size accounting overflows int64, not an unrecoverable Go-host
// panic. The length is the only cheap, user-controlled input that can reach
// the overflow (append's size is bounded by already-allocated slices). Both
// array allocator entry points are covered: AllocateListArray ([]int) and
// AllocateDataArray ([]byte).

func shouldPanic(f func()) {
	defer func() {
		r := recover()
		if r == nil {
			panic("not panicking")
		}
		println("recovered:", r)
	}()
	f()
}

func main() {
	length := int(^uint(0) >> 1) // max int
	shouldPanic(func() { _ = make([]int, length) })  // AllocateListArray
	shouldPanic(func() { _ = make([]byte, length) }) // AllocateDataArray
	println("ok")
}

// Output:
// recovered: runtime error: makeslice: len out of range
// recovered: runtime error: makeslice: len out of range
// ok
