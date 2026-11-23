// PKGPATH: gno.land/r/test/copy
package copy

import store "gno.land/r/gnoland/store/v0"

// A Register refused for a title another listing holds names that listing,
// so GovDAO knows which one a squat would be.
func main(cur realm) {
	store.Register(cross(cur), "boards-copy", "Boards", "A tagline", "social", 0)
}

// Events:
// [
//   {
//     "type": "StoreRegisterRejected",
//     "attrs": [
//       {
//         "key": "path",
//         "value": "gno.land/r/test/copy"
//       },
//       {
//         "key": "reason",
//         "value": "title_taken"
//       },
//       {
//         "key": "holder",
//         "value": "boards"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   }
// ]
