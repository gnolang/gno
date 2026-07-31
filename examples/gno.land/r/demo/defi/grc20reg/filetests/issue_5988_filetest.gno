// PKGPATH: gno.land/r/demo/grc20reg_issue_5988

package grc20reg_issue_5988

import (
	"gno.land/p/demo/tokens/grc20"
	"gno.land/r/demo/defi/grc20reg"
)

func main(cur realm) {
	first, _ := grc20.NewToken("First Token", "DUP", 6, 1, cur)
	second, _ := grc20.NewToken("Second Token", "DUP", 6, 2, cur)

	firstKey := grc20reg.Register(cross(cur), first, "first")
	secondKey := grc20reg.Register(cross(cur), second, "second")

	if firstKey != "gno.land/r/demo/grc20reg_issue_5988.first" {
		panic("unexpected first registry key: " + firstKey)
	}
	if secondKey != "gno.land/r/demo/grc20reg_issue_5988.second" {
		panic("unexpected second registry key: " + secondKey)
	}
	if grc20reg.MustGet(firstKey).ID() != first.ID() {
		panic("first token not registered")
	}
	if grc20reg.MustGet(secondKey).ID() != second.ID() {
		panic("second token not registered")
	}
}

// Events:
// [
//   {
//     "type": "register",
//     "attrs": [
//       {
//         "key": "token_path",
//         "value": "gno.land/r/demo/grc20reg_issue_5988.first"
//       },
//       {
//         "key": "pkgpath",
//         "value": "gno.land/r/demo/grc20reg_issue_5988"
//       },
//       {
//         "key": "slug",
//         "value": "first"
//       },
//       {
//         "key": "symbol",
//         "value": "DUP"
//       },
//       {
//         "key": "token_id",
//         "value": "gno.land/r/demo/grc20reg_issue_5988.DUP.0000001"
//       }
//     ],
//     "pkg_path": "gno.land/r/demo/defi/grc20reg"
//   },
//   {
//     "type": "register",
//     "attrs": [
//       {
//         "key": "token_path",
//         "value": "gno.land/r/demo/grc20reg_issue_5988.second"
//       },
//       {
//         "key": "pkgpath",
//         "value": "gno.land/r/demo/grc20reg_issue_5988"
//       },
//       {
//         "key": "slug",
//         "value": "second"
//       },
//       {
//         "key": "symbol",
//         "value": "DUP"
//       },
//       {
//         "key": "token_id",
//         "value": "gno.land/r/demo/grc20reg_issue_5988.DUP.0000002"
//       }
//     ],
//     "pkg_path": "gno.land/r/demo/defi/grc20reg"
//   }
// ]
