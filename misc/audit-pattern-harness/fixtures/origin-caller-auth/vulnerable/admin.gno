package admin

import "chain/runtime/unsafe"

var owner = address("g1qz8e0fz3y0pl9y4dq9d7c5dwnyu6qf04hs7z0a")
var paused bool

func SetPaused(cur realm, next bool) {
	// Vulnerable: OriginCaller() can hide an intermediate realm caller.
	if unsafe.OriginCaller() != owner {
		panic("owner only")
	}
	paused = next
}

func Paused() bool {
	return paused
}
