// PKGPATH: gno.land/r/demo/grc20zero

// Zero-amount operations succeed and emit their events, like v0, but store
// nothing: an absent key and a zero balance or allowance are the same state,
// so KnownAccounts and the render do not count the touched accounts.
package grc20zero

import (
	"chain"

	"gno.land/p/nt/grc20/v1"
)

func main(cur realm) {
	token, ledger := grc20.NewToken("Zero", "ZERO", 0, 0, cur)
	alice := chain.PackageAddress("alice")
	bob := chain.PackageAddress("bob")

	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	check(ledger.Mint(alice, 0))
	check(ledger.Transfer(alice, bob, 0))
	check(ledger.Approve(alice, bob, 0))
	check(ledger.SpendAllowance(alice, bob, 0))

	println("alice known:", token.HasAddr(alice))
	println("bob known:", token.HasAddr(bob))
	println("known accounts:", token.KnownAccounts())
	println("allowance:", token.Allowance(alice, bob))
	print(token.RenderHome())
}

// Output:
// alice known: false
// bob known: false
// known accounts: 0
// allowance: 0
// # Zero ($ZERO)
//
// * **Decimals**: 0
// * **Total supply**: 0
// * **Known accounts**: 0

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20zero.ZERO.0000000"
//       },
//       {
//         "key": "name",
//         "value": "Zero"
//       },
//       {
//         "key": "symbol",
//         "value": "ZERO"
//       },
//       {
//         "key": "decimals",
//         "value": "0"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v1"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20zero.ZERO.0000000"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1vxf6grxsp469xvdqnrkgjjrqjjydmve3kfs04e"
//       },
//       {
//         "key": "value",
//         "value": "0"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v1"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20zero.ZERO.0000000"
//       },
//       {
//         "key": "from",
//         "value": "g1vxf6grxsp469xvdqnrkgjjrqjjydmve3kfs04e"
//       },
//       {
//         "key": "to",
//         "value": "g1zqwap9sxm4zcvt9vj5tll4uhtrtwxjhe5789f8"
//       },
//       {
//         "key": "value",
//         "value": "0"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v1"
//   },
//   {
//     "type": "Approval",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20zero.ZERO.0000000"
//       },
//       {
//         "key": "owner",
//         "value": "g1vxf6grxsp469xvdqnrkgjjrqjjydmve3kfs04e"
//       },
//       {
//         "key": "spender",
//         "value": "g1zqwap9sxm4zcvt9vj5tll4uhtrtwxjhe5789f8"
//       },
//       {
//         "key": "value",
//         "value": "0"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v1"
//   }
// ]
