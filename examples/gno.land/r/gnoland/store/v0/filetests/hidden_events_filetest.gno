// PKGPATH: gno.land/r/test/hidev
package hidev

import (
	"chain/runtime/unsafe"
	"testing"

	"gno.land/p/nt/testutils/v0"
	store "gno.land/r/gnoland/store/v0"
	"gno.land/r/gov/dao"
	daoinit "gno.land/r/gov/dao/init/v0"
)

var c address = unsafe.OriginCaller()

func init(cur realm) {
	daoinit.InitWithUsers(cross(cur), c)
}

func hide(cur realm, slug string) {
	testing.SetOriginCaller(c)
	testing.SetRealm(testing.NewUserRealm(c))
	pid := dao.MustCreateProposal(cross(cur), store.ProposeHide(cross(cur), slug, "other"))
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	dao.ExecuteProposal(cross(cur), pid)
}

// SubmitRich, Register and Claim on a hidden listing succeed, and emit
// nothing: its hide released its slug.
func main(cur realm) {
	const app = "gno.land/r/test/hidev"
	owner := testutils.TestAddress("hidevowner")
	lib := "gno.land/p/" + owner.String() + "/lib"
	testing.SetRealm(testing.NewCodeRealm(app))
	store.Register(cross(cur), "hid-ev", "Hid Ev", "A tagline", "utilities", 0)
	testing.SetRealm(testing.NewUserRealm(owner))
	store.Claim(cross(cur), lib, "hid-ev-lib", "Hid Ev Lib", "A tagline", "dev-tools", 0, "package")
	hide(cur, "hid-ev")
	hide(cur, "hid-ev-lib")
	println("hidden")

	testing.SetRealm(testing.NewCodeRealm(app))
	store.Register(cross(cur), "hid-ev", "Hid Ev Two", "A tagline", "utilities", 0)
	testing.SetRealm(testing.NewCodeRealm(app))
	store.SubmitRich(cross(cur), "ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi", "")
	testing.SetOriginCaller(owner)
	testing.SetRealm(testing.NewUserRealm(owner))
	store.Claim(cross(cur), lib, "hid-ev-lib", "Hid Ev Lib Two", "A tagline", "dev-tools", 0, "package")
	println("written while hidden")
}

// Output:
// hidden
// written while hidden

// Events:
// [
//   {
//     "type": "StoreRegistered",
//     "attrs": [
//       {
//         "key": "path",
//         "value": "gno.land/r/test/hidev"
//       },
//       {
//         "key": "slug",
//         "value": "hid-ev"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   },
//   {
//     "type": "StoreClaimed",
//     "attrs": [
//       {
//         "key": "path",
//         "value": "gno.land/p/g1dp5kgetkdamkuetjta047h6lta047h6ln32u9k/lib"
//       },
//       {
//         "key": "slug",
//         "value": "hid-ev-lib"
//       },
//       {
//         "key": "kind",
//         "value": "package"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   },
//   {
//     "type": "ProposalCreated",
//     "attrs": [
//       {
//         "key": "id",
//         "value": "0"
//       }
//     ],
//     "pkg_path": "gno.land/r/gov/dao"
//   },
//   {
//     "type": "StoreHidden",
//     "attrs": [
//       {
//         "key": "slug",
//         "value": "hid-ev"
//       },
//       {
//         "key": "path",
//         "value": "gno.land/r/test/hidev"
//       },
//       {
//         "key": "reason",
//         "value": "other"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   },
//   {
//     "type": "ProposalCreated",
//     "attrs": [
//       {
//         "key": "id",
//         "value": "1"
//       }
//     ],
//     "pkg_path": "gno.land/r/gov/dao"
//   },
//   {
//     "type": "StoreHidden",
//     "attrs": [
//       {
//         "key": "slug",
//         "value": "hid-ev-lib"
//       },
//       {
//         "key": "path",
//         "value": "gno.land/p/g1dp5kgetkdamkuetjta047h6lta047h6ln32u9k/lib"
//       },
//       {
//         "key": "reason",
//         "value": "other"
//       }
//     ],
//     "pkg_path": "gno.land/r/gnoland/store/v0"
//   }
// ]
