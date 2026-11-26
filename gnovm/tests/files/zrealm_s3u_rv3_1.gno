// PKGPATH: gno.land/r/zrealm_s3u_rv3_1
// R2a-save rv3: the realm's own stored writer (R2a) writes a fresh *string (unreal heap item holding a primitive), then the pointer is stored
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
// gno.land/r/zrealm_s3u_rv3_1: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/zrealm_s3u_rv3_1 and call it with cross)
