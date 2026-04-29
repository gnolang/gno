// MAXALLOC: 50000
// PKGPATH: gno.land/r/test

// Companion to alloc_13: larger string (1052 bytes) so the byte
// recounting impact is clearly visible in the allocator numbers.
//
// s1 := s shares the backing. In the GC recount:
//   - s counts header (48B) + backing (1052B) = 1100B
//   - s1 counts header (48B) only, because CountStringBytes returned
//     true for s and false for s1 (same backing, same cycle)
//
// If GC double-counted:
//
//	after ≈ 11956 - <dead> + 48 + 1052 ≈ 9448
//	(the extra 1052B for s1's backing would push after above 9k)
//
// Observed: after = 8396, confirming single-count dedup.
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
// ===before GC, MemStats:  Allocator{maxBytes:50000, bytes:11956}
// ===after GC, MemStats:  Allocator{maxBytes:50000, bytes:8396}
