// PKGPATH: gno.land/r/zrealm_s3u_rv2_9
// R2a-save: the realm's own stored writer (by value) on a fresh entry, stored after it returns
package zrealm_s3u_rv2_9

type Entry struct{ Owner string }

var W = func(e *Entry) { e.Owner = "stamped" }

var entries []*Entry

func main(cur realm) {
	e := &Entry{Owner: "victim"}
	W(e)
	entries = append(entries, e)
	println("owner:", entries[0].Owner)
}

// Error:
// gno.land/r/zrealm_s3u_rv2_9: an object written by read-only code would be stored (do the write in a crossing function of gno.land/r/zrealm_s3u_rv2_9, called with cross)
