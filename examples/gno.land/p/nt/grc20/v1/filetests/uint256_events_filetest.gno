// PKGPATH: gno.land/r/demo/grc20u256events

// Events carry amounts as canonical decimal strings, including amounts no
// int64 can hold. A refused mint (past MaxUint256) emits nothing.
package grc20u256events

import (
	"chain"
	"math"

	"gno.land/p/nt/grc20/v1"
)

var (
	token  *grc20.Token
	ledger *grc20.PrivateLedger
)

func init(cur realm) {
	token, ledger = grc20.NewToken("Big", "BIG", 18, 0, cur)
}

func main(cur realm) {
	alice := chain.PackageAddress("alice")
	bob := chain.PackageAddress("bob")

	if err := ledger.Mint(alice, math.MaxUint256); err != nil {
		panic(err)
	}
	if err := ledger.Mint(bob, 1); err != grc20.ErrMintOverflow {
		panic("expected ErrMintOverflow")
	}
	if err := ledger.Transfer(alice, bob, 1<<64); err != nil {
		panic(err)
	}
	if err := ledger.Approve(bob, alice, math.MaxUint256); err != nil {
		panic(err)
	}
	println(token.FormatAmount(token.BalanceOf(bob)))
	println(token.TotalSupply() == math.MaxUint256)
}

// Output:
// 18.446744073709551616
// true

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20u256events.BIG.0000000"
//       },
//       {
//         "key": "name",
//         "value": "Big"
//       },
//       {
//         "key": "symbol",
//         "value": "BIG"
//       },
//       {
//         "key": "decimals",
//         "value": "18"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v1"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20u256events.BIG.0000000"
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
//         "value": "115792089237316195423570985008687907853269984665640564039457584007913129639935"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v1"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20u256events.BIG.0000000"
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
//         "value": "18446744073709551616"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v1"
//   },
//   {
//     "type": "Approval",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20u256events.BIG.0000000"
//       },
//       {
//         "key": "owner",
//         "value": "g1zqwap9sxm4zcvt9vj5tll4uhtrtwxjhe5789f8"
//       },
//       {
//         "key": "spender",
//         "value": "g1vxf6grxsp469xvdqnrkgjjrqjjydmve3kfs04e"
//       },
//       {
//         "key": "value",
//         "value": "115792089237316195423570985008687907853269984665640564039457584007913129639935"
//       }
//     ],
//     "pkg_path": "gno.land/p/nt/grc20/v1"
//   }
// ]
