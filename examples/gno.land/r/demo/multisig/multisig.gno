package multisig

import (
	"strconv"
	"strings"

	"gno.land/p/nt/avl/v0"

	"chain"
	"chain/banker"
	"chain/runtime/unsafe"
)

// ownPkgPath is this realm's own path, used to derive its address
// deterministically. unsafe.CurrentRealm() was tried here originally,
// but it stack-walks to whichever realm is asking -- a realm that
// composes with this one would get its own address back instead of
// the treasury's. chain.PackageAddress on a hardcoded, known-own path
// has no such ambiguity.
const ownPkgPath = "gno.land/r/demo/multisig"

type Proposal struct {
	id          int
	description string
	to          address
	amount      int64
	executed    bool
	approvals   avl.Tree
}

func (p *Proposal) ID() int             { return p.id }
func (p *Proposal) Description() string { return p.description }
func (p *Proposal) To() address         { return p.to }
func (p *Proposal) Amount() int64       { return p.amount }
func (p *Proposal) Executed() bool      { return p.executed }
func (p *Proposal) ApprovalCount() int  { return p.approvals.Size() }

var (
	owners      avl.Tree
	ownerCount  int
	threshold   int
	initialized bool
	deployer    address
	proposals   avl.Tree
	nextID      int
)

func init() {
	deployer = unsafe.OriginCaller()
	nextID = 1
}

// Setup is a one-shot initializer -- init() cannot take parameters, so
// this fills that role instead, guarded by initialized and restricted
// to the deployer. The three owners must be pairwise distinct: without
// that check, Setup(A, A, B, 2) would silently leave two owners behind
// one slot in the tree while ownerCount and the threshold bound still
// assumed three, making some or all thresholds unreachable forever
// (initialized can't be run twice to fix it).
func Setup(cur realm, owner1 address, owner2 address, owner3 address, thresh int) {
	caller := cur.Previous().Address()
	if caller != deployer {
		panic("only deployer can run setup")
	}
	if initialized {
		panic("already initialized")
	}
	if owner1 == owner2 || owner1 == owner3 || owner2 == owner3 {
		panic("owners must be distinct")
	}
	if thresh < 1 || thresh > 3 {
		panic("invalid threshold")
	}
	owners.Set(owner1.String(), true)
	owners.Set(owner2.String(), true)
	owners.Set(owner3.String(), true)
	ownerCount = 3
	threshold = thresh
	initialized = true
}

func assertOwner(caller address) {
	if !initialized {
		panic("multisig not initialized")
	}
	if owners.Get(caller.String()) == nil {
		panic("caller is not an owner")
	}
}

// Propose opens a new payout proposal. amount must be positive --
// without this check a zero or negative amount would only be caught
// by the bank keeper inside Execute, after owners had already spent
// gas approving something that could never settle, with no way to
// remove it from the tree afterward.
func Propose(cur realm, description string, to address, amount int64) int {
	caller := cur.Previous().Address()
	assertOwner(caller)
	if amount <= 0 {
		panic("amount must be positive")
	}

	id := nextID
	nextID++
	p := &Proposal{
		id:          id,
		description: description,
		to:          to,
		amount:      amount,
	}
	proposals.Set(proposalKey(id), p)
	chain.Emit("ProposalCreated", "id", strconv.Itoa(id), "to", to.String(), "amount", strconv.FormatInt(amount, 10))
	return id
}

func Approve(cur realm, proposalID int) {
	caller := cur.Previous().Address()
	assertOwner(caller)

	raw := proposals.Get(proposalKey(proposalID))
	if raw == nil {
		panic("proposal does not exist")
	}
	p := raw.(*Proposal)
	if p.executed {
		panic("proposal already executed")
	}
	if p.approvals.Get(caller.String()) != nil {
		panic("already approved")
	}
	p.approvals.Set(caller.String(), true)
}

// Revoke lets an owner withdraw their own approval from a
// not-yet-executed proposal. Combined with Cancel, this is how an
// owner set can back out of a proposal that turns out to be a mistake
// -- without it, an approval given while the treasury was empty (or
// before an owner changed their mind) sits there and fires the moment
// someone funds the address, with no way to stop it.
func Revoke(cur realm, proposalID int) {
	caller := cur.Previous().Address()
	assertOwner(caller)

	raw := proposals.Get(proposalKey(proposalID))
	if raw == nil {
		panic("proposal does not exist")
	}
	p := raw.(*Proposal)
	if p.executed {
		panic("proposal already executed")
	}
	_, removed := p.approvals.Remove(caller.String())
	if !removed {
		panic("caller has not approved this proposal")
	}
}

// Cancel removes an unexecuted proposal entirely. Any owner can cancel
// -- the owner set collectively controls the treasury, so any one of
// them should be able to pull a proposal that should never execute,
// the same way any of them can approve one.
func Cancel(cur realm, proposalID int) {
	caller := cur.Previous().Address()
	assertOwner(caller)

	raw := proposals.Get(proposalKey(proposalID))
	if raw == nil {
		panic("proposal does not exist")
	}
	p := raw.(*Proposal)
	if p.executed {
		panic("proposal already executed")
	}
	proposals.Remove(proposalKey(proposalID))
}

func Execute(cur realm, proposalID int) {
	raw := proposals.Get(proposalKey(proposalID))
	if raw == nil {
		panic("proposal does not exist")
	}
	p := raw.(*Proposal)
	if p.executed {
		panic("proposal already executed")
	}
	if p.approvals.Size() < threshold {
		panic("not enough approvals")
	}
	p.executed = true

	bkr := banker.NewBanker(banker.BankerTypeRealmSend, cur)
	bkr.SendCoins(cur.Address(), p.to, chain.Coins{{Denom: "ugnot", Amount: p.amount}})
	chain.Emit("ProposalExecuted", "id", strconv.Itoa(proposalID), "to", p.to.String(), "amount", strconv.FormatInt(p.amount, 10))
}

func TreasuryAddress() address {
	return chain.PackageAddress(ownPkgPath)
}

// proposalKey zero-pads the id so proposals.Iterate (which walks keys
// as strings) lists proposals in numeric order instead of lexical
// order (which would otherwise put proposal 10 before proposal 2).
func proposalKey(id int) string {
	s := strconv.Itoa(id)
	for len(s) < 10 {
		s = "0" + s
	}
	return s
}

// Render shows every proposal, or a single proposal's detail when
// called with ?id=<proposalID>.
func Render(path string) string {
	if id, ok := parseIDQuery(path); ok {
		return renderProposal(id)
	}

	var sb strings.Builder
	sb.WriteString("# Multisig Treasury\n\n")
	sb.WriteString("Initialized: " + strconv.FormatBool(initialized) + "\n")
	if initialized {
		sb.WriteString("Owners: " + strconv.Itoa(ownerCount) + " | Threshold: " + strconv.Itoa(threshold) + "\n")
	}
	sb.WriteString("\n")
	proposals.Iterate("", "", func(key string, value interface{}) bool {
		p := value.(*Proposal)
		sb.WriteString(proposalSummaryLine(p))
		return false
	})
	return sb.String()
}

func renderProposal(id int) string {
	raw := proposals.Get(proposalKey(id))
	if raw == nil {
		return "Proposal not found.\n"
	}
	p := raw.(*Proposal)

	var sb strings.Builder
	sb.WriteString("# Proposal #" + strconv.Itoa(p.id) + "\n\n")
	sb.WriteString(p.description + "\n\n")
	sb.WriteString(proposalSummaryLine(p))
	return sb.String()
}

func proposalSummaryLine(p *Proposal) string {
	status := "pending"
	if p.executed {
		status = "executed"
	}
	return "Proposal #" + strconv.Itoa(p.id) + ": " + p.description +
		" | to: " + p.to.String() +
		" | amount: " + strconv.FormatInt(p.amount, 10) +
		" | approvals: " + strconv.Itoa(p.approvals.Size()) + "/" + strconv.Itoa(threshold) +
		" [" + status + "]\n"
}

func parseIDQuery(path string) (int, bool) {
	const prefix = "?id="
	idx := strings.Index(path, prefix)
	if idx == -1 {
		return 0, false
	}
	rest := path[idx+len(prefix):]
	if amp := strings.Index(rest, "&"); amp != -1 {
		rest = rest[:amp]
	}
	id, err := strconv.Atoi(rest)
	if err != nil {
		return 0, false
	}
	return id, true
}
