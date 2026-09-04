// MAXALLOC: 50000
// PKGPATH: gno.land/r/test

// Companion to alloc_13: larger string (1052 bytes) so the byte
// recounting impact is clearly visible in the allocator numbers.
//
// s1 := s shares the backing. In the GC recount:
//   - s counts header (48B) + backing (1052B) = 1100B
//   - s1 counts header (48B) only, because the GC visitor had already
//     seen s's backing in this cycle (same backing, same cycle)
//
// If GC double-counted, s1's backing would add another 1052B to the
// recount (after ≈ 10024).
//
// Observed: after = 8972, confirming single-count dedup.
package test

import (
	"runtime"
)

var s string

func Assign() {
	s = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg
hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh
iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii
jjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjj
kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk
llllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllll
mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm`
}

func init() {
	Assign()
}

func main() {
	s1 := s // shares backing with s
	println("===before GC, MemStats: ", runtime.MemStats())
	runtime.GC()
	println("===after GC, MemStats: ", runtime.MemStats())
	_ = s1
}

// Output:
// ===before GC, MemStats:  Allocator{maxBytes:50000, bytes:11980}
// ===after GC, MemStats:  Allocator{maxBytes:50000, bytes:8972}
