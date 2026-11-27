// Package grants is the lifecycle of a grant program whose grantees are
// realms, not people.
//
// A realm asks for an amount split into tranches. The grantor votes once,
// on the whole request. Every tranche needs a report from the grantee on what
// it promised, a review window after that report, and at least one period
// since the previous payment. The one exception is a small first tranche, an
// advance within the share the rules allow, paid a review window after
// approval. The grantor can cancel at any time, and put a hold that pushes
// the next payment back while it decides; the grantee can withdraw. A grant
// left unattended expires and gives its reservation back. The program can be
// closed to new requests, with or without cancelling the running grants.
//
// The package is pure. It holds no coins and reads no chain state. The realm
// on top derives the grantee with GranteeOf (never from input), supplies the
// time and the balance it holds, and performs the transfer Claim hands back.
// Approve refuses a grant the balance cannot fully cover, so a realm that
// moves coins only through Claim can always pay every approved tranche.
// gno.land/r/gov/grants/v0 is GovDAO's desk built on it.
package grants

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
)

// Limits every input is checked against. Exported so a UI or a grantee
// realm can pre-validate.
const (
	MaxTitle       = 80
	MaxPitch       = 4000
	MaxDeliverable = 280
	MaxReport      = 2000
	MaxReason      = 500
	MaxReview      = 1000
	MaxLink        = 200
	// MaxLines bounds the lines of a pitch or a report, so rendering them,
	// line by line, stays cheap.
	MaxLines = 50
	// MaxReportsPerPayment bounds the reports on one payment: the report
	// and two corrections.
	MaxReportsPerPayment = 3
	MaxTranches          = 12
	// MaxSchedule bounds the periods of a grant added together. A grantee
	// that is slower than its schedule takes longer; the grant then stays
	// open until it completes, expires, or the grantor cancels it.
	MaxSchedule = 366 * 24 * time.Hour
	// MaxListed is where ByGrantee stops counting.
	MaxListed = 1000
)

// WithdrawnReason is the Reason of a grant its grantee withdrew.
const WithdrawnReason = "withdrawn by the grantee"

// maxTotal bounds the sum of a request so no arithmetic on it can overflow.
const maxTotal int64 = 1 << 62

var (
	ErrNotFound    = errors.New("grants: no such grant")
	ErrNotGrantee  = errors.New("grants: only the grantee realm may do this")
	ErrNotPending  = errors.New("grants: grant is not pending")
	ErrNotActive   = errors.New("grants: grant is not active")
	ErrUnderfunded = errors.New("grants: the pool cannot cover this grant")
	ErrNotReported = errors.New("grants: the next tranche has no report yet")
	ErrNotDue      = errors.New("grants: the next tranche is not payable yet")
	ErrReportsFull = errors.New("grants: report limit reached for this payment")
	ErrExpired     = errors.New("grants: grant has expired")
	ErrNotExpired  = errors.New("grants: grant has not expired")
	ErrHeld        = errors.New("grants: this tranche was already held once")
	ErrStaleHold   = errors.New("grants: hold reference already used")
	ErrClosed      = errors.New("grants: the program is closed")
	ErrNotRealm    = errors.New("grants: grantees are realms; call from your realm's code")
	ErrAdvance     = errors.New("grants: the advance needs no report; claim it, then report on payment 2")
)

// Rules are the grantor's terms, fixed when the program is created.
type Rules struct {
	Window    time.Duration // after a report (or approval, for an advance), before payment
	MinPeriod time.Duration // minimum spacing between two payments of a grant
	Expiry    time.Duration // a pending grant, or an unreported tranche, expires after this
	HoldFor   time.Duration // how far a hold pushes the next payment back
	Advance   int           // percent of a grant's total its first tranche may be to be paid without a report; 0 for none
}

// Status is where a grant sits in its lifecycle.
type Status int

const (
	Pending    Status = iota // filed, waiting for the grantor
	Active                   // approved and funded; tranches are being paid
	Completed                // every tranche paid
	Cancelled                // stopped by the grantor
	Superseded               // replaced by a newer request from the same realm
	Expired                  // left unattended past the expiry
)

func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Active:
		return "active"
	case Completed:
		return "completed"
	case Cancelled:
		return "cancelled"
	case Superseded:
		return "superseded"
	case Expired:
		return "expired"
	}
	return "unknown"
}

// Tranche is one payment of a grant and what it pays for. Amounts are in the
// smallest unit of whatever the realm pays in.
type Tranche struct {
	Deliverable string
	Amount      int64
}

// Report is a progress statement the grantee posted on one tranche, with an
// optional link to what it shows: a gno.land path ("/r/...") or an https URL.
type Report struct {
	Tranche int
	Body    string
	Link    string
	At      int64 // unix seconds
}

// Hold is a pause the grantor put on one tranche.
type Hold struct {
	Tranche int
	At      int64 // unix seconds
	Ref     int64 // the evidence it was placed with, a proposal id at the GovDAO desk
}

// Grant is one request and everything that happened to it. Its terms
// (Title, Pitch, Tranches, Period) never change after Apply: what the
// grantor approves is exactly what was filed. Times are unix seconds, zero
// when unset: a time.Time is a stored object of its own.
type Grant struct {
	ID       int
	Grantee  string // package path of the grantee realm, "host#sub" for a sub-realm identity
	Title    string
	Pitch    string
	Tranches []Tranche
	Period   time.Duration // minimum time between two payments
	Advance  bool          // tranche 0 is paid without a report

	Sponsor string // who put the grant to the grantor, recorded on approval
	Review  string // what the sponsor checked, recorded on approval

	Status     Status
	Reason     string   // why it was cancelled, withdrawn or superseded
	Reports    []Report // append-only
	PaidAt     []int64  // PaidAt[i] is when tranche i was paid
	ReportedAt int64    // latest report on the next tranche
	Holds      []Hold   // append-only, at most one per tranche

	AppliedAt  int64
	ApprovedAt int64
	ClosedAt   int64
}

// Total is the sum of every tranche.
func (g Grant) Total() int64 {
	var t int64
	for _, tr := range g.Tranches {
		t += tr.Amount
	}
	return t
}

// Paid is what the tranches released so far add up to.
func (g Grant) Paid() int64 {
	var t int64
	for i := range g.PaidAt {
		t += g.Tranches[i].Amount
	}
	return t
}

// Held reports whether the next tranche of an active grant is under a hold.
func (g Grant) Held() bool {
	n := len(g.Holds)
	return g.Status == Active && n > 0 && g.Holds[n-1].Tranche == len(g.PaidAt)
}

// GranteeOf is the realm that crossed into the caller, the only identity a
// desk should accept as a grantee. rlm is the desk's own cur. Users and
// MsgRun scripts are refused; a sub-realm identity ("host#sub", minted with
// cur.Sub, as a DAO hosted by a realm has) is accepted as a grantee of its
// own, and its host's code is what a reviewer reads.
func GranteeOf(_ int, rlm realm) (string, error) {
	if !rlm.IsCurrent() {
		return "", errors.New("grants: rlm is not the caller's live cur")
	}
	prev := rlm.Previous()
	if !prev.IsCode() || prev.IsEphemeral() {
		return "", ErrNotRealm
	}
	return prev.PkgPath(), nil
}

// Host is the realm whose code runs for a grantee: the path itself, or the
// host of a sub-realm identity.
func Host(grantee string) string {
	if i := strings.IndexByte(grantee, '#'); i >= 0 {
		return grantee[:i]
	}
	return grantee
}

// Program is every grant of one grantor, and what the grantor has promised.
type Program struct {
	rules     Rules
	grants    *avl.Tree // seqid key -> *Grant
	waiting   *avl.Tree // seqid key -> id, pending grants
	active    *avl.Tree // seqid key -> id, active grants
	pending   *avl.Tree // grantee -> id of its pending grant
	byGrantee *avl.Tree // grantee + "\x00" + seqid key -> id
	next      seqid.ID
	committed int64 // unpaid tranches of active grants
	closed    bool
}

// NewProgram returns an empty program run under the given rules.
func NewProgram(r Rules) (*Program, error) {
	if r.Window < 0 || r.MinPeriod < 0 || r.MinPeriod > MaxSchedule || r.Expiry <= 0 || r.HoldFor < 0 ||
		r.Advance < 0 || r.Advance > 100 {
		return nil, errors.New("grants: invalid rules")
	}
	return &Program{
		rules:     r,
		grants:    avl.NewTree(),
		waiting:   avl.NewTree(),
		active:    avl.NewTree(),
		pending:   avl.NewTree(),
		byGrantee: avl.NewTree(),
	}, nil
}

// Rules returns the program's rules.
func (p *Program) Rules() Rules { return p.rules }

// Committed is what active grants can still claim. A realm whose balance
// stays at or above it can pay every approved tranche.
func (p *Program) Committed() int64 { return p.committed }

// Closed reports whether the program was closed to new requests and approvals.
func (p *Program) Closed() bool { return p.closed }

// Size is the number of grants ever filed.
func (p *Program) Size() int { return p.grants.Size() }

// ActiveCount is the number of active grants.
func (p *Program) ActiveCount() int { return p.active.Size() }

// WaitingCount is the number of pending grants.
func (p *Program) WaitingCount() int { return p.waiting.Size() }

// Get returns a copy of the grant, and false when there is none.
func (p *Program) Get(id int) (Grant, bool) {
	g := p.get(id)
	if g == nil {
		return Grant{}, false
	}
	return g.clone(), true
}

// Page returns up to count grants, newest first, skipping offset.
func (p *Program) Page(offset, count int) []Grant {
	return p.collect(p.grants, offset, count, true)
}

// Active returns up to count active grants, oldest first. Only the grantor
// makes a grant active, so this list cannot be flooded by anyone else.
func (p *Program) Active(offset, count int) []Grant {
	return p.collect(p.active, offset, count, false)
}

// Waiting returns up to count pending grants, oldest first: the next to
// expire come first, and a flood of new requests sinks to the end.
func (p *Program) Waiting(offset, count int) []Grant {
	return p.collect(p.waiting, offset, count, false)
}

// PendingOf is the id of grantee's pending request, the one its next Apply
// supersedes, and false when it has none.
func (p *Program) PendingOf(grantee string) (int, bool) {
	id, ok := p.pending.Get(grantee).(int)
	return id, ok
}

// Of returns up to count grants of exactly this grantee, newest first.
func (p *Program) Of(grantee string, count int) []Grant {
	var out []Grant
	p.byGrantee.ReverseIterate(grantee+"\x00", grantee+"\x01", func(_ string, v any) bool {
		if len(out) >= count {
			return true
		}
		out = append(out, p.get(v.(int)).clone())
		return false
	})
	return out
}

// ByGrantee returns up to count grants, skipping offset, of the realm at
// path and everything under it (sub-realms and sub-realm identities), in
// path then id order, with how many there are in all, counting up to
// MaxListed. "gno.land/r/name" lists a whole namespace. Siblings that
// merely share a prefix ("gno.land/r/name2") are never walked.
func (p *Program) ByGrantee(path string, offset, count int) ([]Grant, int) {
	return p.ByGranteeExcept(path, "", offset, count)
}

// ByGranteeExcept is ByGrantee without the sub-realm identities of host, a
// realm under path: a namespace listed without the identities a shared realm
// in it hosts for other namespaces.
func (p *Program) ByGranteeExcept(path, host string, offset, count int) ([]Grant, int) {
	var out []Grant
	total := 0
	cb := func(_ string, v any) bool {
		if total >= offset && len(out) < count {
			out = append(out, p.get(v.(int)).clone())
		}
		total++
		return total >= MaxListed
	}
	// The path itself, then its sub-realm identities, then its sub-realms:
	// "\x00", "#" and "/" are the only bytes that may follow it.
	ranges := [][2]string{{path + "\x00", path + "\x01"}, {path + "#", path + "$"}, {path + "/", path + "0"}}
	if strings.HasPrefix(host, path+"/") {
		ranges = append(ranges[:2], [2]string{path + "/", host + "#"}, [2]string{host + "$", path + "0"})
	}
	for _, r := range ranges {
		if total < MaxListed {
			p.byGrantee.Iterate(r[0], r[1], cb)
		}
	}
	return out, total
}

// Apply files a request for grantee, which must come from GranteeOf. Inputs
// are copied. A pending request from the same grantee is superseded: a
// grantee has at most one request waiting for a decision. Sub-realm
// identities are grantees of their own, so one DAO hosted by a realm cannot
// supersede another's request.
func (p *Program) Apply(grantee, title, pitch string, tranches []Tranche, period time.Duration, now time.Time) (int, error) {
	if p.closed {
		return 0, ErrClosed
	}
	if grantee == "" {
		return 0, errors.New("grants: grantee path is required")
	}
	if err := checkLine("title", title, MaxTitle); err != nil {
		return 0, err
	}
	if err := checkProse("pitch", pitch, MaxPitch); err != nil {
		return 0, err
	}
	n := len(tranches)
	if n == 0 || n > MaxTranches {
		return 0, errors.New("grants: 1 to " + strconv.Itoa(MaxTranches) + " tranches required")
	}
	if n == 1 {
		period = 0
	} else if period < p.rules.MinPeriod || period > MaxSchedule || time.Duration(n-1)*period > MaxSchedule {
		return 0, errors.New("grants: period out of range: at least the minimum, and all periods within a year")
	}
	ts := make([]Tranche, n)
	var total int64
	for i, t := range tranches {
		// Not checkLine: scanning every byte costs gas, and every renderer
		// folds a deliverable's newlines anyway.
		if err := checkText("deliverable", t.Deliverable, MaxDeliverable); err != nil {
			return 0, err
		}
		if t.Amount <= 0 || t.Amount > maxTotal-total {
			return 0, errors.New("grants: tranche " + strconv.Itoa(i) + " amount out of range")
		}
		total += t.Amount
		ts[i] = Tranche{Deliverable: t.Deliverable, Amount: t.Amount}
	}

	id := p.next.Next()
	if v, ok := p.pending.Get(grantee).(int); ok {
		old := p.get(v)
		old.Status = Superseded
		old.Reason = "superseded by #" + strconv.Itoa(int(id))
		old.ClosedAt = now.Unix()
		p.waiting.Remove(seqid.ID(v).String())
	}
	g := &Grant{
		ID:        int(id),
		Grantee:   grantee,
		Title:     title,
		Pitch:     pitch,
		Tranches:  ts,
		Period:    period,
		Advance:   ts[0].Amount <= total/100*int64(p.rules.Advance)+total%100*int64(p.rules.Advance)/100,
		Status:    Pending,
		AppliedAt: now.Unix(),
	}
	p.grants.Set(id.String(), g)
	p.waiting.Set(id.String(), g.ID)
	p.pending.Set(grantee, g.ID)
	p.byGrantee.Set(grantee+"\x00"+id.String(), g.ID)
	return g.ID, nil
}

// Approve activates a pending grant and reserves its total. balance is what
// the grantor holds right now; a grant it cannot fully cover on top of what
// is already committed is refused and nothing changes. sponsor and review,
// who put the grant forward and what they checked, are kept on the grant;
// both may be empty.
func (p *Program) Approve(id int, balance int64, sponsor, review string, now time.Time) error {
	if p.closed {
		return ErrClosed
	}
	if len(sponsor) > MaxLink {
		return errors.New("grants: sponsor must be at most " + strconv.Itoa(MaxLink) + " bytes")
	}
	if review != "" {
		if err := CheckReview(review); err != nil {
			return err
		}
	}
	g := p.get(id)
	if g == nil {
		return ErrNotFound
	}
	if g.Status != Pending {
		return ErrNotPending
	}
	if !now.Before(p.ExpiresAt(*g)) {
		return ErrExpired
	}
	total := g.Total()
	if total > balance-p.committed {
		return ErrUnderfunded
	}
	p.committed += total
	g.Status = Active
	g.Sponsor = sponsor
	g.Review = review
	g.ApprovedAt = now.Unix()
	key := seqid.ID(id).String()
	p.waiting.Remove(key)
	p.active.Set(key, id)
	p.pending.Remove(g.Grantee)
	return nil
}

// Report posts the grantee's statement on its next unpaid tranche and
// restarts that tranche's review window. link may be empty. A grant whose
// report is already overdue (see ExpiresAt) takes none.
func (p *Program) Report(grantee string, id int, body, link string, now time.Time) error {
	g, err := p.mine(grantee, id)
	if err != nil {
		return err
	}
	if g.Status != Active {
		return ErrNotActive
	}
	if at := p.ExpiresAt(*g); !at.IsZero() && !now.Before(at) {
		// Past its deadline the grant is anyone's to expire: a late report
		// must not take that back.
		return ErrExpired
	}
	if err := checkProse("report", body, MaxReport); err != nil {
		return err
	}
	if err := checkLink(link); err != nil {
		return err
	}
	next := len(g.PaidAt)
	if next == 0 && g.Advance {
		// Claim would discard it: a report is on the payment it unlocks.
		return ErrAdvance
	}
	onNext := 0
	for i := len(g.Reports) - 1; i >= 0 && g.Reports[i].Tranche == next; i-- {
		onNext++
	}
	if onNext >= MaxReportsPerPayment {
		return ErrReportsFull
	}
	g.Reports = append(g.Reports, Report{Tranche: next, Body: body, Link: link, At: now.Unix()})
	g.ReportedAt = now.Unix()
	return nil
}

// PayableAt is when the next tranche of an active grant can be claimed.
func (p *Program) PayableAt(id int) (time.Time, error) {
	g := p.get(id)
	if g == nil {
		return time.Time{}, ErrNotFound
	}
	at, err := p.payableAt(g)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(at, 0), nil
}

// Claim releases the next tranche when it is due, and returns where it goes
// and how much. The realm must perform that transfer in the same
// transaction; Claim has already recorded it as paid.
func (p *Program) Claim(id int, now time.Time) (grantee string, amount int64, err error) {
	g := p.get(id)
	if g == nil {
		return "", 0, ErrNotFound
	}
	at, err := p.payableAt(g)
	if err != nil {
		return "", 0, err
	}
	if now.Unix() < at {
		return "", 0, ErrNotDue
	}
	amount = g.Tranches[len(g.PaidAt)].Amount
	g.PaidAt = append(g.PaidAt, now.Unix())
	g.ReportedAt = 0
	p.committed -= amount
	if len(g.PaidAt) == len(g.Tranches) {
		p.finish(g, Completed, now)
	}
	return g.Grantee, amount, nil
}

// Cancel stops a pending or active grant and releases what it had not been
// paid yet. Tranches already paid stay paid.
func (p *Program) Cancel(id int, reason string, now time.Time) error {
	g := p.get(id)
	if g == nil {
		return ErrNotFound
	}
	if err := checkReason(reason); err != nil {
		return err
	}
	if err := p.stop(g, Cancelled, now); err != nil {
		return err
	}
	g.Reason = reason
	return nil
}

// ExpiresAt is when a pending grant, or an active grant whose next tranche
// still waits for its report, can be expired: Expiry after the request, or
// after the period that tranche had to be delivered in. It is zero when
// nothing can expire.
func (p *Program) ExpiresAt(g Grant) time.Time {
	switch {
	case g.Status == Pending:
		return time.Unix(g.AppliedAt+secs(p.rules.Expiry), 0)
	case g.Status == Active && g.ReportedAt == 0 && (len(g.PaidAt) > 0 || !g.Advance):
		since := g.ApprovedAt
		if n := len(g.PaidAt); n > 0 {
			since = g.PaidAt[n-1]
		}
		return time.Unix(since+secs(g.Period)+secs(p.rules.Expiry), 0)
	}
	return time.Time{}
}

// Withdraw lets the grantee stop its own pending or active grant. What it was
// not paid goes back, as with Cancel.
func (p *Program) Withdraw(grantee string, id int, now time.Time) error {
	g, err := p.mine(grantee, id)
	if err != nil {
		return err
	}
	if err := p.stop(g, Cancelled, now); err != nil {
		return err
	}
	g.Reason = WithdrawnReason
	return nil
}

// Expire closes a grant nobody attended to, and can be called by anyone.
// An expired active grant gives its unpaid amount back.
func (p *Program) Expire(id int, now time.Time) error {
	g := p.get(id)
	if g == nil {
		return ErrNotFound
	}
	at := p.ExpiresAt(*g)
	if at.IsZero() || now.Before(at) {
		return ErrNotExpired
	}
	return p.stop(g, Expired, now)
}

// Hold pushes the next payment of an active grant back by HoldFor from
// whichever is later, now or when it would have been paid. ref names the
// evidence for the hold and must be greater than any used before on this
// grant; a tranche can be held once.
func (p *Program) Hold(id int, ref int64, now time.Time) error {
	g := p.get(id)
	if g == nil {
		return ErrNotFound
	}
	if g.Status != Active {
		return ErrNotActive
	}
	if n := len(g.Holds); n > 0 {
		if ref <= g.Holds[n-1].Ref {
			return ErrStaleHold
		}
		if g.Holds[n-1].Tranche == len(g.PaidAt) {
			return ErrHeld
		}
	}
	g.Holds = append(g.Holds, Hold{Tranche: len(g.PaidAt), At: now.Unix(), Ref: ref})
	return nil
}

// Close stops the program taking requests and approving them. Running grants
// go on under the same rules unless cancel is set: then every active grant is
// cancelled with reason and Committed drops to zero, so the realm can return
// its whole balance. A program closed without cancelling can be closed again
// with cancel while it still has active grants. Pending requests are left as they are: nothing can
// approve them any more, and they expire like any other.
//
// Only active grants are walked. The grantor alone makes a grant active, so
// nobody else can make Close too costly to run.
func (p *Program) Close(reason string, cancel bool, now time.Time) error {
	if p.closed && (!cancel || p.active.Size() == 0) {
		return ErrClosed
	}
	if err := checkReason(reason); err != nil {
		return err
	}
	p.closed = true
	if !cancel {
		return nil
	}
	var ids []int
	p.active.Iterate("", "", func(_ string, v any) bool {
		ids = append(ids, v.(int))
		return false
	})
	for _, id := range ids {
		g := p.get(id)
		p.stop(g, Cancelled, now)
		g.Reason = reason
	}
	return nil
}

// stop closes a pending or active grant, releasing what it had reserved.
func (p *Program) stop(g *Grant, s Status, now time.Time) error {
	switch g.Status {
	case Active:
		p.committed -= g.Total() - g.Paid()
	case Pending:
		p.pending.Remove(g.Grantee)
	default:
		return errors.New("grants: grant is already " + g.Status.String())
	}
	p.finish(g, s, now)
	return nil
}

func (p *Program) finish(g *Grant, s Status, now time.Time) {
	key := seqid.ID(g.ID).String()
	if g.Status == Active {
		p.active.Remove(key)
	} else {
		p.waiting.Remove(key)
	}
	g.Status = s
	g.ClosedAt = now.Unix()
}

func (p *Program) get(id int) *Grant {
	if id < 1 {
		return nil
	}
	v := p.grants.Get(seqid.ID(id).String())
	if v == nil {
		return nil
	}
	return v.(*Grant)
}

func (p *Program) mine(grantee string, id int) (*Grant, error) {
	g := p.get(id)
	if g == nil {
		return nil, ErrNotFound
	}
	if g.Grantee != grantee {
		return nil, ErrNotGrantee
	}
	return g, nil
}

func (p *Program) collect(t *avl.Tree, offset, count int, newest bool) []Grant {
	var out []Grant
	cb := func(_ string, v any) bool {
		g, ok := v.(*Grant)
		if !ok {
			g = p.get(v.(int))
		}
		out = append(out, g.clone())
		return false
	}
	if newest {
		t.ReverseIterateByOffset(offset, count, cb)
	} else {
		t.IterateByOffset(offset, count, cb)
	}
	return out
}

// payableAt is the unix time the next tranche of an active grant is due.
func (p *Program) payableAt(g *Grant) (int64, error) {
	if g.Status != Active {
		return 0, ErrNotActive
	}
	next := len(g.PaidAt)
	var at int64
	switch {
	case next == 0 && g.Advance:
		at = g.ApprovedAt + secs(p.rules.Window)
	case g.ReportedAt == 0:
		return 0, ErrNotReported
	default:
		at = g.ReportedAt + secs(p.rules.Window)
		if next > 0 {
			if spaced := g.PaidAt[next-1] + secs(g.Period); spaced > at {
				at = spaced
			}
		}
	}
	if g.Held() {
		if h := g.Holds[len(g.Holds)-1].At; h > at {
			at = h
		}
		at += secs(p.rules.HoldFor)
	}
	return at, nil
}

func secs(d time.Duration) int64 { return int64(d / time.Second) }

// clone copies the grant with its own slices, so nothing a caller does to
// the copy reaches the stored grant.
func (g *Grant) clone() Grant {
	c := *g
	c.Tranches = append([]Tranche(nil), g.Tranches...)
	c.Reports = append([]Report(nil), g.Reports...)
	c.PaidAt = append([]int64(nil), g.PaidAt...)
	c.Holds = append([]Hold(nil), g.Holds...)
	return c
}

// ParseTranches reads "deliverable:amount; deliverable:amount" into
// tranches, in the order written; a trailing ";" is ignored. A deliverable
// may contain colons but not semicolons: the amount is whatever follows the
// last colon. It lets a grantee realm take its request as plain strings.
func ParseTranches(spec string) ([]Tranche, error) {
	var out []Tranche
	parts := strings.Split(spec, ";")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" && i == len(parts)-1 && i > 0 {
			break
		}
		c := strings.LastIndex(part, ":")
		if c < 0 {
			return nil, errors.New("grants: tranche " + strconv.Itoa(i) + " (" + strconv.Quote(part) + ") must read deliverable:amount")
		}
		raw := strings.TrimSpace(part[c+1:])
		amount, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, errors.New("grants: tranche " + strconv.Itoa(i) + ": " + strconv.Quote(raw) + " is not a whole number")
		}
		out = append(out, Tranche{Deliverable: strings.TrimSpace(part[:c]), Amount: amount})
	}
	return out, nil
}

// CheckReview reports whether review fits what Approve keeps: non-blank, at
// most MaxReview bytes and MaxLines lines. A desk can refuse a review when it
// builds an approval rather than when the approval executes.
func CheckReview(review string) error { return checkProse("review", review, MaxReview) }

// checkLink accepts an empty link, a gno.land path ("/r/...") or an https
// URL, of at most MaxLink bytes and only characters that cannot end a
// markdown link destination early. "$" is refused: on gnoweb it opens a
// realm's help or source view, where a link can prefill a transaction.
func checkLink(link string) error {
	if link == "" {
		return nil
	}
	bad := errors.New("grants: link must be a gno.land path (/r/...) or an https:// URL of at most " +
		strconv.Itoa(MaxLink) + " bytes, without spaces, quotes, brackets, parentheses or $")
	if len(link) > MaxLink || !(strings.HasPrefix(link, "https://") || strings.HasPrefix(link, "/") && !strings.HasPrefix(link, "//")) {
		return bad
	}
	for i := 0; i < len(link); i++ {
		c := link[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.IndexByte("-._~:/?#@!&*+,;=%", c) >= 0) {
			return bad
		}
	}
	return nil
}

// checkReason accepts a grantor's reason, never the one a grantee's own
// withdrawal carries, so a cancellation cannot pass for one.
func checkReason(reason string) error {
	if reason == WithdrawnReason {
		return errors.New("grants: that reason is reserved for a grantee's own withdrawal")
	}
	return checkText("reason", reason, MaxReason)
}

// checkLine accepts a non-blank single line of at most max bytes.
func checkLine(field, s string, max int) error {
	if strings.ContainsAny(s, "\r\n") {
		return errors.New("grants: " + field + " must be a single line")
	}
	return checkText(field, s, max)
}

// checkProse accepts non-blank text of at most max bytes and MaxLines lines.
func checkProse(field, s string, max int) error {
	if err := checkText(field, s, max); err != nil {
		return err
	}
	if strings.Count(s, "\n") >= MaxLines {
		return errors.New("grants: " + field + " must be at most " + strconv.Itoa(MaxLines) + " lines")
	}
	return nil
}

// checkText accepts text of at most max bytes with something visible in it:
// spaces and format characters (zero-width, bidi controls), which renderers
// strip, do not count.
func checkText(field, s string, max int) error {
	if len(s) > max || strings.TrimFunc(s, invisible) == "" {
		return errors.New("grants: " + field + " must be 1 to " + strconv.Itoa(max) + " bytes")
	}
	return nil
}

func invisible(r rune) bool { return unicode.IsSpace(r) || unicode.Is(unicode.Cf, r) }
