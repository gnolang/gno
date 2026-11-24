// PKGPATH: gno.land/r/demo/curmethok
package curmethok

// Positive control for the method wiring: an ordinary crossing method, called
// honestly, keeps working and sees itself as the current realm.

type S struct{}

func (s *S) inner(cur realm) {
	println("inner IsCurrent:", cur.IsCurrent())
	println("inner PkgPath:", cur.PkgPath())
}

func (s *S) outer(cur realm) {
	s.inner(cur)
}

// Crossing methods reached through an INTERFACE. Worth pinning because a lazy
// interface bind reports IsCrossing() false until it is resolved, while the
// resolved method is genuinely crossing — so this path only started installing
// an identity when methods began entering the check at all.
type I interface{ inner(cur realm) }

func viaIface(cur realm, i I) { i.inner(cur) }

func main(cur realm) {
	s := &S{}
	s.outer(cur)
	s.outer(cross(cur))
	viaIface(cur, s)
	// and an ordinary crossing call afterwards, to show the frame flags the
	// dispatch above leaves behind still resolve correctly.
	s.outer(cur)
}

// Output:
// inner IsCurrent: true
// inner PkgPath: gno.land/r/demo/curmethok
// inner IsCurrent: true
// inner PkgPath: gno.land/r/demo/curmethok
// inner IsCurrent: true
// inner PkgPath: gno.land/r/demo/curmethok
// inner IsCurrent: true
// inner PkgPath: gno.land/r/demo/curmethok
