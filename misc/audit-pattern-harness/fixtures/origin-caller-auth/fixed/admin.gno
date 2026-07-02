package admin

var owner = address("g1qz8e0fz3y0pl9y4dq9d7c5dwnyu6qf04hs7z0a")
var paused bool

func SetPaused(cur realm, next bool) {
	// Fixed: authorize the immediate caller, not the transaction origin.
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	if cur.Previous().Address() != owner {
		panic("owner only")
	}
	paused = next
}

func Paused() bool {
	return paused
}
