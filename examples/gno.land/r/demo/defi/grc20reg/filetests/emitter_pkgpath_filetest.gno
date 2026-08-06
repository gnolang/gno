// PKGPATH: gno.land/r/demo/grc20emitpath

// The observable payoff of registry-owned emission: pkg_path.
//
// The VM stamps pkg_path from the package that calls chain.Emit, so a token that
// opted into the registry has every event labelled with the registry's path,
// and that label cannot be forged — a counterfeit emitter's events carry the
// counterfeit's own path. A token created with plain NewToken emits under
// grc20's path, so the two are trivially distinguishable in the event stream.
package grc20emitpath

import (
	"gno.land/p/demo/tokens/grc20"
	"gno.land/r/demo/defi/grc20reg"
)

func main(cur realm) {
	owned, ownedLedger := grc20.NewTokenWithEmitter("Owned", "OWN", 6, grc20reg.Emitter(), cur)
	legacy, legacyLedger := grc20.NewToken("Legacy", "LEG", 6, 0, cur)

	println("owned  id:", owned.ID())
	println("legacy id:", legacy.ID())

	ownedLedger.Mint(cur.Address(), 1)
	legacyLedger.Mint(cur.Address(), 1)
}

// Output:
// owned  id: gno.land/r/demo/grc20emitpath.OWN.1
// legacy id: gno.land/r/demo/grc20emitpath.LEG.0000000

// Events:
// [
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20emitpath.OWN.1"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1y4j5mkkjel0n43knx9jt82hq4ndwe47nnwcc70"
//       },
//       {
//         "key": "value",
//         "value": "1"
//       }
//     ],
//     "pkg_path": "gno.land/r/demo/defi/grc20reg"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20emitpath.LEG.0000000"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1y4j5mkkjel0n43knx9jt82hq4ndwe47nnwcc70"
//       },
//       {
//         "key": "value",
//         "value": "1"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   }
// ]
