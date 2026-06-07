// Smoke test for the chain/runtime/unsafe package: all four symbols
// must resolve, produce values shaped like their chain/runtime/banker
// equivalents, and have OriginSend wired through the new X_originSend
// binding (was at chain/banker.X_originSend pre-move).

package main

import (
	"chain/runtime/unsafe"
)

func main() {
	caller := unsafe.OriginCaller()
	println("caller:", caller)

	cur := unsafe.CurrentRealm()
	println("cur:", cur.String())

	prev := unsafe.PreviousRealm()
	println("prev:", prev.String())

	send := unsafe.OriginSend()
	println("send len:", len(send))
}

// Output:
// caller: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// cur: CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
// prev: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// send len: 0
