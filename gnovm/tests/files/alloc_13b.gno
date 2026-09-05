// MAXALLOC: 100000
// PKGPATH: gno.land/r/test

// Companion to alloc_13/alloc_13a: the SOURCE dies, its copy and a slice
// survive. The backing must stay counted exactly once after GC.
//
//	s  := base + base   // runtime concat: minted through NewString
//	s1 := s             // full copy, shares the backing
//	s2 := s[1:11]       // slice, shares the backing (GetSlice)
//	s, base = "", ""    // source (and the literal ref) die
//
// After GC, exactly one full backing (1024B) must remain in the count,
// reached only through s1/s2:
//   - counting it twice (once per surviving value) is the pre-#4885
//     double-count bug;
//   - counting it zero times (source gone, so "nobody owns it") is the
//     ref-flag undercount that sinks a #5082-style per-value design.
package test

import (
	"runtime"
)

func main() {
	base := `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
gggggggggggggggggggggggggg` // 512 bytes
	s := base + base // 1024B fresh backing: charged and tracked
	s1 := s          // copy shares the backing
	s2 := s[1:11]    // slice shares the backing
	s = ""           // source dies
	base = ""        // literal ref dies too
	runtime.GC()
	println("after GC: ", runtime.MemStats())
	println(len(s1), len(s2))
}

// Output:
// after GC:  Allocator{maxBytes:100000, bytes:8920}
// 1024 10
