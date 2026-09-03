// PKGPATH: gno.land/r/demo/grc20identity

package grc20identity

import (
	"chain"

	"gno.land/p/demo/tokens/grc20"
)

func main(cur realm) {
	first, firstLedger := newToken(cur)
	second, secondLedger := newToken(cur)

	if first.ID() == second.ID() {
		panic("duplicate token IDs")
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
//         "value": "gno.land/r/demo/grc20identity:6"
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
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity:7"
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
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity:6"
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
//         "value": "gno.land/r/demo/grc20identity:6"
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
//         "value": "gno.land/r/demo/grc20identity:7"
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
//         "value": "gno.land/r/demo/grc20identity:7"
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
