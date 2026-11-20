// PKGPATH: gno.land/r/zrealm_s3u_fresh_slot
// S3u: a fresh entry already attached to the realm's stored state stays
// refused to a by-value callback.
package zrealm_s3u_fresh_slot

import "gno.land/r/tests/vm/s3uprobe"

func main(cur realm) {
	s3uprobe.AddSlot(cross(cur), func(es []*s3uprobe.Entry) { es[0] = nil })
	println("owner:", s3uprobe.Owner0())
}

// Error:
// gno.land/r/zrealm_s3u_fresh_slot's code was entered without crossing and is read-only (call its crossing function): es<VPBlock(1,0)>[(const (0 int))]
