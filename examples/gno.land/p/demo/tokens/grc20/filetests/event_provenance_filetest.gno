// PKGPATH: gno.land/r/demo/grc20provenance

// This filetest is the reference for how an indexer attributes a grc20 event to
// the realm that issued the token.
//
// Three fields matter, and none is sufficient alone:
//
//   - pkg_path is stamped by the VM from the package whose code calls
//     chain.Emit. Every genuine grc20 event therefore carries
//     gno.land/p/demo/tokens/grc20 — a constant, so it identifies the library
//     but never the token.
//   - the "token" attribute carries Token.TokenPath(), rlmPath.symbol as
//     captured in NewToken from the live crossing frame under IsCurrent(). A
//     realm cannot fabricate one naming a different realm. It is readable, but
//     it does not separate two tokens one realm issued under one symbol.
//   - the "id" attribute carries Token.ID(), the address derived from the token
//     object's own VM object ID. The realm chooses no part of it, and it is
//     unique per object — that is what separates those two tokens. It is empty
//     until the token has been persisted.
//
// Chained, they are sufficient: pkg_path proves the event came from grc20's
// code, and grc20's code only ever writes an IsCurrent-verified path into
// "token" and a VM-derived address into "id".
//
// The impostor's token is created in init because an object has no id until it
// is first persisted, which happens as the creating frame returns. Created in
// main it would read "", which is the unstamped form, not a claim on anything.
package grc20provenance

import (
	"chain"

	"gno.land/p/demo/tokens/grc20"
)

// Stand-in for a token issued and registered by some other realm. Written as a
// literal rather than imported: a filetest under a /p/ package that imports an
// /r/ realm which itself imports that /p/ package forms a genesis dependency
// cycle (grc20 -> realm -> grc20), which breaks package loading chain-wide.
const canonical = "gno.land/r/demo/canonicaltoken.FOO"

var (
	impostor       *grc20.Token
	impostorLedger *grc20.PrivateLedger
)

func init(cur realm) {
	// Route 1: ask grc20 for a token using the canonical token's own
	// name/symbol/decimals. NewToken derives the token path from the live
	// crossing frame, so it names *this* realm; the id is the token object's
	// own, allocated under this realm. No argument changes either.
	impostor, impostorLedger = grc20.NewToken("Foo", "FOO", 4, cur)
}

func main(cur realm) {
	println("canonical token: ", canonical)
	println("impostor token:  ", impostor.TokenPath())
	println("impostor claims canonical token:", impostor.TokenPath() == canonical)
	println("impostor has its own id:", impostor.ID() != "")

	// A genuine grc20 event from the impostor. It carries grc20's pkg_path —
	// identical to the canonical token's events — but its own "token" and "id",
	// so an indexer never confuses the two ledgers.
	impostorLedger.Mint(cur.Address(), 999_999)

	// Route 2: skip grc20 and hand-emit an event carrying the canonical name.
	// The attributes are fully attacker-chosen, but pkg_path is not: the VM
	// stamps this realm. An indexer that accepts Transfer only from grc20's
	// pkg_path drops this event.
	chain.Emit(
		"Transfer",
		"id", "",
		"token", canonical,
		"from", "",
		"to", cur.Address().String(),
		"value", "999999",
	)
}

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "id",
//         "value": ""
//       },
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20provenance.FOO"
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
//         "key": "id",
//         "value": "g1chj4jxyc9djsldazyhl6szgyquyuumgjaph42m"
//       },
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20provenance.FOO"
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
//         "key": "id",
//         "value": ""
//       },
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/canonicaltoken.FOO"
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

// Output:
// canonical token:  gno.land/r/demo/canonicaltoken.FOO
// impostor token:   gno.land/r/demo/grc20provenance.FOO
// impostor claims canonical token: false
// impostor has its own id: true
