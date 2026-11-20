// PKGPATH: gno.land/r/zrealm_s3u_fresh_pass
// S3u: a fresh entry already attached to the realm's stored state stays
// refused to a by-value callback.
package zrealm_s3u_fresh_pass

import "gno.land/r/tests/vm/s3uprobe"

func main(cur realm) {
	s3uprobe.AddPass(cross(cur), func(e *s3uprobe.Entry) { e.Owner = "attacker" })
	println("owner:", s3uprobe.Owner0())
}

// Error:
// gno.land/r/zrealm_s3u_fresh_pass's code was entered without crossing and is read-only (call its crossing function): e<VPBlock(1,0)>.Owner
