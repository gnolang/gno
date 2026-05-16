// PKGPATH: gno.land/r/demo/cross2stalerlm
package cross2stalerlm

// Runtime IsCurrent-strict catches a stale rlm: thread a realm value
// through cross-calls as a normal (non-cur) parameter, then try to
// use it via cross2 from a deeper frame. The original outer realm is
// no longer the topmost crossing frame's Cur — IsCurrent-strict
// rejects.
//
// Flow:
//   main(cur=M) -> A(cur=A1, outerCur=M) -> B(cur=B, outerCur=M)
//                  -> A(cur=A2, outerCur=M)
// At the second A invocation, the topmost crossing frame's Cur is
// A2 (freshly minted). outerCur is M (main's cur, no longer
// topmost). cross2(outerCur) panics on the runtime IsCurrent-strict
// check.

func callee(cur realm) string {
	return cur.PkgPath()
}

func A(cur realm, outerCur realm, depth int) {
	if depth > 0 {
		// Second invocation: outerCur is now stale (M, not A2).
		println(callee(cross2(outerCur)))
		return
	}
	// First invocation: outerCur is M, A's own cur is A1.
	// Pass outerCur through B and back into A.
	B(cross2(cur), outerCur)
}

func B(cur realm, outerCur realm) {
	A(cross2(cur), outerCur, 1)
}

func main(cur realm) {
	A(cross2(cur), cur, 0)
}

// Error:
// cross2: rlm is not the current cur (stale capture or sibling frame)
