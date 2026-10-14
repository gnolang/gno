// PKGPATH: gno.land/r/test/eventsapp
package eventsapp

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	store "gno.land/r/gnoland/store/v0"
)

// Every successful write emits an event: an indexer sees the store's state
// change, not only an addpkg.
func main(cur realm) {
	store.Register(cross(cur), "events-app", "Events App", "A tagline", "utilities", 0)
	store.SubmitRich(cross(cur), "ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi", "")

	fan := testutils.TestAddress("eventsfan")
	testing.SetRealm(testing.NewUserRealm(fan))
	store.Star(cross(cur), "events-app")
	store.Unstar(cross(cur), "events-app")
	store.Claim(cross(cur), "gno.land/p/"+fan.String()+"/lib", "events-lib", "Events Lib", "A tagline", "dev-tools", 0, "package")
}

// Events:
// [
//   {
//     "type": "StoreRegistered",
//     "attrs": [
//       {
//         "key": "path",
//         "value": "gno.land/r/test/eventsapp"
//       },
//       {
//         "key": "slug",
//         "value": "events-app"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   },
//   {
//     "type": "StoreRichSubmitted",
//     "attrs": [
//       {
//         "key": "slug",
//         "value": "events-app"
//       },
//       {
//         "key": "path",
//         "value": "gno.land/r/test/eventsapp"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   },
//   {
//     "type": "StoreStarred",
//     "attrs": [
//       {
//         "key": "slug",
//         "value": "events-app"
//       },
//       {
//         "key": "ranked",
//         "value": "false"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   },
//   {
//     "type": "StoreUnstarred",
//     "attrs": [
//       {
//         "key": "slug",
//         "value": "events-app"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   },
//   {
//     "type": "StoreClaimed",
//     "attrs": [
//       {
//         "key": "path",
//         "value": "gno.land/p/g1v4mx2mn5wdnxzmjlta047h6lta047h6lumj0f8/lib"
//       },
//       {
//         "key": "slug",
//         "value": "events-lib"
//       },
//       {
//         "key": "kind",
//         "value": "package"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   }
// ]
