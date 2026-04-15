// PKGPATH: gno.land/r/test
// MAXALLOC: 100000
package test

import (
	"runtime"
	"strings"
)

var s string

func Assign() {
	s = strings.Repeat("a", 8000)
}

func main() {
	Assign()
	s1 := s
	println("===before GC, MemStats: ", runtime.MemStats())
	// The string underlying bytes should only be counted once
	// in GC recount process, otherwise GC will broken.
	runtime.GC()
	println("===after GC, MemStats: ", runtime.MemStats())
}

// Output:
// ===before GC, MemStats:  Allocator{maxBytes:100000, bytes:98392}
// ===after GC, MemStats:  Allocator{maxBytes:100000, bytes:14744}

// TypeCheckError:
// gno.land/r/test/alloc_13.gno:18:2: declared and not used: s1
