// A frozen owner can withdraw a pending handover offer, and nothing else:
// no withdrawal without an offer, so the journal cannot be flooded.

// PKGPATH: gno.land/r/test/walletflood
package walletflood

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
	wallet.TransferOwnership(cross(cur), "team", "g1rfznvu6qfa0sc76cplk5wpqexvefqccjunady0")
	pid := dao.MustCreateProposal(cross(cur), grants.NewFreezeRequest(cross(cur), "gno.land/r/gov/grants/wallet/v0#team", true, "Suspicious handover."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	println("frozen:", dao.ExecuteProposal(cross(cur), pid))
	println("withdraw form:", strings.Contains(wallet.Render("team"), `summary="Withdraw the handover offer"`))

	wallet.TransferOwnership(cross(cur), "team", member)
	println("withdrawn:", strings.Contains(wallet.Render("team"), "handover withdrawn"))
	wallet.TransferOwnership(cross(cur), "team", member)
}

// Output:
// frozen: true
// withdraw form: true
// withdrawn: true

// Error:
// no handover offer to withdraw
