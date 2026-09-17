// PKGPATH: gno.land/r/demo/curreassign
package curreassign

// A crossing function's `cur` (or any realm-typed variable) may not be
// reassigned. Rebinding `cur = cur.Previous()` and forwarding it into a
// crossing helper entered without cross() would install a forged realm
// identity as that helper's frame Cur, satisfying rlm.IsCurrent() and
// cross(rlm) for an identity that is not the caller's — the laundering
// primitive behind the grc20reg / chain-banker drains. Reject it at
// preprocess.

func steal(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	cur = cur.Previous()
	steal(cur)
}

// Error:
// gno.land/r/demo/curreassign/zrealm_cur_reassign.gno:17:2-22: cannot reassign the crossing `cur` parameter: it is a caller-bound capability and rebinding it would forge the current-realm identity
