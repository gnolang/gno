// PKGPATH: gno.land/r/demo/grc20identity

package grc20identity

import (
	"chain"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/onbloc/identifier"
)

var tokenIDGen *identifier.Generator

func main(cur realm) {
	tokenIDGen = identifier.NewGenerator(0, cur)

	first, firstLedger := newToken(cur)
	second, secondLedger := newToken(cur)

	if first.ID() != "gno.land/r/demo/grc20identity.DUP.8730z7x0xz79m" {
		panic("unexpected first token ID: " + first.ID())
	}
	if second.ID() != "gno.land/r/demo/grc20identity.DUP.j4eanafzxnc9j" {
		panic("unexpected second token ID: " + second.ID())
	}

	holder := chain.PackageAddress("holder")
	spender := chain.PackageAddress("spender")
	firstLedger.Mint(holder, 1)
	firstLedger.Approve(holder, spender, 1)
	secondLedger.Mint(holder, 2)
	secondLedger.Approve(holder, spender, 2)
}

func newToken(cur realm) (*grc20.Token, *grc20.PrivateLedger) {
	return grc20.NewToken("Same", "DUP", 6, tokenIDGen, cur)
}

// Events:
// [
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP.8730z7x0xz79m"
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
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "Approval",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP.8730z7x0xz79m"
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
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP.j4eanafzxnc9j"
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
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "Approval",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP.j4eanafzxnc9j"
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
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   }
// ]
