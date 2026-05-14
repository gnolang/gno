// PKGPATH: gno.land/r/demo/cross2closurecap
package cross2closurecap

// cross2(rlm) — rlm cannot be passed as a closure capture. Same rule
// as bare `cur`: if cur could leak through a func lit, an attacker
// could stash it and use it from a sibling frame.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	// Try to capture cur in a closure and pass it through cross2.
	// Preprocessor rejects this at compile time.
	closure := func() {
		println(callee(cross2(cur)))
	}
	closure()
}

// Error:
// gno.land/r/demo/cross2closurecap/zrealm_cross2_closurecap.gno:16:11-30: cross2(rlm) — rlm cannot be passed as a closure capture, but found func func(){ (const (println func(...interface {})))(callee<VPBlock(4,0)>((const (cross2 func(.uverse.realm) .uverse.realm))(cur<VPBlock(2,0)>))) }
