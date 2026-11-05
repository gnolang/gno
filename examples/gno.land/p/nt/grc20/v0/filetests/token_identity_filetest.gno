// PKGPATH: gno.land/r/demo/grc20identity

// Every event a token emits carries both of its names: "token", the readable
// path it shares with any sibling the realm gave the same symbol, and "id",
// the identity the VM minted for that one Token object. The two tokens below
// are deliberately indistinguishable in everything the realm chose, so the id
// is the only thing in the stream that tells their Transfers and Approvals
// apart (gnolang/gno#6026).
//
// Created in init, so both are persisted before main emits anything.
package grc20identity

import (
	"chain"

	"gno.land/p/nt/grc20/v0"
)

var (
	first, second             *grc20.Token
	firstLedger, secondLedger *grc20.PrivateLedger
)

func init(cur realm) {
	first, firstLedger = newToken(cur)
	second, secondLedger = newToken(cur)
}

func main(cur realm) {
	if first.TokenPath() != "gno.land/r/demo/grc20identity.DUP" {
		panic("unexpected first token path: " + first.TokenPath())
	}
	if second.TokenPath() != first.TokenPath() {
		panic("unexpected second token path: " + second.TokenPath())
	}
	if first.ID() == "" || second.ID() == "" {
		panic("a persisted token must have an id")
	}
	if first.ID() == second.ID() {
		panic("two tokens must not share an id: " + first.ID())
	}

	holder := chain.PackageAddress("holder")
	spender := chain.PackageAddress("spender")
	firstLedger.Mint(holder, 1)
	firstLedger.Approve(holder, spender, 1)
	secondLedger.Mint(holder, 2)
	secondLedger.Approve(holder, spender, 2)
}

func newToken(cur realm) (*grc20.Token, *grc20.PrivateLedger) {
	return grc20.NewToken("Same", "DUP", 6, cur)
}

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
//       },
//       {
//         "key": "id",
//         "value": ""
//       },
//       {
//         "key": "name",
//         "value": "Same"
//       },
//       {
//         "key": "symbol",
//         "value": "DUP"
//       },
//       {
//         "key": "decimals",
//         "value": "6"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v0"
//   },
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
//       },
//       {
//         "key": "id",
//         "value": ""
//       },
//       {
//         "key": "name",
//         "value": "Same"
//       },
//       {
//         "key": "symbol",
//         "value": "DUP"
//       },
//       {
//         "key": "decimals",
//         "value": "6"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v0"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
//       },
//       {
//         "key": "id",
//         "value": "054f7b1709e31b44b2e39810b7be83378b0c25b7:11"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g18rwjsdf0kcy673ta0q4ujwvsncz4k2a8gcv94t"
//       },
//       {
//         "key": "value",
//         "value": "1"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v0"
//   },
//   {
//     "type": "Approval",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
//       },
//       {
//         "key": "id",
//         "value": "054f7b1709e31b44b2e39810b7be83378b0c25b7:11"
//       },
//       {
//         "key": "owner",
//         "value": "g18rwjsdf0kcy673ta0q4ujwvsncz4k2a8gcv94t"
//       },
//       {
//         "key": "spender",
//         "value": "g148tp8xkvvk3l73lmxywpdlsxgdujjlclx7wfyg"
//       },
//       {
//         "key": "value",
//         "value": "1"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v0"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
//       },
//       {
//         "key": "id",
//         "value": "054f7b1709e31b44b2e39810b7be83378b0c25b7:17"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g18rwjsdf0kcy673ta0q4ujwvsncz4k2a8gcv94t"
//       },
//       {
//         "key": "value",
//         "value": "2"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v0"
//   },
//   {
//     "type": "Approval",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
//       },
//       {
//         "key": "id",
//         "value": "054f7b1709e31b44b2e39810b7be83378b0c25b7:17"
//       },
//       {
//         "key": "owner",
//         "value": "g18rwjsdf0kcy673ta0q4ujwvsncz4k2a8gcv94t"
//       },
//       {
//         "key": "spender",
//         "value": "g148tp8xkvvk3l73lmxywpdlsxgdujjlclx7wfyg"
//       },
//       {
//         "key": "value",
//         "value": "2"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v0"
//   }
// ]
