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

// Both layers, which is the point: the gate refuses the deploy, and a stored
// package carrying this shape is still caught at call time by doOpCall.

// Error:
// crossing function gno.land/r/demo/curreassigndef.probe was entered without cross(), so it takes its caller's identity — but the value passed is not the caller's own cur (a stale capture from another frame, a sibling frame, or a rebound cur). Pass the caller's own `cur` unchanged, or use cross(cur) to enter as this realm

// TypeCheckError:
// gno.land/r/demo/curreassigndef/zrealm_cur_reassign_define.gno:15:2: cannot reassign the crossing `cur` parameter: it names the realm this frame is executing as, and that binding is fixed for the life of the frame
