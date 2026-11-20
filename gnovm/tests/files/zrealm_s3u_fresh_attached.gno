// PKGPATH: gno.land/r/zrealm_s3u_fresh_attached
// S3u: a fresh entry already attached to the realm's stored state stays
// refused to a by-value callback.
package zrealm_s3u_fresh_attached

import "gno.land/r/tests/vm/s3uprobe"

func main(cur realm) {
	s3uprobe.Add(cross(cur), func() {
		s3uprobe.Last().Owner = "attacker"
	})
	println("owner:", s3uprobe.Owner0())
}

// Error:
// gno.land/r/zrealm_s3u_fresh_attached's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/s3uprobe) package{})).Last().Owner
