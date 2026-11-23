// PKGPATH: gno.land/r/zrealm_s3u_rv4_4
// R2a-save rv4: R2a stored writer sets p.Owner on a local struct through &x; caller stores &x
package zrealm_s3u_rv4_4

type Entry struct{ Owner string }

var W = func(p *Entry) { p.Owner = "stamped" }

var ptrs []*Entry

func main(cur realm) {
	var x Entry
	W(&x)
	ptrs = append(ptrs, &x)
	println("owner:", ptrs[0].Owner)
}

// Error:
// gno.land/r/zrealm_s3u_rv4_4: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/zrealm_s3u_rv4_4)
