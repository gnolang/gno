// Package grants is GovDAO's grant desk for realms.
//
// A grantee is code, never a person: an account in the shared wallet at
// gno.land/r/gov/grants/wallet/v0, which any team or project opens for its
// namespace, or a realm of its own (a project, a team, a DAO hosted
// by a realm) that calls the desk from its code. A GovDAO member sponsors the
// request, with a review statement, as one proposal. If it passes, the
// grant's total is reserved in the pool, and each payment then becomes
// claimable on its own: after the grantee reports on what it promised, a
// review window passes, and a minimum period has gone by since the previous
// payment. A small first payment, an advance, is the only one paid without a
// report. GovDAO can cancel a grant or freeze a grantee at any time, and the
// member who files a cancellation can hold the next payment while GovDAO
// votes; the grantee can withdraw. Grants nobody attends to expire. GovDAO
// can close the desk to new requests in one vote. The rules and their
// durations are on the desk's page.
//
// The pool is this realm's own address. Whoever holds the funds GovDAO
// assigns to grants sends ugnot to it: a GovDAO treasury payment proposal,
// or the holder of the Ecosystem Treasury account. Coins leave it by exactly
// these paths, and nothing else in this realm moves coins:
//
//   - a due payment, to the address of the grantee, which is derived from the
//     caller's identity when it applies and never taken as input;
//   - NewReturnRequest and NewCloseRequest, GovDAO proposals sending
//     uncommitted funds to the address the vote names, normally where they
//     came from.
//
// A grant is only approved if the pool can pay all of it on top of what is
// already promised, so an approved grant is always fully funded.
//
// What the grantee does with the coins is decided by its code. Every wallet
// account runs the same code, to review once, and journals every spend. A realm
// of its own needs that review before the vote: every grant page carries
// the checklist.
package grants

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"errors"
	"strconv"
	"strings"
	"time"

	"gno.land/p/nt/avl/v0"
	gp "gno.land/p/nt/grants/v0"
	"gno.land/p/nt/markdown/sanitize/v0"

	"gno.land/r/gov/dao"
)

const (
	// Denom is the only coin the desk holds or pays.
	Denom = "ugnot"

	path       = "gno.land/r/gov/grants/v0"
	walletPath = "gno.land/r/gov/grants/wallet/v0"
	day        = 24 * time.Hour
)

// Tranche is one payment of a grant; see gno.land/p/nt/grants/v0.
type Tranche = gp.Tranche

var (
	desk   *gp.Program
	frozen = avl.NewTree() // grantee -> why GovDAO froze it
	// built is the cancellation built last, so Hold can tell it was built in
	// the caller's own MsgRun.
	built cancelRef
)

type cancelRef struct {
	id     int
	by     address
	height int64
}

func init() {
	var err error
	desk, err = gp.NewProgram(gp.Rules{
		Window:    14 * day,
		MinPeriod: 30 * day,
		Expiry:    60 * day,
		HoldFor:   14 * day,
		Advance:   20,
	})
	check(err)
}

// Address is the pool: send ugnot here to fund grants.
func Address() address { return chain.PackageAddress(path) }

// Balance is what the pool holds.
func Balance() int64 { return banker.NewReadonlyBanker().GetCoin(Address(), Denom) }

// Closed reports whether GovDAO closed the desk to new requests and approvals.
func Closed() bool { return desk.Closed() }

// Committed is what approved grants can still claim.
func Committed() int64 { return desk.Committed() }

// Available is what the pool can still promise.
func Available() int64 { return Balance() - desk.Committed() }

// Rules are the desk's terms: its review window, minimum period, expiry,
// hold and advance share.
func Rules() gp.Rules { return desk.Rules() }

// GrantsOf lists the latest 20 grants of exactly the grantee at path, so a
// grantee can show them on its own page. The grants are copies.
func GrantsOf(path string) []gp.Grant {
	return desk.Of(path, 20)
}

// Apply files a grant request for the calling realm and returns its id. It
// must be called from a realm's code: users and MsgRun scripts are refused,
// because the grantee is the realm and the money goes to its address.
// Amounts are in ugnot. period is the minimum time between two payments,
// ignored for a single payment. A request the realm still has pending is
// superseded.
func Apply(cur realm, title, pitch string, tranches []Tranche, period time.Duration) int {
	grantee, err := gp.GranteeOf(0, cur)
	check(err)
	if hostFrozen(grantee) {
		panic("grants: GovDAO froze " + grantee + " or its host; it cannot apply until it is unfrozen")
	}
	prev, superseding := desk.PendingOf(grantee)
	id, err := desk.Apply(grantee, title, pitch, tranches, period, time.Now())
	check(err)
	if superseding {
		chain.Emit("GrantSuperseded", "id", itoa(prev), "by", itoa(id), "grantee", grantee)
	}
	chain.Emit("GrantApplied", "id", itoa(id), "grantee", grantee)
	return id
}

// Report posts the grantee's statement on its next unpaid payment: what
// that tranche's deliverable promised, with numbers, and link, which may be
// empty, to what shows it (a gno.land path such as "/r/team/game:v3", or an
// https URL). Posting again adds a report and restarts the review window.
func Report(cur realm, id int, body, link string) {
	grantee, err := gp.GranteeOf(0, cur)
	check(err)
	check(desk.Report(grantee, id, body, link, time.Now()))
	chain.Emit("GrantReported", "id", itoa(id), "grantee", grantee)
}

// Withdraw gives up grant id, pending or active, for the calling grantee.
// What it was not paid goes back to the pool.
func Withdraw(cur realm, id int) {
	grantee, err := gp.GranteeOf(0, cur)
	check(err)
	check(desk.Withdraw(grantee, id, time.Now()))
	chain.Emit("GrantWithdrawn", "id", itoa(id), "grantee", grantee)
}

// Claim pays the next payment of grant id to its grantee once it is due and
// the grantee is not frozen. Anyone may call it; the coins can only go to the
// grantee.
func Claim(cur realm, id int) {
	if g := mustGet(id); Frozen(g.Grantee) {
		panic("grants: GovDAO froze " + g.Grantee + "; its payments wait until it is unfrozen")
	}
	grantee, amount, err := desk.Claim(id, time.Now())
	if err == gp.ErrNotDue {
		at, _ := desk.PayableAt(id)
		panic("grants: next payment of grant #" + itoa(id) + " is payable at " + when(at))
	}
	check(err)
	send(cur, chain.PackageAddress(grantee), amount)
	chain.Emit("GrantPaid", "id", itoa(id), "grantee", grantee, "amount", i64(amount))
}

// Frozen reports whether GovDAO froze grantee: the desk pays it nothing, it
// cannot apply, and a wallet account cannot spend or change hands, until
// GovDAO unfreezes it.
func Frozen(grantee string) bool { return frozen.Has(grantee) }

// FrozenReason is the reason GovDAO gave for freezing grantee, or "".
func FrozenReason(grantee string) string {
	why, _ := frozen.Get(grantee).(string)
	return why
}

// Expire closes a grant left unattended: a request nobody approved in time,
// or a grant whose next payment was never reported. Anyone may call it; an
// expired grant gives what it had not been paid back to the pool.
func Expire(cur realm, id int) {
	check(desk.Expire(id, time.Now()))
	g, _ := desk.Get(id)
	chain.Emit("GrantExpired", "id", itoa(id), "grantee", g.Grantee)
}

// Hold pushes the next payment of grant id back while GovDAO votes on
// cancelling it, once per payment. Only the author of the newest GovDAO
// proposal can place it, only if that proposal is the cancellation this desk
// builds for that grant, and only in the block the author built it in, so it
// is the last line of the MsgRun that files the cancellation, before any vote:
//
//	pid := dao.MustCreateProposal(cross(cur), grants.NewCancelRequest(cross(cur), id, reason))
//	grants.Hold(cross(cur), id, int64(pid))
//
// Authorship proves membership: GovDAO only accepts proposals from members.
// The proposal is matched by its stored title and its executor's creation
// realm, never by calling the executor. Builders of other realms that take
// any title (r/sys/params has some) fail the second check; what is left is a
// member deploying an executor type of their own, so what bounds the brake is
// one hold per payment. GovDAO exposes no proposal status, so a later hold
// could not tell an open vote from a lost one.
func Hold(cur realm, id int, pid int64) {
	g := mustGet(id)
	p, err := dao.GetProposal(dao.ProposalID(pid))
	check(err)
	if _, err := dao.GetProposal(dao.ProposalID(pid + 1)); err == nil {
		panic("grants: Hold must follow the cancellation in the same MsgRun; proposal " + i64(pid) + " is not the newest")
	}
	if p.Author() != cur.Previous().Address() {
		panic("grants: only the author of the cancellation can hold")
	}
	if p.Title() != cancelTitle(g) || p.ExecutorCreationRealm() != path {
		panic("grants: proposal " + i64(pid) + " is not the cancellation of grant #" + itoa(id))
	}
	if built != (cancelRef{id, cur.Previous().Address(), runtime.ChainHeight()}) {
		panic("grants: Hold must follow the cancellation in the same MsgRun")
	}
	check(desk.Hold(id, pid, time.Now()))
	chain.Emit("GrantHeld", "id", itoa(id), "grantee", g.Grantee, "proposal", i64(pid))
}

// NewApproveRequest builds the GovDAO proposal that approves grant id and
// reserves its total in the pool. review is the sponsor's statement, shown to
// voters and kept on the grant: who controls the grantee and how they are
// known, what was checked in its code, any conflict of interest. The sponsor
// is the address that builds the request, in its own MsgRun; GovDAO records
// the member who files it as the proposal's author. If the pool is short when the
// proposal is executed, execution fails and changes nothing, so it can be
// executed again once the pool is funded. A request that expired or is no
// longer pending makes execution fail with an error, which
// ExecuteOrRejectProposal turns into a rejection.
func NewApproveRequest(cur realm, id int, review string) dao.ProposalRequest {
	if desk.Closed() {
		check(gp.ErrClosed)
	}
	if gp.CheckReview(review) != nil || strings.HasPrefix(strings.TrimSpace(review), "REVIEW:") {
		panic("grants: write your own review, 1 to " + itoa(gp.MaxReview) + " bytes and at most " + itoa(gp.MaxLines) +
			" lines: who controls the grantee, what you checked, any conflict of interest")
	}
	if !cur.Previous().IsUser() {
		panic("grants: build the approval in your own MsgRun, so the review is signed by you")
	}
	sponsor := cur.Previous().Address().String()
	g := mustGet(id)
	if hostFrozen(g.Grantee) {
		panic("grants: GovDAO froze " + g.Grantee + " or its host; unfreeze it before approving its grants")
	}
	if g.Status != gp.Pending {
		panic("grants: grant #" + itoa(id) + " is not pending")
	}
	if !time.Now().Before(desk.ExpiresAt(g)) {
		panic("grants: grant #" + itoa(id) + " has expired; the grantee must apply again")
	}
	grantee := g.Grantee
	cb := func(cur realm) error {
		if hostFrozen(grantee) {
			return errors.New("grants: GovDAO froze " + grantee + " or its host after this was proposed")
		}
		err := desk.Approve(id, Balance(), sponsor, review, time.Now())
		if err == gp.ErrUnderfunded {
			// A panic cannot be turned into a rejection, so the passed
			// proposal stays executable once the pool is funded. Returning
			// the error would let anyone kill it with ExecuteOrRejectProposal.
			panic(err.Error() + "; fund " + Address().String() + " and execute again")
		}
		if err != nil {
			return err
		}
		chain.Emit("GrantApproved", "id", itoa(id), "grantee", grantee, "sponsor", sponsor)
		return nil
	}
	return dao.NewProposalRequest(
		approveTitle(g),
		approveText(g, sponsor, review),
		dao.NewSimpleExecutor(0, cur, cb, "Approves grant #"+itoa(id)+" of "+path),
	)
}

// NewCancelRequest builds the GovDAO proposal that cancels grant id. Paid
// tranches stay paid; the rest goes back to the pool's free balance. Its
// author can hold the next payment while it is voted: see Hold. Voters read
// what the grant was when proposed: if it is approved, or stops, before the
// vote executes, execution fails.
func NewCancelRequest(cur realm, id int, reason string) dao.ProposalRequest {
	g := mustGet(id)
	if g.Status != gp.Pending && g.Status != gp.Active {
		panic("grants: grant #" + itoa(id) + " is " + g.Status.String())
	}
	mustReason(reason)
	built = cancelRef{id, cur.Previous().Address(), runtime.ChainHeight()}
	grantee, was := g.Grantee, g.Status
	cb := func(cur realm) error {
		if now := mustGet(id); now.Status != was {
			return errors.New("grants: grant #" + itoa(id) + " was " + was.String() + " when this was proposed and is " +
				now.Status.String() + " now; file a new cancellation")
		}
		if err := desk.Cancel(id, reason, time.Now()); err != nil {
			return err
		}
		chain.Emit("GrantCancelled", "id", itoa(id), "grantee", grantee)
		return nil
	}
	return dao.NewProposalRequest(
		cancelTitle(g),
		cancelText(g, reason),
		dao.NewSimpleExecutor(0, cur, cb, "Cancels grant #"+itoa(id)+" of "+path),
	)
}

// NewReturnRequest builds the GovDAO proposal that sends amount ugnot of
// the pool's uncommitted funds to to, normally the account they came from.
func NewReturnRequest(cur realm, to address, amount int64, reason string) dao.ProposalRequest {
	mustDestination(to)
	if amount <= 0 || amount > Available() {
		panic("grants: amount must be between 1 and the pool's available " + i64(Available()) + Denom)
	}
	mustReason(reason)
	cb := func(cur realm) error {
		if amount > Available() {
			return errors.New("grants: the pool has only " + i64(Available()) + Denom + " uncommitted")
		}
		pay(cur, to, amount)
		return nil
	}
	return dao.NewProposalRequest(
		"Return "+gnot(amount)+" from the grants pool to "+to.String(),
		"If this passes, "+gnot(amount)+" of the grants pool's uncommitted funds go to "+sanitize.InlineCode(to.String())+". "+
			"Approved grants are not affected.\n\nReason: "+escape(reason)+"\n",
		dao.NewSimpleExecutor(0, cur, cb, "Returns "+i64(amount)+Denom+" from "+path+" to "+to.String()),
	)
}

// NewFreezeRequest builds the GovDAO proposal that freezes grantee, or
// unfreezes it when freeze is false. While frozen, the desk pays it nothing
// and a wallet account cannot spend; its grants are not cancelled, and a
// payment that falls due waits.
func NewFreezeRequest(cur realm, grantee string, freeze bool, reason string) dao.ProposalRequest {
	// ByGrantee lists the path itself first, then the identities it hosts.
	if l, _ := desk.ByGrantee(grantee, 0, 1); len(l) == 0 || l[0].Grantee != grantee && gp.Host(l[0].Grantee) != grantee {
		panic("grants: neither " + grantee + " nor an identity it hosts has ever applied")
	}
	if freeze == Frozen(grantee) {
		panic("grants: " + grantee + " is already " + frozenWord(freeze))
	}
	mustReason(reason)
	cb := func(cur realm) error {
		if freeze == Frozen(grantee) {
			return errors.New("grants: " + grantee + " is already " + frozenWord(freeze))
		}
		if freeze {
			frozen.Set(grantee, reason)
		} else {
			frozen.Remove(grantee)
		}
		chain.Emit("GranteeFrozen", "grantee", grantee, "frozen", strconv.FormatBool(freeze))
		return nil
	}
	effect := "the desk pays it nothing, it cannot apply and, if it is a wallet account, it cannot spend or change hands, " +
		"until GovDAO unfreezes it. Its grants are not cancelled: a payment that falls due waits."
	if !strings.Contains(grantee, "#") {
		effect += " Identities it hosts (" + sanitize.InlineCode(grantee+"#...") + ") cannot apply or be approved either, " +
			"but those already approved stay payable: freeze each one to stop its payments."
	}
	verb := "Freeze "
	if !freeze {
		verb = "Unfreeze "
		effect = "its due payments can be released again and, if it is a wallet account, it can spend again."
	}
	return dao.NewProposalRequest(
		verb+short(grantee),
		"If this passes, "+sanitize.InlineCode(grantee)+" is "+frozenWord(freeze)+": "+effect+
			" Other grantees of its namespace, such as other wallet accounts its owner opens, are grantees of their own."+
			"\n\nGrants in its namespace: [the list]("+link("ns/"+namespace(grantee))+").\n\nReason: "+escape(reason)+"\n",
		dao.NewSimpleExecutor(0, cur, cb, verb+grantee+" at "+path),
	)
}

// NewCloseRequest builds the GovDAO proposal that closes the desk to new
// requests and approvals, to stop a round or move to a new desk. Running
// grants go on, paid from their reservations, and the pool's uncommitted
// funds go to to, normally where they came from. With cancelGrants, every
// active grant is cancelled as well (paid tranches stay paid) and the whole
// pool goes; a desk closed without it can be closed again with it.
func NewCloseRequest(cur realm, to address, reason string, cancelGrants bool) dao.ProposalRequest {
	if desk.Closed() && (!cancelGrants || desk.ActiveCount() == 0) {
		check(gp.ErrClosed)
	}
	mustDestination(to)
	mustReason(reason)
	cb := func(cur realm) error {
		var cancelled []gp.Grant
		if cancelGrants {
			cancelled = desk.Active(0, desk.ActiveCount())
		}
		if err := desk.Close(reason, cancelGrants, time.Now()); err != nil {
			return err
		}
		for _, g := range cancelled {
			chain.Emit("GrantCancelled", "id", itoa(g.ID), "grantee", g.Grantee)
		}
		chain.Emit("DeskClosed", "cancelled", itoa(len(cancelled)))
		if a := Available(); a > 0 {
			pay(cur, to, a)
		}
		return nil
	}
	dest := sanitize.InlineCode(to.String())
	effect := "its active grants (" + itoa(desk.ActiveCount()) + " when proposed) "
	if cancelGrants {
		effect += "are cancelled, paid tranches staying paid, and the whole pool (" + gnot(Balance()) + " when proposed) goes to " + dest + "."
	} else {
		effect += "go on as approved, and the uncommitted funds (" + gnot(Available()) + " when proposed) go to " + dest + "."
	}
	return dao.NewProposalRequest(
		"Close the grants desk, funds to "+to.String(),
		"If this passes, the grants desk takes no new request or approval, pending requests can no longer be approved, "+
			effect+"\n\nReason: "+escape(reason)+"\n",
		dao.NewSimpleExecutor(0, cur, cb, "Closes "+path+", funds to "+to.String()),
	)
}

func pay(cur realm, to address, amount int64) {
	send(cur, to, amount)
	chain.Emit("PoolReturned", "amount", i64(amount), "to", to.String())
}

// hostFrozen reports whether GovDAO froze grantee or the realm whose code
// runs for it: a frozen realm cannot bring new money in under a fresh
// sub-identity. Payments to identities already approved are stopped only by
// freezing each one.
func hostFrozen(grantee string) bool { return Frozen(grantee) || Frozen(gp.Host(grantee)) }

func mustDestination(to address) {
	if !to.IsValid() || to == Address() {
		panic("grants: name a valid address, other than the pool, to send the funds to")
	}
}

func send(cur realm, to address, amount int64) {
	banker.NewBanker(banker.BankerTypeRealmSend, cur).
		SendCoins(cur.Address(), to, chain.NewCoins(chain.NewCoin(Denom, amount)))
}

func cancelTitle(g gp.Grant) string { return "Cancel grant #" + itoa(g.ID) + ": " + g.Title }

func approveTitle(g gp.Grant) string {
	return "Grant #" + itoa(g.ID) + ": " + gnot(g.Total()) + " to " + short(g.Grantee) + ", " + g.Title
}

func frozenWord(freeze bool) string {
	if freeze {
		return "frozen"
	}
	return "unfrozen"
}

func mustGet(id int) gp.Grant {
	g, ok := desk.Get(id)
	if !ok {
		check(gp.ErrNotFound)
	}
	return g
}

func mustReason(reason string) {
	if strings.TrimSpace(reason) == "" || len(reason) > gp.MaxReason || reason == "REASON" {
		panic("grants: write a reason of 1 to " + itoa(gp.MaxReason) + " bytes")
	}
	if reason == gp.WithdrawnReason {
		panic("grants: that reason is reserved for a grantee's own withdrawal")
	}
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func itoa(n int) string  { return strconv.Itoa(n) }
func i64(n int64) string { return strconv.FormatInt(n, 10) }
