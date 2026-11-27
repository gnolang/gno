// A paused names check refuses every address: account pages must not
// blame their owners for it.

// PKGPATH: gno.land/r/test/walletpause
package walletpause

import (
	"strings"
	"testing"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
	wallet "gno.land/r/gov/grants/wallet/v0"
	"gno.land/r/sys/names"
)

const member address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

func init(cur realm) {
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, member, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"))
	names.Enable(cross(cur))
	testing.SetRealm(testing.NewUserRealm(member))
	wallet.Open(cross(cur), member.String())
	println("before pause:", strings.Contains(wallet.Render(member.String()), "Owner outside the namespace"))

	pid := dao.MustCreateProposal(cross(cur), names.ProposeSetPaused(cross(cur), true))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	println("paused:", dao.ExecuteProposal(cross(cur), pid), names.IsPaused())
	println("after pause:", strings.Contains(wallet.Render(member.String()), "Owner outside the namespace"))
}

// Output:
// before pause: false
// paused: true true
// after pause: false
