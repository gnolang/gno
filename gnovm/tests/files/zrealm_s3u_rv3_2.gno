// PKGPATH: gno.land/r/zrealm_s3u_rv3_2
// R2a-save rv3: the realm's own stored writer (R2a) writes a fresh entry; the realm stores a copy of it by value
package zrealm_s3u_rv3_2

type Entry struct{ Owner string }

var W = func(e *Entry) { e.Owner = "stamped" }

var entries []Entry

func main(cur realm) {
	e := &Entry{Owner: "victim"}
	W(e)
	entries = append(entries, *e)
	println("owner:", entries[0].Owner)
}

// Output:
// owner: stamped
