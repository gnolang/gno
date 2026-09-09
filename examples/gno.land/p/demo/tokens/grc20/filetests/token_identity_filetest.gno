// PKGPATH: gno.land/r/demo/grc20identity

// Two tokens built by the same code path, with the same name/symbol/decimals.
// They share a token path — that is just rlmPath.symbol — but not an identity:
// Token.ID() is the address derived from each token object's own VM object ID,
// which the realm has no say in. So the Transfer/Approval events below name
// exactly one of the two ledgers through their "id" attribute, while "token"
// stays the readable name both were issued under.
//
// This is what makes the event stream attributable. Before, the only identifier
// was the path plus a caller-supplied sequence number, and a realm that passed
// the same one twice produced two independent ledgers behind one identifier —
// with no way for an indexer to tell their events apart.
//
// The tokens are created in init rather than main because an object ID is
// stamped the first time the object is persisted, which happens as the creating
// frame returns. For the rest of init there is nothing to derive an address
// from and both tokens read ""; by the time main runs they carry their final,
// distinct ids. That is why the two NewToken events below carry an empty "id"
// and a non-empty "token".
package grc20identity

import (
	"chain"

	"gno.land/p/demo/tokens/grc20"
)

var (
	first, second             *grc20.Token
	firstLedger, secondLedger *grc20.PrivateLedger
)

func init(cur realm) {
	first, firstLedger = grc20.NewToken("Same", "DUP", 6, cur)
	second, secondLedger = grc20.NewToken("Same", "DUP", 6, cur)
}

func main(cur realm) {
	if first.ID() == "" {
		panic("first token has no ID")
	}
	if first.ID() == second.ID() {
		panic("duplicate token ID: " + first.ID())
	}
	if first.TokenPath() != second.TokenPath() {
		panic("same realm and symbol should share a token path")
	}

	holder := chain.PackageAddress("holder")
	spender := chain.PackageAddress("spender")
	firstLedger.Mint(holder, 1)
	firstLedger.Approve(holder, spender, 1)
	secondLedger.Mint(holder, 2)
	secondLedger.Approve(holder, spender, 2)
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
//         "value": "gno.land/r/demo/grc20identity.DUP"
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
//         "key": "id",
//         "value": ""
//       },
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
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
//         "key": "id",
//         "value": "g16y53pkwwqjdu0f2uyrxs046gh4nw2gvfrqzhts"
//       },
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
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
//         "key": "id",
//         "value": "g16y53pkwwqjdu0f2uyrxs046gh4nw2gvfrqzhts"
//       },
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
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
//         "key": "id",
//         "value": "g1velwwsg7044kr92yar8yz4uhn6ra6zfj2l5k6m"
//       },
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
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
//         "key": "id",
//         "value": "g1velwwsg7044kr92yar8yz4uhn6ra6zfj2l5k6m"
//       },
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20identity.DUP"
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
