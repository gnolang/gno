// PKGPATH: gno.land/r/crossrealm
package crossrealm

// A typed-nil conversion makes the cur-call's callee a *ConstExpr at
// preprocess time, so tryEvalStatic takes its early return. The static
// same-realm check must not panic on it (there is no concrete func to
// inspect); the call fails at runtime instead.

type F func(cur realm)

func main(cur realm) {
	F(nil)(cur)
}

// Error:
// runtime error: call of nil function
