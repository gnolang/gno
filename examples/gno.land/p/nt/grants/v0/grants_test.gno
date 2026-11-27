package grants

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

const (
	golf  = "gno.land/r/test/golf"
	other = "gno.land/r/test/other"
	day   = 24 * time.Hour
	month = 30 * day
)

var (
	t0    = time.Unix(1_800_000_000, 0)
	rules = Rules{Window: 14 * day, MinPeriod: month, Expiry: 60 * day, HoldFor: 14 * day, Advance: 25}
)

func at(d time.Duration) time.Time { return t0.Add(d) }

// two is a grant whose first payment, a quarter of it, is an advance under
// the test rules.
func two() []Tranche {
	return []Tranche{{"start v3", 100}, {"v3 live with 100 players", 300}}
}

// three is a grant whose first payment is an advance too.
func three() []Tranche {
	return []Tranche{{"a", 1}, {"b", 2}, {"c", 1}}
}

func program(t *testing.T) *Program {
	p, err := NewProgram(rules)
	urequire.NoError(t, err)
	return p
}

func apply(t *testing.T, p *Program, grantee string, ts []Tranche, now time.Time) int {
	id, err := p.Apply(grantee, "Gnogolf v3", "More players on gno.land.", ts, month, now)
	urequire.NoError(t, err)
	return id
}

// approved returns a program holding one approved two-tranche grant.
func approved(t *testing.T) (*Program, int) {
	p := program(t)
	id := apply(t, p, golf, two(), t0)
	urequire.NoError(t, p.Approve(id, 400, "", "", t0))
	return p, id
}

func claim(t *testing.T, p *Program, id int, now time.Time) int64 {
	_, amount, err := p.Claim(id, now)
	urequire.NoError(t, err)
	return amount
}

func status(p *Program, id int) string {
	g, _ := p.Get(id)
	return g.Status.String()
}

func TestNewProgramRefusesBadRules(t *testing.T) {
	_, err := NewProgram(Rules{Expiry: 0})
	uassert.Error(t, err)
	_, err = NewProgram(Rules{Expiry: day, MinPeriod: MaxSchedule + 1})
	uassert.Error(t, err)
	_, err = NewProgram(Rules{Expiry: day, Advance: 101})
	uassert.Error(t, err)
}

func TestApplyValidates(t *testing.T) {
	p := program(t)
	tooMany := make([]Tranche, MaxTranches+1)
	for i := range tooMany {
		tooMany[i] = Tranche{"d", 1}
	}
	cases := []struct {
		name, grantee, title, pitch string
		tranches                    []Tranche
		period                      time.Duration
		want                        string
	}{
		{"no grantee", "", "t", "p", two(), month, "grantee"},
		{"blank title", golf, "  ", "p", two(), month, "title"},
		{"invisible title", golf, "\u200b\u200e \u2066", "p", two(), month, "title"},
		{"long title", golf, strings.Repeat("x", MaxTitle+1), "p", two(), month, "title"},
		{"multiline title", golf, "a\nb", "p", two(), month, "single line"},
		{"blank pitch", golf, "t", "", two(), month, "pitch"},
		{"no tranche", golf, "t", "p", nil, month, "tranches"},
		{"too many tranches", golf, "t", "p", tooMany, month, "tranches"},
		{"zero amount", golf, "t", "p", []Tranche{{"d", 0}}, 0, "amount"},
		{"negative amount", golf, "t", "p", []Tranche{{"d", -5}}, 0, "amount"},
		{"overflow", golf, "t", "p", []Tranche{{"a", maxTotal}, {"b", 1}}, month, "amount"},
		{"blank deliverable", golf, "t", "p", []Tranche{{" ", 1}}, 0, "deliverable"},
		{"invisible deliverable", golf, "t", "p", []Tranche{{"\ufeff", 1}}, 0, "deliverable"},
		{"period under minimum", golf, "t", "p", two(), month - 1, "period"},
		{"schedule over a year", golf, "t", "p", []Tranche{{"a", 1}, {"b", 1}, {"c", 1}}, 200 * day, "period"},
		{"huge period", golf, "t", "p", two(), time.Duration(1 << 62), "period"},
	}
	for _, c := range cases {
		_, err := p.Apply(c.grantee, c.title, c.pitch, c.tranches, c.period, t0)
		uassert.ErrorContains(t, err, c.want, c.name)
	}
	uassert.Equal(t, 0, p.Size())

	// One payment needs no schedule.
	id, err := p.Apply(golf, "t", "p", []Tranche{{"d", 1}}, 0, t0)
	urequire.NoError(t, err)
	g, _ := p.Get(id)
	uassert.Equal(t, int64(0), int64(g.Period))
}

// Neither the caller's slice nor a returned copy reaches the stored terms.
func TestTermsCannotBeChangedFromOutside(t *testing.T) {
	p := program(t)
	in := two()
	id := apply(t, p, golf, in, t0)
	in[0].Amount = 1_000_000
	g, _ := p.Get(id)
	g.Tranches[1].Amount = 1_000_000
	for _, l := range [][]Grant{p.Page(0, 1), p.Waiting(0, 1), p.Of(golf, 1)} {
		l[0].Tranches[0].Amount = 1_000_000
	}
	g, _ = p.Get(id)
	uassert.Equal(t, int64(400), g.Total())
	uassert.Equal(t, 1, id)
}

// A grantee has one request waiting at a time; a new one replaces it. A
// sub-realm identity is a grantee of its own: one DAO hosted by a realm
// cannot push another's request out.
func TestApplySupersedesPendingOfTheSameGrantee(t *testing.T) {
	p := program(t)
	a := apply(t, p, golf, two(), t0)
	d1 := apply(t, p, golf+"#dao1", two(), t0)
	d2 := apply(t, p, golf+"#dao2", two(), t0)
	b := apply(t, p, golf, two(), at(day))
	uassert.Equal(t, "superseded", status(p, a))
	uassert.Equal(t, "pending", status(p, d1))
	uassert.Equal(t, "pending", status(p, d2))
	uassert.Equal(t, "pending", status(p, b))
	g, _ := p.Get(a)
	uassert.Equal(t, "superseded by #4", g.Reason)
	uassert.ErrorIs(t, p.Approve(a, 1000, "", "", at(day)), ErrNotPending)
	uassert.Equal(t, 3, p.WaitingCount())

	// An active grant is not superseded: a team can line up its next budget.
	urequire.NoError(t, p.Approve(b, 400, "", "", at(day)))
	c := apply(t, p, golf, two(), at(2*day))
	uassert.Equal(t, "active", status(p, b))
	uassert.Equal(t, "pending", status(p, c))
	uassert.Equal(t, "gno.land/r/test/golf", Host(golf+"#dao1"))
}

func TestApproveReservesAndRefuses(t *testing.T) {
	p := program(t)
	a := apply(t, p, golf, two(), t0)
	b := apply(t, p, other, two(), t0)

	uassert.ErrorIs(t, p.Approve(a, 399, "", "", t0), ErrUnderfunded)
	urequire.NoError(t, p.Approve(a, 400, "", "", t0))
	uassert.Equal(t, int64(400), p.Committed())

	// 700 held, 400 already promised: b's 400 does not fit.
	uassert.ErrorIs(t, p.Approve(b, 700, "", "", t0), ErrUnderfunded)
	urequire.NoError(t, p.Approve(b, 800, "", "", t0))
	uassert.Equal(t, int64(800), p.Committed())

	uassert.ErrorIs(t, p.Approve(a, 10_000, "", "", t0), ErrNotPending)
	uassert.ErrorIs(t, p.Approve(99, 10_000, "", "", t0), ErrNotFound)

	c := apply(t, p, "gno.land/r/test/late", two(), t0)
	uassert.ErrorIs(t, p.Approve(c, 10_000, "", "", at(60*day)), ErrExpired)
	uassert.Equal(t, "pending", status(p, c))
}

// The whole path: tranche 0 a window after the vote; tranche 1 after a
// report, a window, and a full period since tranche 0 was paid.
func TestPaymentSchedule(t *testing.T) {
	p, id := approved(t)

	_, _, err := p.Claim(id, at(14*day-time.Second))
	uassert.ErrorIs(t, err, ErrNotDue)
	uassert.Equal(t, int64(100), claim(t, p, id, at(14*day)))
	uassert.Equal(t, int64(300), p.Committed())

	_, _, err = p.Claim(id, at(9*month))
	uassert.ErrorIs(t, err, ErrNotReported)

	uassert.ErrorIs(t, p.Report(other, id, "done", "", t0), ErrNotGrantee)
	urequire.NoError(t, p.Report(golf, id, "v3 live: https://example.com/v3", "", at(20*day)))

	// Window passed (day 34), period since payment 0 not yet (day 44).
	_, _, err = p.Claim(id, at(43*day))
	uassert.ErrorIs(t, err, ErrNotDue)
	uassert.Equal(t, int64(300), claim(t, p, id, at(44*day)))

	g, _ := p.Get(id)
	uassert.Equal(t, "completed", g.Status.String())
	uassert.Equal(t, int64(0), p.Committed())
	uassert.Equal(t, int64(400), g.Paid())
	uassert.Equal(t, 1, g.Reports[0].Tranche)
	uassert.Equal(t, 0, p.ActiveCount()+p.WaitingCount())

	_, _, err = p.Claim(id, at(9*month))
	uassert.ErrorIs(t, err, ErrNotActive)
}

// A grantee behind schedule still gets at most one payment per period:
// spacing counts from the previous payment, not from approval.
func TestLateGranteeIsStillSpaced(t *testing.T) {
	p := program(t)
	id := apply(t, p, golf, three(), t0)
	urequire.NoError(t, p.Approve(id, 4, "", "", t0))
	claim(t, p, id, at(89*day))
	urequire.NoError(t, p.Report(golf, id, "b", "", at(89*day)))
	claim(t, p, id, at(119*day))
	urequire.NoError(t, p.Report(golf, id, "c", "", at(119*day)))
	_, _, err := p.Claim(id, at(133*day))
	uassert.ErrorIs(t, err, ErrNotDue)
	claim(t, p, id, at(149*day))
}

// A later report restarts the window.
func TestReReportRestartsWindow(t *testing.T) {
	p, id := approved(t)
	claim(t, p, id, at(14*day))
	urequire.NoError(t, p.Report(golf, id, "first", "", at(40*day)))
	urequire.NoError(t, p.Report(golf, id, "correction", "", at(50*day)))
	_, _, err := p.Claim(id, at(54*day))
	uassert.ErrorIs(t, err, ErrNotDue)
	claim(t, p, id, at(64*day))
}

// A hold always pushes the payment back, whenever it is placed: one placed
// early, before the report, is not used up for nothing.
func TestHoldAlwaysDelays(t *testing.T) {
	p, id := approved(t)
	claim(t, p, id, at(14*day))

	urequire.NoError(t, p.Hold(id, 5, at(15*day)))
	uassert.ErrorIs(t, p.Hold(id, 6, at(16*day)), ErrHeld)
	urequire.NoError(t, p.Report(golf, id, "done", "", at(40*day)))
	pay, _ := p.PayableAt(id)
	uassert.True(t, pay.Equal(at(54*day+14*day)), "report + window + hold")

	// The next tranche can be held again, never with a used reference.
	claim(t, p, id, at(68*day))
	q, qid := approved(t)
	urequire.NoError(t, q.Hold(qid, 3, t0))
	claim(t, q, qid, at(28*day))
	uassert.ErrorIs(t, q.Hold(qid, 3, at(29*day)), ErrStaleHold)
	urequire.NoError(t, q.Hold(qid, 4, at(29*day)))

	r := program(t)
	pending := apply(t, r, golf, two(), t0)
	uassert.ErrorIs(t, r.Hold(pending, 1, t0), ErrNotActive)
}

func TestCancelReleasesWhatIsUnpaid(t *testing.T) {
	p, id := approved(t)
	claim(t, p, id, at(14*day))

	uassert.ErrorContains(t, p.Cancel(id, "", t0), "reason")
	urequire.NoError(t, p.Cancel(id, "v3 never shipped", at(20*day)))
	uassert.Equal(t, int64(0), p.Committed())
	g, _ := p.Get(id)
	uassert.Equal(t, "cancelled", g.Status.String())
	uassert.Equal(t, int64(100), g.Paid())

	uassert.ErrorContains(t, p.Cancel(id, "again", t0), "already cancelled")
	uassert.ErrorIs(t, p.Report(golf, id, "late", "", t0), ErrNotActive)

	q := program(t)
	pending := apply(t, q, golf, two(), t0)
	urequire.NoError(t, q.Cancel(pending, "out of scope", t0))
	uassert.Equal(t, int64(0), q.Committed())
	// Its host may apply again without superseding anything.
	again := apply(t, q, golf, two(), t0)
	uassert.Equal(t, "cancelled", status(q, pending))
	uassert.Equal(t, "pending", status(q, again))
}

func TestExpire(t *testing.T) {
	p := program(t)
	id := apply(t, p, golf, two(), t0)
	uassert.ErrorIs(t, p.Expire(id, at(59*day)), ErrNotExpired)
	urequire.NoError(t, p.Expire(id, at(60*day)))
	uassert.Equal(t, "expired", status(p, id))

	// An advance needs no report, so it never expires.
	p, id = approved(t)
	uassert.ErrorIs(t, p.Expire(id, at(9*month)), ErrNotExpired)

	// A due tranche left unreported expires Expiry after its period, and
	// gives its reservation back.
	claim(t, p, id, at(14*day))
	uassert.ErrorIs(t, p.Expire(id, at(14*day+month+59*day)), ErrNotExpired)
	urequire.NoError(t, p.Expire(id, at(14*day+month+60*day)))
	uassert.Equal(t, int64(0), p.Committed())

	// A report, even late, keeps it alive.
	p, id = approved(t)
	claim(t, p, id, at(14*day))
	urequire.NoError(t, p.Report(golf, id, "late but done", "", at(100*day)))
	uassert.ErrorIs(t, p.Expire(id, at(200*day)), ErrNotExpired)

	// Past the deadline, a report cannot take the expiry back.
	p, id = approved(t)
	claim(t, p, id, at(14*day))
	uassert.ErrorIs(t, p.Report(golf, id, "too late", "", at(14*day+month+60*day)), ErrExpired)
	urequire.NoError(t, p.Expire(id, at(14*day+month+60*day)))
}

// Claiming a late payment never leaves the next one already expired: its
// clock starts at that payment.
func TestClaimThenExpireIsRefused(t *testing.T) {
	p := program(t)
	id := apply(t, p, golf, three(), t0)
	urequire.NoError(t, p.Approve(id, 4, "", "", t0))
	claim(t, p, id, at(14*day))
	urequire.NoError(t, p.Report(golf, id, "b", "", at(89*day)))
	claim(t, p, id, at(120*day))
	uassert.ErrorIs(t, p.Expire(id, at(120*day)), ErrNotExpired)
}

// Close cancels the active grants, which only the grantor can create, and
// leaves pending requests alone: nothing can approve them any more and they
// expire as usual. However many requests are filed, Close walks only what
// the grantor approved.
func TestClose(t *testing.T) {
	p, id := approved(t)
	claim(t, p, id, at(14*day))
	pending := apply(t, p, other, two(), t0)
	for i := 0; i < 30; i++ {
		apply(t, p, golf+"#spam"+strconv.Itoa(i), two(), t0)
	}

	uassert.ErrorContains(t, p.Close(" ", true, t0), "reason")
	urequire.NoError(t, p.Close("moving to v1", true, at(20*day)))
	uassert.Equal(t, int64(0), p.Committed())
	uassert.Equal(t, 0, p.ActiveCount())
	uassert.Equal(t, "cancelled", status(p, id))
	uassert.Equal(t, "pending", status(p, pending))
	uassert.ErrorIs(t, p.Approve(pending, 10_000, "", "", at(20*day)), ErrClosed)
	urequire.NoError(t, p.Expire(pending, at(60*day)))

	_, err := p.Apply(golf, "t", "p", two(), month, t0)
	uassert.ErrorIs(t, err, ErrClosed)
	uassert.ErrorIs(t, p.Close("again", false, t0), ErrClosed)
}

// Closed without cancelling, the program takes nothing new but running
// grants keep their reservation and their schedule.
func TestCloseKeepsRunningGrants(t *testing.T) {
	p, id := approved(t)
	pending := apply(t, p, other, two(), t0)
	urequire.NoError(t, p.Close("moving to v1", false, at(day)))
	uassert.True(t, p.Closed())
	uassert.Equal(t, int64(400), p.Committed())
	uassert.ErrorIs(t, p.Approve(pending, 10_000, "", "", at(day)), ErrClosed)
	claim(t, p, id, at(14*day))
	urequire.NoError(t, p.Report(golf, id, "done", "", at(20*day)))
	claim(t, p, id, at(44*day))
	uassert.Equal(t, "completed", status(p, id))
	uassert.Equal(t, int64(0), p.Committed())
}

// A first payment above the advance share waits for a report like the
// others, and a single-payment grant is paid only once its work is reported.
func TestFirstPaymentNeedsReportUnlessAdvance(t *testing.T) {
	p := program(t)
	id := apply(t, p, golf, []Tranche{{"a", 26}, {"b", 74}}, t0)
	one := apply(t, p, other, []Tranche{{"shipped v3", 50}}, t0)
	g, _ := p.Get(id)
	uassert.False(t, g.Advance)
	urequire.NoError(t, p.Approve(id, 150, "", "", t0))
	urequire.NoError(t, p.Approve(one, 150, "", "", t0))

	_, _, err := p.Claim(id, at(9*month))
	uassert.ErrorIs(t, err, ErrNotReported)
	_, _, err = p.Claim(one, at(9*month))
	uassert.ErrorIs(t, err, ErrNotReported)

	// Unreported, it expires Expiry after its period from approval.
	uassert.ErrorIs(t, p.Expire(id, at(month+60*day-time.Second)), ErrNotExpired)
	uassert.ErrorIs(t, p.Expire(one, at(60*day-time.Second)), ErrNotExpired)
	urequire.NoError(t, p.Expire(one, at(60*day)))

	urequire.NoError(t, p.Report(golf, id, "a done", "/r/test/golf", at(10*day)))
	_, _, err = p.Claim(id, at(24*day-time.Second))
	uassert.ErrorIs(t, err, ErrNotDue)
	uassert.Equal(t, int64(26), claim(t, p, id, at(24*day)))

	// No rule, no advance.
	q, err := NewProgram(Rules{Window: day, Expiry: day})
	urequire.NoError(t, err)
	qid := apply(t, q, golf, []Tranche{{"a", 1}, {"b", 99}}, t0)
	g, _ = q.Get(qid)
	uassert.False(t, g.Advance)
}

// Every payment needs its own report: one report never pays two tranches,
// and a request cannot be reported on before it is approved.
func TestEachPaymentNeedsItsOwnReport(t *testing.T) {
	p := program(t)
	id := apply(t, p, golf, three(), t0)
	uassert.ErrorIs(t, p.Report(golf, id, "early", "", t0), ErrNotActive)
	urequire.NoError(t, p.Approve(id, 4, "", "", t0))
	claim(t, p, id, at(14*day))
	urequire.NoError(t, p.Report(golf, id, "b and c done", "", at(20*day)))
	claim(t, p, id, at(44*day))
	_, _, err := p.Claim(id, at(9*month))
	uassert.ErrorIs(t, err, ErrNotReported)
	urequire.NoError(t, p.Report(golf, id, "c done", "", at(80*day)))
	claim(t, p, id, at(94*day))
}

// A report on an unclaimed advance would be discarded by Claim, so it is
// refused; a grantor's reason never reads as the grantee's withdrawal; and a
// program closed without cancelling can still cancel everything later.
func TestGuards(t *testing.T) {
	p, id := approved(t)
	uassert.ErrorIs(t, p.Report(golf, id, "early", "", t0), ErrAdvance)
	uassert.ErrorContains(t, p.Cancel(id, WithdrawnReason, t0), "reserved")
	uassert.ErrorContains(t, p.Close(WithdrawnReason, true, t0), "reserved")
	uassert.False(t, p.Closed())

	urequire.NoError(t, p.Close("pause intake", false, t0))
	uassert.ErrorIs(t, p.Close("again", false, t0), ErrClosed)
	urequire.NoError(t, p.Close("stop everything", true, at(day)))
	uassert.Equal(t, "cancelled", status(p, id))
	uassert.Equal(t, int64(0), p.Committed())
}

// The grantee can give a grant up, pending or active; what it was not paid
// goes back.
func TestWithdraw(t *testing.T) {
	p, id := approved(t)
	claim(t, p, id, at(14*day))
	uassert.ErrorIs(t, p.Withdraw(other, id, at(15*day)), ErrNotGrantee)
	urequire.NoError(t, p.Withdraw(golf, id, at(15*day)))
	g, _ := p.Get(id)
	uassert.Equal(t, "cancelled", g.Status.String())
	uassert.Equal(t, "withdrawn by the grantee", g.Reason)
	uassert.Equal(t, int64(0), p.Committed())
	uassert.ErrorContains(t, p.Withdraw(golf, id, at(16*day)), "already")

	pending := apply(t, p, golf, two(), at(16*day))
	urequire.NoError(t, p.Withdraw(golf, pending, at(16*day)))
	again := apply(t, p, golf, two(), at(17*day))
	uassert.Equal(t, "pending", status(p, again))
}

// Who sponsored a grant and what they checked stay on it.
func TestApproveRecordsTheSponsor(t *testing.T) {
	p := program(t)
	id := apply(t, p, golf, two(), t0)
	uassert.ErrorContains(t, p.Approve(id, 400, "g1x", strings.Repeat("x", MaxReview+1), t0), "review")
	uassert.ErrorContains(t, p.Approve(id, 400, strings.Repeat("g", MaxLink+1), "ok", t0), "sponsor")
	urequire.NoError(t, p.Approve(id, 400, "g1sponsor", "Code: hosted wallet. Owner: known.", t0))
	g, _ := p.Get(id)
	uassert.Equal(t, "g1sponsor", g.Sponsor)
	uassert.Equal(t, "Code: hosted wallet. Owner: known.", g.Review)
}

// A report link is a gno.land path or an https URL that cannot break out of
// a markdown link.
func TestReportLink(t *testing.T) {
	p, id := approved(t)
	claim(t, p, id, at(14*day))
	for _, ok := range []string{"", "/r/test/golf:v3", "https://gno.land/r/test/golf?x=1#top"} {
		uassert.NoError(t, p.Report(golf, id, "r", ok, at(15*day)), ok)
		p, id = approved(t)
		claim(t, p, id, at(14*day))
	}
	for _, bad := range []string{"javascript:alert(1)", "//evil.com", "http://x.com", "https://x.com/a b",
		"https://x.com/)[y](z", "/r/x\"", "https://x.com/<b>", strings.Repeat("/x", MaxLink),
		"/r/gov/dao$help&func=MustVoteOnProposalSimple&pid=1&option=YES"} {
		uassert.ErrorContains(t, p.Report(golf, id, "r", bad, at(15*day)), "link", bad)
	}
	urequire.NoError(t, p.Report(golf, id, "r", "/r/test/golf", at(15*day)))
	g, _ := p.Get(id)
	uassert.Equal(t, "/r/test/golf", g.Reports[0].Link)
}

// Every hold stays on the record, one per payment.
func TestHoldsAreKept(t *testing.T) {
	p, id := approved(t)
	urequire.NoError(t, p.Hold(id, 1, t0))
	claim(t, p, id, at(28*day))
	urequire.NoError(t, p.Report(golf, id, "done", "", at(40*day)))
	urequire.NoError(t, p.Hold(id, 2, at(41*day)))
	g, _ := p.Get(id)
	urequire.Equal(t, 2, len(g.Holds))
	uassert.Equal(t, 0, g.Holds[0].Tranche)
	uassert.Equal(t, int64(1), g.Holds[0].Ref)
	uassert.Equal(t, 1, g.Holds[1].Tranche)
	uassert.True(t, g.Held())
}

func TestProseIsBoundedInLines(t *testing.T) {
	p, id := approved(t)
	claim(t, p, id, at(14*day))
	_, err := p.Apply(other, "t", strings.Repeat("\n", MaxLines)+"x", two(), month, t0)
	uassert.ErrorContains(t, err, "lines")
	uassert.ErrorContains(t, p.Report(golf, id, "x"+strings.Repeat("\n", MaxLines), "", t0), "lines")
	urequire.NoError(t, p.Report(golf, id, "x"+strings.Repeat("\n", MaxLines-1), "", t0))
}

// Of lists exactly one grantee, newest first; ByGrantee stops counting at
// MaxListed.
func TestOfAndCountingCap(t *testing.T) {
	p := program(t)
	for i := 0; i < 25; i++ {
		apply(t, p, golf, two(), t0)
	}
	apply(t, p, golf+"#dao", two(), t0)
	list := p.Of(golf, 20)
	urequire.Equal(t, 20, len(list))
	uassert.Equal(t, 25, list[0].ID)
	for _, g := range list {
		uassert.Equal(t, golf, g.Grantee)
	}

	for i := 0; i < MaxListed; i++ {
		apply(t, p, "gno.land/r/many/x"+strconv.Itoa(i), []Tranche{{"d", 1}}, t0)
	}
	_, n := p.ByGrantee("gno.land/r/many", 0, 0)
	uassert.Equal(t, MaxListed, n)
}

// Reports are capped per payment, so using them all up early never stops a
// grantee from reporting the next one.
func TestReportLimits(t *testing.T) {
	p, id := approved(t)
	uassert.ErrorContains(t, p.Report(golf, id, " ", "", t0), "report")
	uassert.ErrorContains(t, p.Report(golf, id, strings.Repeat("x", MaxReport+1), "", t0), "report")
	claim(t, p, id, at(14*day))
	for i := 0; i < MaxReportsPerPayment; i++ {
		urequire.NoError(t, p.Report(golf, id, "r", "", at(20*day)))
	}
	uassert.ErrorIs(t, p.Report(golf, id, "one more", "", at(20*day)), ErrReportsFull)

	q := program(t)
	qid := apply(t, q, golf, three(), t0)
	urequire.NoError(t, q.Approve(qid, 4, "", "", t0))
	claim(t, q, qid, at(14*day))
	for i := 0; i < MaxReportsPerPayment; i++ {
		urequire.NoError(t, q.Report(golf, qid, "b", "", at(20*day)))
	}
	claim(t, q, qid, at(44*day))
	urequire.NoError(t, q.Report(golf, qid, "c", "", at(50*day)))
}

// A realm path lists that realm, its sub-realms and its sub-realm
// identities, never a sibling whose path merely starts the same way.
func TestListing(t *testing.T) {
	p := program(t)
	for _, g := range []string{golf, golf + "2", golf + "/v2", golf + "#dao", "gno.land/r/test", "gno.land/r/team/x", "gno.land/r/testnet/y"} {
		_, err := p.Apply(g, "t", "p", two(), month, t0)
		urequire.NoError(t, err)
	}

	page := p.Page(1, 2)
	urequire.Equal(t, 2, len(page))
	uassert.Equal(t, 6, page[0].ID)
	uassert.Equal(t, 5, page[1].ID)

	_, n := p.ByGrantee(golf, 0, 0)
	uassert.Equal(t, 3, n)
	_, n = p.ByGrantee("gno.land/r/test", 0, 0)
	uassert.Equal(t, 5, n)
	list, n := p.ByGrantee("gno.land/r/test", 1, 2)
	uassert.Equal(t, 5, n)
	uassert.Equal(t, 2, len(list))

	// A namespace without the identities a realm in it hosts.
	list, n = p.ByGranteeExcept("gno.land/r/test", golf, 0, 10)
	uassert.Equal(t, 4, n)
	for _, g := range list {
		uassert.NotEqual(t, golf+"#dao", g.Grantee)
	}
	_, n = p.ByGranteeExcept(golf, "gno.land/r/team/x", 0, 0)
	uassert.Equal(t, 3, n)

	_, ok := p.Get(0)
	uassert.False(t, ok)
}

func TestParseTranches(t *testing.T) {
	ts, err := ParseTranches("start v3: 100; port gno:land tooling:250;")
	urequire.NoError(t, err)
	urequire.Equal(t, 2, len(ts))
	uassert.Equal(t, "start v3", ts[0].Deliverable)
	uassert.Equal(t, int64(100), ts[0].Amount)
	uassert.Equal(t, "port gno:land tooling", ts[1].Deliverable)
	uassert.Equal(t, int64(250), ts[1].Amount)

	_, err = ParseTranches("no amount")
	uassert.ErrorContains(t, err, `("no amount") must read deliverable:amount`)
	_, err = ParseTranches("x: 1,000")
	uassert.ErrorContains(t, err, `"1,000" is not a whole number`)
	_, err = ParseTranches("")
	uassert.Error(t, err)
}
