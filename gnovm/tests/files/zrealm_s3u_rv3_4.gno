// PKGPATH: gno.land/r/zrealm_s3u_rv3_4
// S3d rv3: stored writer (S3) assigns a fresh slice through a pointer to a nil local slice; the pointer is then stored
package zrealm_s3u_rv3_4

var W = func(p *[]string) { *p = []string{"stamped"} }

var ptrs []*[]string

func main(cur realm) {
	var s []string
	p := &s
	W(p)
	ptrs = append(ptrs, p)
	println("owner:", (*ptrs[0])[0])
}

// Error:
// gno.land/r/zrealm_s3u_rv3_4: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/zrealm_s3u_rv3_4)
