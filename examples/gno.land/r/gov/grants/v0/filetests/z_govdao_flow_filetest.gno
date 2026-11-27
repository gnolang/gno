// The whole life of two grants through the real GovDAO proxy: a member
// sponsors, votes and executes; payment 1 is claimable a window after
// approval, payment 2 after a report, a window and a period; a second grant
// is held and cancelled; GovDAO closes the desk and takes the pool back.
// This filetest realm is the grantee.

// PKGPATH: gno.land/r/test/golf
package golf

import (
	"chain"
	"chain/banker"
	"strconv"
	"testing"
	"time"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
	grants "gno.land/r/gov/grants/v0"
)

const member address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

const day = 24 * time.Hour

func init(cur realm) {
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, member, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))
}

func balance(a address) int64 { return banker.NewReadonlyBanker().GetCoin(a, "ugnot") }

func golf() int64 { return balance(chain.PackageAddress("gno.land/r/test/golf")) }

func skip(d time.Duration) {
	ctx := testing.GetContext()
	ctx.Time = ctx.Time.Add(d)
	testing.SetContext(ctx)
}

func main(cur realm) {
	// The realm asks, from its own code.
	id := grants.Apply(cross(cur), "Gnogolf v3", "New courses and a weekly cup to bring players to gno.land.",
		[]grants.Tranche{{Deliverable: "v3 design published", Amount: 80}, {Deliverable: "v3 live", Amount: 320}},
		30*day)
	testing.IssueCoins(grants.Address(), chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	// A member sponsors it; GovDAO votes and executes. Nothing is paid yet.
	testing.SetRealm(testing.NewUserRealm(member))
	pid := dao.MustCreateProposal(cross(cur), grants.NewApproveRequest(cross(cur), id, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	println("approve executed:", dao.ExecuteProposal(cross(cur), pid), "golf:", golf(), "committed:", grants.Committed())

	// Two weeks later anyone claims payment 1.
	skip(14 * day)
	testing.SetRealm(testing.NewUserRealm(member))
	grants.Claim(cross(cur), id)
	println("golf after payment 1:", golf())

	// The realm reports; a period after payment 1, anyone claims payment 2.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/golf"))
	grants.Report(cross(cur), id, "v3 is live: 4 new courses, 120 players this week.", "/r/test/gnogolf:v3")
	skip(30 * day)
	testing.SetRealm(testing.NewUserRealm(member))
	grants.Claim(cross(cur), id)
	println("golf after payment 2:", golf(), "committed:", grants.Committed())

	// A second grant: approved, then a member files a cancellation and
	// holds the next payment in the same breath. GovDAO cancels.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/golf"))
	id2 := grants.Apply(cross(cur), "Gnogolf cup", "A monthly cup.",
		[]grants.Tranche{{Deliverable: "cup rules published", Amount: 50}, {Deliverable: "first cup held", Amount: 250}}, 30*day)
	testing.SetRealm(testing.NewUserRealm(member))
	pid2 := dao.MustCreateProposal(cross(cur), grants.NewApproveRequest(cross(cur), id2, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid2), "YES")
	dao.ExecuteProposal(cross(cur), pid2)
	pid3 := dao.MustCreateProposal(cross(cur), grants.NewCancelRequest(cross(cur), id2, "The cup moved to another chain."))
	grants.Hold(cross(cur), id2, int64(pid3))
	println(grants.Render("grant/" + strconv.Itoa(id2)))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid3), "YES")
	println("cancel executed:", dao.ExecuteProposal(cross(cur), pid3), "committed:", grants.Committed(), "available:", grants.Available())

	// GovDAO closes the desk and takes the whole pool back.
	pid4 := dao.MustCreateProposal(cross(cur), grants.NewCloseRequest(cross(cur), chain.PackageAddress("gno.land/r/gov/dao/treasury/v0"), "Round closed.", true))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid4), "YES")
	println("close executed:", dao.ExecuteProposal(cross(cur), pid4))
	println("pool:", grants.Balance(), "treasury:", balance(chain.PackageAddress("gno.land/r/gov/dao/treasury/v0")))

	// What members read before voting on the first approval.
	p, _ := dao.GetProposal(pid)
	println()
	println(p.Title())
	println(p.Description())
}

// Output:
// approve executed: true golf: 0 committed: 400
// golf after payment 1: 80
// golf after payment 2: 400 committed: 0
// # <gno-icon name="lock" /> Grant #2: Gnogolf cup
//
// <gno-icon name="lock" /> **active** · [`r/test/golf`](/r/test/golf) ([source](/r/test/golf$source)) · namespace [test](/r/gov/grants/v0:ns/test)
//
// > [!NOTE] Payment 1 due 2009-04-26 23:31 UTC
// > 0.00005 GNOT to the grantee realm, unless GovDAO cancels before.
//
// > [!WARNING] Held
// > A GovDAO member filed [proposal #2](/r/gov/dao:2) to cancel this grant and held payment 1 14 days more while GovDAO votes.
//
// <gno-columns>
// <gno-frame>
// <gno-icon name="coins" /> Total
//
// ### 0.0003 GNOT
// </gno-frame>
// <gno-columns-sep />
// <gno-frame>
// <gno-icon name="check-circle" /> Paid
//
// ### 0 GNOT
// ▱▱
// </gno-frame>
// <gno-columns-sep />
// <gno-frame>
// <gno-icon name="calendar" /> Next payment
//
// ### 2009-04-26 23:31 UTC
// </gno-frame>
// <gno-columns-sep />
// <gno-frame>
// <gno-icon name="wallet" /> Paid to
//
// ### `r/test/golf`
// </gno-frame>
// </gno-columns>
//
// Payouts go to `g14y9694fdtdf8wuwfwturkzftcypv8nyprhpuum`, the grantee's own address. Where they go next is on [its page](/r/test/golf).
//
// ## <gno-icon name="list" /> Payments
//
// | # | Deliverable, in the grantee's words | Amount | Due | State |
// | --- | --- | --- | --- | --- |
// | 1 | cup rules published | 0.00005 GNOT | advance, 14 days after approval | <gno-icon name="arrow-right-circle" /> next |
// | 2 | first cup held | 0.00025 GNOT | report + 14 days, ≥ 30 days after payment 1 | <gno-icon name="circle" /> to come |
//
// ## <gno-icon name="timeline" /> Timeline
//
// | When | What |
// | --- | --- |
// | 2009-03-29 23:31 UTC | <gno-icon name="create" /> Requested by the grantee |
// | 2009-03-29 23:31 UTC | <gno-icon name="check" /> Approved by GovDAO, sponsored by `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`, 0.0003 GNOT reserved |
// | 2009-03-29 23:31 UTC | <gno-icon name="lock" /> Payment 1 held by [proposal #2](/r/gov/dao:2) |
//
// ## <gno-icon name="document" /> Pitch
//
// _Written by the grantee:_
//
// > A monthly cup\.
//
// ## <gno-icon name="shield-check" /> Sponsor's review
//
// _Written by `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`, who built the approval request a GovDAO member filed:_
//
// > Reviewed: test grantee\.
//
// ## <gno-icon name="chat" /> Reports
//
// None yet.
//
// ## <gno-icon name="users" /> For GovDAO members
//
// <gno-details summary="Scripts and review checklist">
//
// Cancel it with this MsgRun (payment 1 is already held):
//
// ```go
// package main
//
// import (
// 	"gno.land/r/gov/dao"
// 	grants "gno.land/r/gov/grants/v0"
// )
//
// func main(cur realm) {
// 	dao.MustCreateProposal(cross(cur), grants.NewCancelRequest(cross(cur), 2, "REASON"))
// }
// ```
//
// Freeze the grantee (no payment reaches it, it cannot apply, a wallet account cannot spend) with:
//
// ```go
// package main
//
// import (
// 	"gno.land/r/gov/dao"
// 	grants "gno.land/r/gov/grants/v0"
// )
//
// func main(cur realm) {
// 	dao.MustCreateProposal(cross(cur), grants.NewFreezeRequest(cross(cur), "gno.land/r/test/golf", true, "REASON"))
// }
// ```
//
// **This grantee is a realm of its own: read [its code](/r/test/golf$source) before you vote.**
//
// - Who controls it, and is that person or team known to GovDAO? Name them in your review.
// - Do you, or anyone voting, have an interest in it? Say so in your review, and abstain.
// - Do past reports and spends match what was promised? See [its other grants](/r/gov/grants/v0:ns/test).
// - Is `private = true` in its `gnomod.toml`? A private realm can be redeployed with new code at the same address after the vote: do not fund one.
// - Who can make it call `Report`? That report is the gate before each payment.
// - Every `SendCoins` and `NewBanker`: to whom, behind which check?
// - Can an admin change who that is: an ownership transfer, a stored address, func or interface?
// - Does it pass `cur` to code it does not own, or import realms with admins of their own?
// - Can the coins get stuck, with no way to spend them?
//
// Automated help: [the audit pattern harness](https://github.com/gnolang/gno/tree/master/misc/audit-pattern-harness) and [the contract review guide](https://github.com/gnolang/gno/blob/master/docs/resources/gno-ai-contract-review.md). If it only holds and spends grant money, suggest the wallet instead.
//
// </gno-details>
//
// cancel executed: true committed: 0 available: 600
// close executed: true
// pool: 0 treasury: 600
//
// Grant #1: 0.0004 GNOT to r/test/golf, Gnogolf v3
// If this passes, `gno.land/r/test/golf` can receive up to **0.0004 GNOT** in 2 payments: the first, an advance, 14 days after this executes; each later one only after a public report, 14 days to object, and at least 30 days after the previous one. No further vote is needed unless GovDAO cancels or freezes it.
//
// - Grantee: a realm of its own, [`r/test/golf`](/r/test/golf) ([source](/r/test/golf$source)): **read it before voting**. Payouts go to `g14y9694fdtdf8wuwfwturkzftcypv8nyprhpuum`.
// - Sponsor: `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`, whose review is below.
// - Grants filed by this namespace, this one included: 1 ([list](/r/gov/grants/v0:ns/test)).
// - Pool when proposed: 0.001 GNOT available of 0.001 GNOT.
// - Must execute before 2009-04-14 23:31 UTC, or the request expires.
// - Reports, reviewer checklist: [grant #1](/r/gov/grants/v0:grant/1).
//
// | # | Amount | Due | Deliverable, in the grantee's words |
// | --- | --- | --- | --- |
// | 1 | 0.00008 GNOT | advance, 14 days after approval | v3 design published |
// | 2 | 0.00032 GNOT | after a report | v3 live |
//
// Pitch, in the grantee's words (the full text is on the grant page): New courses and a weekly cup to bring players to gno\.land\.
//
// Sponsor's review, in full:
//
// > Reviewed: test grantee\.
