// PKGPATH: gno.land/r/zrealm_s3u_rv4_3
// R2a-save rv4: non-crossing helper does *p++ and *p += 5 on a caller's local int; caller stores the pointer
package zrealm_s3u_rv4_3

import "gno.land/r/tests/vm/s3urv"

var ptrs []*int

func main(cur realm) {
	n := 1
	p := &n
	s3urv.IncInt(p)
	s3urv.AddInt(p)
	ptrs = append(ptrs, p)
	println("n:", *ptrs[0])
}

// Error:
// gno.land/r/zrealm_s3u_rv4_3: an object written by read-only code would be stored (do the write in a crossing function of gno.land/r/zrealm_s3u_rv4_3, called with cross)
