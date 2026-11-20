// PKGPATH: gno.land/r/zrealm_s3u_fresh_relay
// S3u: a stranger relays the victim's own stored function into its gate; the
// fresh nested object it writes is already part of stored state.
package zrealm_s3u_fresh_relay

import "gno.land/r/tests/vm/s3uprobe"

func main(cur realm) {
	s3uprobe.AddBox(cross(cur), s3uprobe.Stamp)
	println("owner:", s3uprobe.BoxOwner0())
}

// Error:
// gno.land/r/tests/vm/s3uprobe: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/s3uprobe)
