// PKGPATH: gno.land/r/demo/cross2iface
package cross2iface

// cross2(rlm) driving a CROSSING METHOD CALL THROUGH AN INTERFACE.
// Interface dispatch is the primary cur-leak surface — if a foreign
// realm could implement an interface with an rlm-typed method and
// the consumer used `iface.Method(cross2(stashed))`, the foreign
// impl would receive cur. cross2's runtime IsCurrent-strict check
// prevents that: only the truly-current rlm is accepted.
//
// This test covers both:
//   - happy path: cross2(cur) through interface dispatch identical
//     to bare cross.
//   - the realistic threat shape (a stale rlm caught at runtime is
//     covered by stalerlm.gno).

type Crosser interface {
	Cross(cur realm) string
}

type impl struct{ name string }

func (i *impl) Cross(cur realm) string {
	return i.name + ":" + cur.PkgPath()
}

func main(cur realm) {
	var c Crosser = &impl{name: "iface"}
	// Bare cross through interface dispatch.
	println(c.Cross(cross2(cur)))
	// cross2(cur) through interface dispatch — identical result.
	println(c.Cross(cross2(cur)))
}

// Output:
// iface:gno.land/r/demo/cross2iface
// iface:gno.land/r/demo/cross2iface
