// MAXALLOC: 1600000
package main

import (
	"strings"

	cid "gno.land/p/omarsy/cid/v0"
)

// Oversized input costs no more than the longest CID.
func main() {
	// Only a 46-character "Qm" string reaches the quadratic base58btc
	// decoder as a CIDv0; a longer one fails the multibase length cap.
	_, err := cid.Parse("Qm" + strings.Repeat("z", 510))
	println(err)

	// DecodeFirst copies at most maxCIDLen bytes: MAXALLOC leaves room for
	// this 1 MiB buffer but not for a second copy of it.
	_, n, err := cid.DecodeFirst(make([]byte, 1<<20))
	println(n, err)
}

// Output:
// cid: invalid multibase string
// 0 cid: invalid CID
