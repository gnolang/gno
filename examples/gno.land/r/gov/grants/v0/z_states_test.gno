package grants

import (
	"chain"
	"chain/banker"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/r/gov/dao"
)

func setNow(t time.Time) {
	ctx := testing.GetContext()
	ctx.Time = t
	testing.SetContext(ctx)
}

// Every state a grant can be in shows, on its page and in the desk's
// tables, what it waits for and the call anyone can make when one is due.
// States are built on the desk directly, with explicit times.
func TestEveryStateTellsWhatComesNext(cur realm, t *testing.T) {
	now := time.Now()
	ago := func(d int) time.Time { return now.Add(-time.Duration(d) * day) }
	big := int64(1 << 50)
	mk := func(name string, applied int) int {
		id, err := desk.Apply("gno.land/r/states/"+name, "State "+name, "p", tranches(), 30*day, ago(applied))
		urequire.NoError(t, err)
		return id
	}
	approved := func(name string, at int) int {
		id := mk(name, at+1)
		urequire.NoError(t, desk.Approve(id, big, "", "", ago(at)))
		return id
	}
	claim := func(id, at int) {
		_, _, err := desk.Claim(id, ago(at))
		urequire.NoError(t, err)
	}

	fresh := mk("fresh", 1)
	stale := mk("stale", 61)
	dueLater := approved("duelater", 1)
	claimable := approved("claimable", 15)
	waiting := approved("waiting", 50)
	claim(waiting, 36)
	overdue := approved("overdue", 200)
	claim(overdue, 186)
	held := approved("held", 1)
	urequire.NoError(t, desk.Hold(held, 1_000_000, now))
	done := approved("done", 100)
	claim(done, 86)
	urequire.NoError(t, desk.Report("gno.land/r/states/done", done, "r", "", ago(50)))
	claim(done, 30)
	cancelled := approved("cancelled", 10)
	urequire.NoError(t, desk.Cancel(cancelled, "not delivered", ago(5)))
	replaced := mk("replaced", 3)
	mk("replaced", 2)
	gone := mk("gone", 70)
	urequire.NoError(t, desk.Expire(gone, ago(1)))

	cases := []struct {
		id         int
		alert, act string
	}{
		{fresh, "[!INFO] Waiting for GovDAO", "NewApproveRequest"},
		{stale, "[!CAUTION] Expired unapproved", `<gno-form exec="Expire">`},
		{dueLater, "[!NOTE] Payment 1 due", "NewCancelRequest"},
		{claimable, "[!SUCCESS] Payment 1 can be released", `<gno-form exec="Claim">`},
		{waiting, "[!NOTE] Waiting for the grantee's report on payment 2", "no sooner than"},
		{overdue, "[!CAUTION] Payment 2 never reported", `<gno-form exec="Expire">`},
		{held, "[!WARNING] Held", "already held"},
		{done, "[!SUCCESS] Paid in full", "paid "},
		{cancelled, "[!CAUTION] Cancelled", "not delivered"},
		{replaced, "[!CAUTION] Superseded", "superseded by"},
		{gone, "[!CAUTION] Expired", "nothing was reserved"},
	}
	// A hold pushes the earliest date shown while the report is awaited.
	heldWait := approved("heldwait", 50)
	claim(heldWait, 36)
	urequire.NoError(t, desk.Hold(heldWait, 1, now))
	cases = append(cases, struct {
		id         int
		alert, act string
	}{heldWait, "[!NOTE] Waiting for the grantee's report", "no sooner than " + when(now.Add(desk.Rules().HoldFor))})
	// An expired grant gave back what it was not paid.
	lapsed := approved("lapsed", 200)
	claim(lapsed, 186)
	urequire.NoError(t, desk.Expire(lapsed, ago(1)))
	cases = append(cases, struct {
		id         int
		alert, act string
	}{lapsed, "[!CAUTION] Expired", "Left unattended"})

	for _, c := range cases {
		page := Render("grant/" + itoa(c.id))
		uassert.True(t, strings.Contains(page, c.alert), itoa(c.id)+": "+c.alert)
		uassert.True(t, strings.Contains(page, c.act), itoa(c.id)+": "+c.act)
	}

	home := Render("")
	for _, want := range []string{"payment 1 ready", "expired unapproved", "never reported", "needs a GovDAO vote",
		"Needs attention", "State claimable", "State stale"} {
		uassert.True(t, strings.Contains(home, want), "home: "+want)
	}
}

// A report history longer than the grant page shows gets its own page.
func TestReportsPage(cur realm, t *testing.T) {
	now := time.Now()
	ago := func(d int) time.Time { return now.Add(-time.Duration(d) * day) }
	const who = "gno.land/r/states/reporter"
	id, err := desk.Apply(who, "Reporter", "p", []Tranche{{"a", 1}, {"b", 2}, {"c", 2}}, 30*day, ago(121))
	urequire.NoError(t, err)
	urequire.NoError(t, desk.Approve(id, 1<<50, "", "", ago(120)))
	_, _, err = desk.Claim(id, ago(106))
	urequire.NoError(t, err)
	for i := 0; i < 3; i++ {
		urequire.NoError(t, desk.Report(who, id, "report on two, take "+itoa(i), "", ago(100)))
	}
	_, _, err = desk.Claim(id, ago(70))
	urequire.NoError(t, err)
	urequire.NoError(t, desk.Report(who, id, "report on three", "", ago(10)))

	page := Render("grant/" + itoa(id))
	uassert.True(t, strings.Contains(page, "All 4 reports"))
	uassert.False(t, strings.Contains(page, "take 0"), "the oldest report belongs on the reports page")
	all := Render("grant/" + itoa(id) + "/reports")
	for _, want := range []string{"take 0", "take 1", "take 2", "report on three"} {
		uassert.True(t, strings.Contains(all, want), want)
	}
}

func TestListsPaginate(cur realm, t *testing.T) {
	for i := 0; i < pageSize+1; i++ {
		_, err := desk.Apply("gno.land/r/pages/p"+itoa(i), "Page "+itoa(i), "p", tranches(), 30*day, time.Now())
		urequire.NoError(t, err)
	}
	uassert.True(t, strings.Contains(Render(""), "requests\" />"), "home links to the rest")
	uassert.False(t, strings.HasPrefix(Render("waiting/2"), "# Not found"))
	uassert.True(t, strings.HasPrefix(Render("waiting/99"), "# Not found"))
	uassert.False(t, strings.HasPrefix(Render("all/2"), "# Not found"))
	uassert.True(t, strings.Contains(Render("waiting/1"), `label="Next"`))
	uassert.True(t, strings.Contains(Render(""), link("waiting/1")+`" label="All `), "the button opens the first page")
}

// A hold really delays the payment at the desk, by HoldFor after it would
// have been due.
func TestHoldDelaysClaim(cur realm, t *testing.T) {
	start := time.Now()
	const who = "gno.land/r/states/heldpay"
	id, err := desk.Apply(who, "Held", "p", tranches(), 30*day, start)
	urequire.NoError(t, err)
	urequire.NoError(t, desk.Approve(id, 1<<50, "", "", start))
	urequire.NoError(t, desk.Hold(id, 2_000_000, start.Add(day)))
	testing.IssueCoins(Address(), coins(Committed()))

	setNow(start.Add(14 * day))
	uassert.AbortsContains(t, cur, "is payable at", func() { Claim(cross(cur), id) })
	setNow(start.Add(28 * day))
	uassert.NotAborts(t, cur, func() { Claim(cross(cur), id) })
	uassert.Equal(t, int64(80), banker.NewReadonlyBanker().GetCoin(chain.PackageAddress(who), Denom))
	setNow(start)
}

// The exact quoting the pages rely on: every line escaped and quoted, a
// line of "=" unable to make a heading.
func TestQuote(t *testing.T) {
	uassert.Equal(t, "> \\#\\# x\n> \\===\n\n", quote("## x\n==="))
	uassert.Equal(t, "> \\`\\`\\`go\n> evil\\(\\)\n\n", quote("```go\nevil()"))
}

// GrantsOf lists exactly one grantee, newest first.
func TestGrantsOfIsExactAndNewestFirst(cur realm, t *testing.T) {
	const who = "gno.land/r/states/many"
	var last int
	for i := 0; i < 22; i++ {
		id, err := desk.Apply(who, "Many", "p", tranches(), 30*day, time.Now())
		urequire.NoError(t, err)
		last = id
	}
	_, err := desk.Apply(who+"#sub", "Sub", "p", tranches(), 30*day, time.Now())
	urequire.NoError(t, err)
	list := GrantsOf(who)
	urequire.Equal(t, 20, len(list))
	uassert.Equal(t, last, list[0].ID)
	for _, g := range list {
		uassert.Equal(t, who, g.Grantee)
	}
}

// Proposals that can no longer do what they say fail with an error, so
// ExecuteOrRejectProposal closes them instead of leaving them executable.
func TestStaleProposalsAreRejectable(cur realm, t *testing.T) {
	seatMember(cur)
	testing.IssueCoins(Address(), coins(Committed()+1_000))

	// A Return the pool can no longer cover once a grant took the funds.
	testing.SetRealm(testing.NewUserRealm(member))
	ret := dao.MustCreateProposal(cross(cur), NewReturnRequest(cross(cur), returnTo, Available(), "surplus"))
	dao.MustVoteOnProposalSimple(cross(cur), int64(ret), "YES")
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/stale/taker"))
	taker := Apply(cross(cur), "t", "p", []Tranche{{Deliverable: "all", Amount: Available()}}, 0)
	testing.SetRealm(testing.NewUserRealm(member))
	take := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), taker, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(take), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), take))
	uassert.False(t, dao.ExecuteOrRejectProposal(cross(cur), ret))

	// An approval whose request was superseded meanwhile.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/stale/changer"))
	first := Apply(cross(cur), "t", "p", tranches(), 30*day)
	testing.SetRealm(testing.NewUserRealm(member))
	old := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), first, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(old), "YES")
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/stale/changer"))
	Apply(cross(cur), "t", "p again", tranches(), 30*day)
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.False(t, dao.ExecuteOrRejectProposal(cross(cur), old))
}

// A Return sends exactly its amount to the treasury and leaves committed
// funds alone.
func TestReturnSendsExactlyTheAmount(cur realm, t *testing.T) {
	seatMember(cur)
	testing.IssueCoins(Address(), coins(Committed()+700))
	treasury := returnTo
	before := banker.NewReadonlyBanker().GetCoin(treasury, Denom)
	committed := Committed()

	testing.SetRealm(testing.NewUserRealm(member))
	pid := dao.MustCreateProposal(cross(cur), NewReturnRequest(cross(cur), returnTo, 300, "surplus"))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), pid))
	uassert.Equal(t, before+300, banker.NewReadonlyBanker().GetCoin(treasury, Denom))
	uassert.Equal(t, committed, Committed())
}
