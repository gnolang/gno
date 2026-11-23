// PKGPATH: gno.land/r/zrealm_s3u_rv3_3
// R2a-save rv3: init calls the realm's own package-var func by value; it writes a package var of the fresh package
package zrealm_s3u_rv3_3

var G = "victim"

var W = func() { G = "stamped" }

func init() { W() }

func main(cur realm) {
	println("G:", G)
}

// Output:
// G: stamped
