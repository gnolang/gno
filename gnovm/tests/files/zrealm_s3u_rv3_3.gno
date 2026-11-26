// PKGPATH: gno.land/r/zrealm_s3u_rv3_3
// init calls the realm's own literal stored in a package var: a stored function
// value is not called by name, so it cannot write the realm's package var.
package zrealm_s3u_rv3_3

var G = "victim"

var W = func() { G = "stamped" }

func init() { W() }

func main(cur realm) {
	println("G:", G)
}

// Error:
// gno.land/r/zrealm_s3u_rv3_3's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_s3u_rv3_3) package{})).G
