// MAXALLOC: 50000
// PKGPATH: gno.land/r/test

// VM panic text is untracked until recover hands it to user code; once
// held, the GC recount must include its bytes.
package test

import "runtime"

var (
	held any
	arr  []int
)

func grab(i int) {
	defer func() { held = recover() }()
	_ = arr[i] // runtime error: index out of range [7] with length 0
}

func main() {
	grab(7)
	runtime.GC()
	println(len(held.(error).Error()))
	println(runtime.MemStats())
}

// Output:
// 46
// Allocator{maxBytes:50000, bytes:12332}
