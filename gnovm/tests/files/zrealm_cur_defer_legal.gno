// PKGPATH: gno.land/r/demo/curdeferok
package curdeferok

// Positive control: ordinary deferred crossing calls — plain, method, and
// wrapped in a closure, plus method-expression and method-value spellings —
// made from the frame whose identity they carry, must
// keep working. This guards against the defer fix over-rejecting; it passes
// with and without the fix by design, because IsCurrent() already answered
// true through the deferring frame. What it would catch is a fix that refused
// deferred crossing calls outright instead of inheriting for them.

type S struct{}

func (s *S) mdefer(cur realm) {
	println("method defer IsCurrent:", cur.IsCurrent())
}

func target(cur realm) {
	println("defer IsCurrent:", cur.IsCurrent())
	println("defer PkgPath:", cur.PkgPath())
}

func main(cur realm) {
	s := &S{}
	f := (*S).mdefer // method expression: receiver as an argument
	g := s.mdefer    // method value: receiver bound into the closure
	defer f(s, cur)
	defer g(cur)
	defer s.mdefer(cur)
	defer target(cur)
	defer func() { target(cur) }()
	println("body done")
}

// Output:
// body done
// defer IsCurrent: true
// defer PkgPath: gno.land/r/demo/curdeferok
// defer IsCurrent: true
// defer PkgPath: gno.land/r/demo/curdeferok
// method defer IsCurrent: true
// method defer IsCurrent: true
// method defer IsCurrent: true
