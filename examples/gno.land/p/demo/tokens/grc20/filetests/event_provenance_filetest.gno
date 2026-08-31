// PKGPATH: gno.land/r/demo/grc20provenance

// This filetest shows how an indexer attributes a grc20 event to its issuer.
//
// Two fields matter, and neither is sufficient alone:
//
//   - pkg_path is stamped by the VM from the package whose code calls
//     chain.Emit. Every genuine grc20 event therefore carries
//     gno.land/p/demo/tokens/grc20 — a constant, so it identifies the library
//     but never the token.
//   - the "token" attribute carries Token.ID(). The VM creates this opaque ID
//     from the issuing realm's package identifier and a persistent counter.
//
// Together they prove that grc20 emitted the event and that the VM assigned
// its token ID to the issuing realm. Callers cannot forge that package identity.
//
// The realm below is an impostor trying to pass its own transfers off as
// another realm's token, by both routes available to it, and failing at both.
package grc20provenance

import (
	"chain"

	"gno.land/p/demo/tokens/grc20"
)

const canonical = "gno:realm-id:unforgeable-package-id:1"

func main(cur realm) {
	println("canonical id:", canonical)

	impostor, impostorLedger := grc20.NewToken("Foo", "FOO", 4, cur)
	println("impostor id: ", impostor.ID())
	println("impostor claims canonical id:", impostor.ID() == canonical)

	impostorLedger.Mint(cur.Address(), 999_999)

	chain.Emit(
		"Transfer",
		"token", canonical,
		"from", "",
		"to", cur.Address().String(),
		"value", "999999",
	)
}

// Output:
// canonical id: gno:realm-id:unforgeable-package-id:1
// impostor id:  gno:realm-id:08f54204e7bb748ba312daa76fd11641ce50906d:5
// impostor claims canonical id: false

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno:realm-id:08f54204e7bb748ba312daa76fd11641ce50906d:5"
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
//         "value": "gno:realm-id:08f54204e7bb748ba312daa76fd11641ce50906d:5"
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
//         "value": "gno:realm-id:unforgeable-package-id:1"
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
