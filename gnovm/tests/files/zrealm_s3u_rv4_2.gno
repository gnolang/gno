// PKGPATH: gno.land/r/zrealm_s3u_rv4_2
// S3d rv4: non-crossing helper writes p[0] of a caller's local array via pointer (write lands on the inner array, not the heap item); caller stores the pointer
package zrealm_s3u_rv4_2

import "gno.land/r/tests/vm/s3urv"

var ptrs []*[2]string

func main(cur realm) {
	var a [2]string
	p := &a
	s3urv.FillArr(p)
	ptrs = append(ptrs, p)
	println("owner:", ptrs[0][0])
}

// Error:
// gno.land/r/tests/vm/s3urv's code was entered without crossing and is read-only (call its crossing function): *(p<VPBlock(1,0)>)[(const (0 int))]
