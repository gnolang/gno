// PKGPATH: gno.land/r/demo/curdefer
package curdefer

// A DEFERRED crossing call must be treated like any other. doOpReturnCallDefers
// duplicates doOpCall's block setup — its own comment says so — and the copy
// left out the part that installs the frame's identity and checks it. So this
// ran happily, while the identical call without `defer` was refused.
//
// The shape is the same stale capture as zrealm_cur_backstop.gno: cb captured
// main's cur but runs inside exec's frame. Deferring it changes nothing about
// whose identity is being taken, only which dispatch path carries it.

func target(cur realm) {
	println("target:", cur.PkgPath())
}

func exec(cur realm, cb func()) {
	cb()
}

func main(cur realm) {
	cb := func() { defer target(cur) }
	exec(cross(cur), cb)
}

// Error:
// crossing function gno.land/r/demo/curdefer.target was entered without cross(), so it takes its caller's identity — but the value passed is not the caller's own cur (a stale capture from another frame, a sibling frame, or a rebound cur). Pass the caller's own `cur` unchanged, or use cross(cur) to enter as this realm
