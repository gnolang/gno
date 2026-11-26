// PKGPATH: gno.land/r/zrealm_s3u_rv4_1
// R2a-save rv4: another realm's non-crossing helper writes a caller's fresh *string; caller stores the pointer (false-refusal check)
package zrealm_s3u_rv4_1

import "gno.land/r/tests/vm/s3urv"

var ptrs []*string

func main(cur realm) {
	s := "v"
	p := &s
	s3urv.FillStr(p)
	ptrs = append(ptrs, p)
	println("owner:", *ptrs[0])
}

// Error:
// gno.land/r/zrealm_s3u_rv4_1: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/zrealm_s3u_rv4_1 and call it with cross)
