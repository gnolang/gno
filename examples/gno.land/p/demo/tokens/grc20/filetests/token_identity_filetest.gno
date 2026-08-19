// PKGPATH: gno.land/r/demo/grc20identity

package grc20identity

import (
	"chain"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/seqid/v0"
)

// This realm issues ids for its own tokens instead of taking them from a
// registry, so its tokens are not registrable there.
var tokenIDSeq seqid.ID

func nextIDSeq() string {
	return tokenIDSeq.Next().String()
}

func generator(cur realm) *grc20.IDGenerator {
	return grc20.NewIDGenerator(0, cur, nextIDSeq)
}

func main(cur realm) {
	first, firstLedger := newToken(cur, generator(cur))
	second, secondLedger := newToken(cur, generator(cur))

	// Same realm, same symbol, distinct ids: each token drew its own id.
	if first.ID() != "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000001" {
		panic("unexpected first token ID: " + first.ID())
	}
	if second.ID() != "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000002" {
		panic("unexpected second token ID: " + second.ID())
	}

	// Copying the generator copies nothing: both copies drive the same counter,
	// so the ids they produce are still distinct.
	g := generator(cur)
	a, b := *g, *g
	third, _ := newToken(cur, &a)
	fourth, _ := newToken(cur, &b)
	if third.ID() == fourth.ID() {
		panic("copied generator produced a duplicate id: " + third.ID())
	}

	holder := chain.PackageAddress("holder")
	spender := chain.PackageAddress("spender")
	firstLedger.Mint(holder, 1)
	firstLedger.Approve(holder, spender, 1)
	secondLedger.Mint(holder, 2)
	secondLedger.Approve(holder, spender, 2)
}

func newToken(cur realm, gen *grc20.IDGenerator) (*grc20.Token, *grc20.PrivateLedger) {
	return grc20.NewToken("Same", "DUP", 6, gen, cur)
}

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000001"
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
//         "value": "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000002"
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
//         "value": "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000003"
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
//         "value": "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000004"
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
//         "value": "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000001"
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
//         "value": "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000001"
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
//         "value": "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000002"
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
//         "value": "gno.land/r/demo/grc20identity.DUP.gno.land/r/demo/grc20identity:0000002"
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
