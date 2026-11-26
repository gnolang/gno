// PKGPATH: gno.land/r/zrealm_s3u_rv4_9
// R2a-save rv4: R2a stored writer: p[0] on a local array via pointer, *n++, append via pointer; pointers stored
package zrealm_s3u_rv4_9

var WA = func(p *[2]string) { p[0] = "stamped" }
var WI = func(p *int) { *p++; *p += 5 }
var WS = func(p *[]string) { (*p)[0] = "stamped" }

var pa []*[2]string
var pi []*int
var ps []*[]string

func main(cur realm) {
	var a [2]string
	n := 1
	s := make([]string, 1)
	WA(&a)
	pa = append(pa, &a)
	println("a:", pa[0][0])
	WI(&n)
	pi = append(pi, &n)
	println("n:", *pi[0])
	WS(&s)
	ps = append(ps, &s)
	println("s:", (*ps[0])[0])
}

// Error:
// gno.land/r/zrealm_s3u_rv4_9: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/zrealm_s3u_rv4_9 and call it with cross)
