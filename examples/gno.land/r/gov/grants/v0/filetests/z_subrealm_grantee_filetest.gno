// A DAO hosted by a realm applies under its own sub-realm identity
// (cur.Sub), the way the Common DAO Spec gives every DAO an address. It is a
// grantee of its own: its request does not supersede the host's, it reports
// under that identity, and once approved its payments reach the DAO's
// derived address, not the host's.

// PKGPATH: gno.land/r/test/golf
package golf

import (
	"chain"
	"chain/banker"
	"testing"
	"time"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
	grants "gno.land/r/gov/grants/v0"
)

const member address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

func init(cur realm) {
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, member, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))
}

func balance(p string) int64 {
	return banker.NewReadonlyBanker().GetCoin(chain.PackageAddress(p), "ugnot")
}

func main(cur realm) {
	tr := []grants.Tranche{{Deliverable: "all", Amount: 70}}
	first := grants.Apply(cross(cur), "Host request", "p", tr, 0)
	sub := cur.Sub("dao1")
	second := grants.Apply(cross(sub), "DAO request", "p", tr, 0)
	println(first, second)

	testing.IssueCoins(grants.Address(), chain.NewCoins(chain.NewCoin("ugnot", 70)))
	testing.SetRealm(testing.NewUserRealm(member))
	pid := dao.MustCreateProposal(cross(cur), grants.NewApproveRequest(cross(cur), second, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	dao.ExecuteProposal(cross(cur), pid)
	// A single payment is paid once its work is reported.
	grants.Report(cross(sub), second, "all done", "")
	ctx := testing.GetContext()
	ctx.Time = ctx.Time.Add(14 * 24 * time.Hour)
	testing.SetContext(ctx)
	testing.SetRealm(testing.NewUserRealm(member))
	grants.Claim(cross(cur), second)
	println("dao1:", balance("gno.land/r/test/golf#dao1"), "host:", balance("gno.land/r/test/golf"))
	println(grants.Render("ns/test"))
}

// Output:
// 1 2
// dao1: 70 host: 0
// # <gno-icon name="users" /> Grants to the `test` namespace
//
// Its [wallet accounts](/r/gov/grants/wallet/v0) and its realms under `gno.land/r/test`.
//
// <gno-columns>
// <gno-frame>
// <gno-icon name="list" /> Grants filed
//
// ### 2
// </gno-frame>
// <gno-columns-sep />
// <gno-frame>
// <gno-icon name="coins" /> Still to pay, approved grants
//
// ### 0 GNOT
// </gno-frame>
// <gno-columns-sep />
// <gno-frame>
// <gno-icon name="check-circle" /> Paid
//
// ### 0.00007 GNOT
// </gno-frame>
// </gno-columns>
//
// | # | Grant | Grantee | Paid | Status |
// | --- | --- | --- | --- | --- |
// | 1 | [Host request](/r/gov/grants/v0:grant/1) | `r/test/golf` | ▱ 0 GNOT of 0.00007 GNOT | <gno-icon name="clock" /> **pending** |
// | 2 | [DAO request](/r/gov/grants/v0:grant/2) | `r/test/golf#dao1` | ▰ 0.00007 GNOT of 0.00007 GNOT | <gno-icon name="check-circle" /> **paid in full** |
