// PKGPATH: gno.land/r/zrealm_s3u_rv4_8
// R2a-save rv4: range assignment into *p by a non-crossing helper; caller stores p
package zrealm_s3u_rv4_8

import "gno.land/r/tests/vm/s3urv"

var ptrs []*string

func main(cur realm) {
	s := "v"
	p := &s
	s3urv.RangeInto(p, []string{"a", "ranged"})
	ptrs = append(ptrs, p)
	println("s:", *ptrs[0])
}

// Error:
// gno.land/r/zrealm_s3u_rv4_8: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/zrealm_s3u_rv4_8 and call it with cross)
