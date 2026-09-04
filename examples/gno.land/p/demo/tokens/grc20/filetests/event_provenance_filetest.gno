// PKGPATH: gno.land/r/demo/grc20provenance

// This filetest is the reference for how an indexer attributes a grc20 event to
// the realm that issued the token. It is the cross-realm counterpart to
// newtoken_event_filetest.gno, which covers the within-realm duplicate signal.
//
// Two fields matter, and neither is sufficient alone:
//
//   - pkg_path is stamped by the VM from the package whose code calls
//     chain.Emit. Every genuine grc20 event therefore carries
//     gno.land/p/demo/tokens/grc20 — a constant, so it identifies the library
//     but never the token.
//   - the "token" attribute carries Token.ID(), whose leading component is
//     rlm.PkgPath() captured under IsCurrent() in NewToken. A realm cannot
//     produce a Token whose id claims a different realm.
//
// Chained, they are sufficient: pkg_path proves the event came from grc20's
// code, and grc20's code only ever writes an IsCurrent-verified realm prefix
// into "token". So for any event bearing grc20's pkg_path, the realm prefix in
// "token" is unforgeable.
//
// The realm below is an impostor trying to pass its own transfers off as
// another realm's token, by both routes available to it, and failing at both.
package grc20provenance

import (
	"chain"

	"gno.land/p/demo/tokens/grc20"
)

// Stand-in for a token issued and registered by some other realm. Written as a
// literal rather than imported: a filetest under a /p/ package that imports an
// /r/ realm which itself imports that /p/ package forms a genesis dependency
// cycle (grc20 -> realm -> grc20), which breaks package loading chain-wide.
const canonical = "gno.land/r/demo/canonicaltoken.FOO.0000000"

func main(cur realm) {
	println("canonical id:", canonical)

	// Route 1: ask grc20 for a token using the canonical token's own
	// name/symbol/decimals/id. NewToken derives the id prefix from the live
	// crossing frame, so the id names *this* realm. No argument changes that.
	impostor, impostorLedger := grc20.NewToken("Foo", "FOO", 4, 0, cur)
	println("impostor id: ", impostor.ID())
	println("impostor claims canonical id:", impostor.ID() == canonical)

	// A genuine grc20 event from the impostor. It carries grc20's pkg_path —
	// identical to the canonical token's events — but the "token" prefix is
	// this realm, so an indexer attributes it here.
	impostorLedger.Mint(cur.Address(), 999_999)

	// Route 2: skip grc20 and hand-emit an event carrying the canonical id.
	// The attributes are fully attacker-chosen, but pkg_path is not: the VM
	// stamps this realm. An indexer that accepts Transfer only from grc20's
	// pkg_path drops this event.
	chain.Emit(
		"Transfer",
		"token", canonical,
		"from", "",
		"to", cur.Address().String(),
		"value", "999999",
	)
}

// Output:
// canonical id: gno.land/r/demo/canonicaltoken.FOO.0000000
// impostor id:  gno.land/r/demo/grc20provenance.FOO.0000000
// impostor claims canonical id: false

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20provenance.FOO.0000000"
//       },
//       {
//         "key": "name",
//         "value": "Foo"
//       },
//       {
//         "key": "symbol",
//         "value": "FOO"
//       },
//       {
//         "key": "decimals",
//         "value": "4"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20provenance.FOO.0000000"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1vz4y807zq7p3s63dmyng9g80l58nmuc3lgmfjs"
//       },
//       {
//         "key": "value",
//         "value": "999999"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/canonicaltoken.FOO.0000000"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1vz4y807zq7p3s63dmyng9g80l58nmuc3lgmfjs"
//       },
//       {
//         "key": "value",
//         "value": "999999"
//       }
//     ],
//     "pkg_path": "gno.land/r/demo/grc20provenance"
//   }
// ]
