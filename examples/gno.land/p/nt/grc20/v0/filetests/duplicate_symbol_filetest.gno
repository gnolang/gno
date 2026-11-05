// PKGPATH: gno.land/r/demo/grc20dupsymbol

// The shape of gnolang/gno#6026, closed.
//
// A realm issues two independent tokens and names them both "DUP". Everything
// the realm chooses about them is identical, so nothing the realm chooses can
// tell their Transfer and Approval events apart. That used to include
// Token.ID(), whose trailing component was a caller supplied seqid the realm
// could pass twice. It is now the ObjectID the VM minted for each Token
// object, which no realm chooses, reuses or forges.
//
// The tokens are created in init so both are persisted by the time main runs.
// A token created and used inside one call has a path but no id yet, which the
// init below asserts rather than hides: see Token.ID.
package grc20dupsymbol

import (
	"gno.land/p/nt/grc20/v0"
)

var (
	first, second             *grc20.Token
	firstLedger, secondLedger *grc20.PrivateLedger
)

func init(cur realm) {
	first, firstLedger = grc20.NewToken("Same", "DUP", 6, cur)
	second, secondLedger = grc20.NewToken("Same", "DUP", 6, cur)

	// Both are inside the stamping window here, so the two NewToken events
	// below carry an empty id. That is the honest report, not a collision: the
	// realm is told the identity is not minted yet rather than handed a name
	// two tokens share.
	if first.ID() != "" || second.ID() != "" {
		panic("tokens created in init cannot be stamped yet")
	}
}

func main(cur realm) {
	// The realm named them identically, and that is allowed.
	println("first path: ", first.TokenPath())
	println("second path:", second.TokenPath())
	println("paths collide:", first.TokenPath() == second.TokenPath())

	// The VM did not. This is what #6026 needs: two tokens one realm issued
	// under one symbol, told apart by something the realm could not choose.
	println("first id: ", first.ID())
	println("second id:", second.ID())
	println("ids collide:", first.ID() == second.ID())

	// So these two Transfers are attributable, with no special casing in the
	// consumer: two independent ledgers, two ids.
	firstLedger.Mint(cur.Address(), 1)
	secondLedger.Mint(cur.Address(), 999999)
}

// Output:
// first path:  gno.land/r/demo/grc20dupsymbol.DUP
// second path: gno.land/r/demo/grc20dupsymbol.DUP
// paths collide: true
// first id:  095c132548af5675291edd1a020f2f82bd0608c1:10
// second id: 095c132548af5675291edd1a020f2f82bd0608c1:16
// ids collide: false

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20dupsymbol.DUP"
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
//         "value": "gno.land/r/demo/grc20dupsymbol.DUP"
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
//         "value": "gno.land/r/demo/grc20dupsymbol.DUP"
//       },
//       {
//         "key": "id",
//         "value": "095c132548af5675291edd1a020f2f82bd0608c1:10"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1tt7q94jxk2wk7lxmfanp4zq8tukrjegmnhxhxp"
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
//         "value": "gno.land/r/demo/grc20dupsymbol.DUP"
//       },
//       {
//         "key": "id",
//         "value": "095c132548af5675291edd1a020f2f82bd0608c1:16"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1tt7q94jxk2wk7lxmfanp4zq8tukrjegmnhxhxp"
//       },
//       {
//         "key": "value",
//         "value": "999999"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v0"
//   }
// ]
