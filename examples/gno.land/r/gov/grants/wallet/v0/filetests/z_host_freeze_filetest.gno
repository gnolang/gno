// Once GovDAO freezes the grant wallet itself, an account's page stops
// offering the Apply form, which would abort.

// PKGPATH: gno.land/r/test/walletfreeze
package walletfreeze

import (
	"strings"
	"testing"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
	grants "gno.land/r/gov/grants/v0"
	wallet "gno.land/r/gov/grants/wallet/v0"
)

const member address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

func init(cur realm) {
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, member, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(member))
	wallet.Open(cross(cur), "team")
	wallet.Apply(cross(cur), "team", "Team", "A pitch.", "Done: 1", 0)
	println(strings.Contains(wallet.Render("team"), `exec="Apply"`))

	pid := dao.MustCreateProposal(cross(cur), grants.NewFreezeRequest(cross(cur), "gno.land/r/gov/grants/wallet/v0", true, "Pause intake."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	println(dao.ExecuteProposal(cross(cur), pid))
	page := wallet.Render("team")
	println(strings.Contains(page, `exec="Apply"`), strings.Contains(page, "froze the grant wallet"))
}

// Output:
// true
// true
// false true
