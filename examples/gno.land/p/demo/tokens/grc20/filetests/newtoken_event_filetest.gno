// PKGPATH: gno.land/r/demo/grc20dupsignal

// Two tokens with identical metadata receive distinct VM assigned IDs.
// Their NewToken and Transfer events remain attributable to one token.
package grc20dupsignal

import (
	"gno.land/p/demo/tokens/grc20"
)

func main(cur realm) {
	first, firstLedger := grc20.NewToken("Same", "DUP", 6, cur)
	second, secondLedger := grc20.NewToken("Same", "DUP", 6, cur)

	println("same id:", first.ID() == second.ID())

	firstLedger.Mint(cur.Address(), 1)
	secondLedger.Mint(cur.Address(), 999999)
}

// Output:
// same id: false

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno:realm-id:083acbcef4b740a693cac246bf49a6128be52c82:5"
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
//         "value": "gno:realm-id:083acbcef4b740a693cac246bf49a6128be52c82:6"
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
//         "value": "gno:realm-id:083acbcef4b740a693cac246bf49a6128be52c82:5"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1k29lxz0d8gx7m94n032ulum875p7gg0kduus5m"
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
//         "value": "gno:realm-id:083acbcef4b740a693cac246bf49a6128be52c82:6"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1k29lxz0d8gx7m94n032ulum875p7gg0kduus5m"
//       },
//       {
//         "key": "value",
//         "value": "999999"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   }
// ]
