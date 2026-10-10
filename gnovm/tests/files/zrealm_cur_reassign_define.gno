// PKGPATH: gno.land/r/demo/curreassigndef
package curreassigndef

// The `:=` spelling of a rebind. A DEFINE whose name is already declared in
// the same block assigns that slot rather than declaring a new one, so
// `cur, x := ...` at the function body's own level writes the crossing
// parameter. The DEFINE carve-out in the write rule used to skip it, leaving
// only doOpCall's runtime check to notice at call time.

func probe(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	cur, x := cur.Previous(), 1
	_ = x
	probe(cur)
}

// Error:
// gno.land/r/demo/curreassigndef/zrealm_cur_reassign_define.gno:15:2-29: cannot reassign the crossing `cur` parameter: it names the realm this frame is executing as, and that binding is fixed for the life of the frame
