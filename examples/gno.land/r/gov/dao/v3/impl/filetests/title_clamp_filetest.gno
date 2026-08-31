// PKGPATH: gno.land/r/test/titleclamp
package titleclamp

// A proposal title is attacker-chosen (by a member) and is escaped on two
// pages: the proposal page and the list page. Escaping costs about 6,990 gas a
// byte, and the list page escapes one title per proposal shown, so an oversized
// title priced the whole list out of the query cap: five proposals with 90 KB
// titles cost 3,172,507,361 gas against a 3,000,000,000 cap. Clamping the title
// before it is escaped brings the same five to 66,871,916.
//
// Runs in its own realm because it creates proposals, and the unit tests in the
// impl package assert hard-coded proposal ids against shared state.

import (
	"strings"
	"testing"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

const user address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

func init(cur realm) {
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, user, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))
}

func main(cur realm) {
	testing.SetOriginCaller(user)
	testing.SetRealm(testing.NewUserRealm(user))

	short := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"An ordinary title", "d", nil))
	long := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		strings.Repeat("t", 50000), "d", nil))

	// An ordinary title is untouched. Real titles in examples/ are about 40
	// bytes, so the bound never reaches them.
	sout := dao.Render(cross(cur), short.String())
	println("an ordinary title is left alone:",
		strings.Contains(sout, "An ordinary title") && !strings.Contains(sout, "… truncated"))

	// The proposal page cuts the title before escaping it. Asserting the marker
	// rather than a length: escaping first and cutting second would also
	// produce a short page, so length alone cannot tell the two apart.
	pout := dao.Render(cross(cur), long.String())
	println("the proposal page clamps a huge title:",
		strings.Contains(pout, "… truncated") && len(pout) < 4000)

	// The list page escapes one title per proposal it shows, so it is the page
	// the bound actually protects.
	lout := dao.Render(cross(cur), "")
	println("the list page clamps it too:",
		strings.Contains(lout, "… truncated") && len(lout) < 4000)
}

// Output:
// an ordinary title is left alone: true
// the proposal page clamps a huge title: true
// the list page clamps it too: true
