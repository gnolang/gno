// PKGPATH: gno.land/r/test/deferfixnil
package deferfixnil

// Test 9: nil-receiver defer-method. The method must handle nil
// gracefully. PushFrameCall hits the obj == nil branch and returns
// without an implicit switch. Both pre-fix and post-fix should pass;
// guards against the fix breaking nil handling.

type S struct{}

func (s *S) Greet() {
	if s == nil {
		println("nil greeted")
		return
	}
	println("non-nil greeted")
}

func main(cur realm) {
	var s *S = nil
	defer s.Greet()
	println("body done")
}

// Output:
// body done
// nil greeted
