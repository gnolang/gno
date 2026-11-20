// PKGPATH: gno.land/r/zrealm_s3u_rv3_1
// S3d rv3: the realm's own stored writer (S3) writes a fresh *string (unreal heap item holding a primitive), then the pointer is stored
package zrealm_s3u_rv3_1

var W = func(p *string) { *p = "stamped" }

var ptrs []*string

func main(cur realm) {
	s := "victim"
	p := &s
	W(p)
	ptrs = append(ptrs, p)
	println("owner:", *ptrs[0])
}

// Error:
// gno.land/r/zrealm_s3u_rv3_1: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/zrealm_s3u_rv3_1)
