// PKGPATH: gno.land/r/test/app
package app

import store "gno.land/r/gnoland/store/v0"

// A refused Register does not panic: it emits StoreRegisterRejected, so the
// caller's init keeps running.
func main(cur realm) {
	store.Register(cross(cur), "test-app", "Test App", "A tagline", "casino", 0)
	println(store.Render("c/utilities"))
}

// Output:
// # Utilities
//
// - **[Coin Balances](/r/gnoland/coins)** — Check any address's balances without an explorer
//
// [Back to Explore](/r/gnoland/store/v0)

// Events:
// [
//   {
//     "type": "StoreRegisterRejected",
//     "attrs": [
//       {
//         "key": "path",
//         "value": "gno.land/r/test/app"
//       },
//       {
//         "key": "reason",
//         "value": "bad_category"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   }
// ]
