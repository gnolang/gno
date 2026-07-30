package commondao

import (
	"chain"
	"chain/banker"
	"errors"
	"strings"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao/v0"
)

// hasLiveProperAncestor checks if any proper ancestor of dao is not
// dissolved.
func hasLiveProperAncestor(dao *commondao.CommonDAO) bool {
	for p := dao.Parent(); p != nil; p = p.Parent() {
		if !p.IsDeleted() {
			return true
		}
	}
	return false
}

// isProperAncestor checks if dao is a proper ancestor of target.
// Parent pointers are set only at construction and no re-parenting path
// exists, so ancestry is stable for the lifetime of a proposal.
func isProperAncestor(dao, target *commondao.CommonDAO) bool {
	for p := target.Parent(); p != nil; p = p.Parent() {
		if p.ID() == dao.ID() {
			return true
		}
	}
	return false
}

// assertIsProperAncestor validates the ancestor relation that authorizes
// treasury controls over a descendant (docs/CONSTITUTION.md :1507). The
// relation is strictly proper: a DAO can never claw back or unfreeze
// itself.
func assertIsProperAncestor(dao, target *commondao.CommonDAO) error {
	if target.ID() == dao.ID() {
		return errors.New("a DAO cannot target itself")
	}
	if !isProperAncestor(dao, target) {
		return errors.New("DAO is not an ancestor of the target DAO")
	}
	return nil
}

// treasuryBalance returns the current balance of a DAO's treasury address.
func treasuryBalance(dao *commondao.CommonDAO) chain.Coins {
	return banker.NewReadonlyBanker().GetCoins(dao.Address())
}

// newTreasurySpendPropDefinition creates a proposal definition that sends
// coins from the DAO's own treasury.
func newTreasurySpendPropDefinition(dao *commondao.CommonDAO, to address, coin chain.Coin) treasurySpendPropDefinition {
	if dao == nil {
		panic("DAO is required")
	}
	if !to.IsValid() {
		panic("invalid recipient address")
	}
	if coin.Denom == "" {
		panic("coin denomination is empty")
	}
	if !coin.IsPositive() {
		panic("spend amount must be positive")
	}

	return treasurySpendPropDefinition{dao, to, coin}
}

// treasurySpendPropDefinition defines a proposal type for spending funds
// from the DAO's own treasury (docs/CONSTITUTION.md :1542-1543).
type treasurySpendPropDefinition struct {
	dao  *commondao.CommonDAO
	to   address
	coin chain.Coin
}

func (treasurySpendPropDefinition) Title() string               { return "Treasury Spend" }
func (treasurySpendPropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

// isTrustedMarkdownBody marks Body as self-assembled markdown; the embedded
// recipient and amount are formatted by md helpers / EscapeText in Body.
func (treasurySpendPropDefinition) isTrustedMarkdownBody() {}

// Threshold returns the tally threshold: the constitution attaches no
// spend-specific rule, so the supermajority default applies.
func (treasurySpendPropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

func (p treasurySpendPropDefinition) Body() string {
	var b strings.Builder

	b.WriteString(md.Paragraph(md.Bold("Recipient:") + "\n" + userLink(p.to)))
	b.WriteString(md.Paragraph(md.Bold("Amount:") + "\n" + md.EscapeText(p.coin.String())))

	return b.String()
}

// Validate runs at proposal creation and again at execution, so a
// treasury frozen or drained after the proposal passed still fails it
// cleanly (StatusFailed, coins untouched) instead of panicking the tx.
func (p treasurySpendPropDefinition) Validate() error {
	if p.dao.IsDeleted() {
		return errors.New("DAO has already been dissolved")
	}
	if p.dao.IsTreasuryFrozen() {
		return errors.New("DAO treasury is frozen")
	}
	if treasuryBalance(p.dao).AmountOf(p.coin.Denom) < p.coin.Amount {
		return errors.New("insufficient treasury balance")
	}
	return nil
}

func (p treasurySpendPropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

// FundingDAOID returns the ID of the DAO whose treasury funds the spend:
// its own (the host).
func (p treasurySpendPropDefinition) FundingDAOID() uint64 {
	return p.dao.ID()
}

func (p treasurySpendPropDefinition) execute(_ int, sub realm) error {
	// sub is this DAO's sub-identity, minted by the host: send from it.
	// Banker sends move bank balances without invoking recipient code, so
	// there is no reentrancy vector. Validate ran in this same Execute
	// call, so the balance check is current.
	b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
	b.SendCoins(sub.Address(), p.to, chain.NewCoins(p.coin))
	return nil
}

// newTreasuryClawbackPropDefinition creates a proposal definition that
// sweeps a descendant DAO's treasury one step up the tree.
func newTreasuryClawbackPropDefinition(dao, target *commondao.CommonDAO) treasuryClawbackPropDefinition {
	if dao == nil {
		panic("DAO is required")
	}
	if target == nil {
		panic("target DAO is required")
	}

	return treasuryClawbackPropDefinition{dao, target}
}

// treasuryClawbackPropDefinition defines a proposal type for an ancestor
// DAO to reclaim a descendant's treasury. The destination is fixed — the
// target's parent — so funds move one step up the tree toward their
// origin and can never be drained out of the tree entirely. Clawback
// remains valid against soft-deleted and frozen descendants.
type treasuryClawbackPropDefinition struct {
	dao    *commondao.CommonDAO // proposing DAO, must be a proper ancestor
	target *commondao.CommonDAO
}

func (treasuryClawbackPropDefinition) Title() string               { return "Treasury Clawback" }
func (treasuryClawbackPropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

// isTrustedMarkdownBody marks Body as self-assembled markdown (a DAO link).
func (treasuryClawbackPropDefinition) isTrustedMarkdownBody() {}

// Threshold returns the tally threshold: simple majority, the
// constitutional wording for this ancestor power.
func (treasuryClawbackPropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSimpleMajority
}

func (p treasuryClawbackPropDefinition) Body() string {
	var b strings.Builder

	b.WriteString(md.Paragraph(md.Bold("Target DAO:") + "\n" + daoMDLink(p.target)))
	b.WriteString(md.Paragraph(
		md.Bold("Destination:") + "\n" +
			"The target's parent DAO receives the target's full balance at execution time.",
	))

	return b.String()
}

func (p treasuryClawbackPropDefinition) Validate() error {
	return assertIsProperAncestor(p.dao, p.target)
}

func (p treasuryClawbackPropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

// FundingDAOID returns the ID of the DAO whose treasury the clawback
// sweeps: the target, not the proposing ancestor that hosts the proposal.
func (p treasuryClawbackPropDefinition) FundingDAOID() uint64 {
	return p.target.ID()
}

func (p treasuryClawbackPropDefinition) execute(_ int, sub realm) error {
	balance := treasuryBalance(p.target)
	if balance.IsZero() {
		return nil
	}

	// sub is the target's sub-identity, minted by the host. A proper
	// ancestor exists, so the target always has a parent. Full-balance
	// send: cost is O(number of denoms held). A target dusted with many
	// realm-minted denoms can push this past block gas (known limitation —
	// see the treasury ADR §Sweep gas-bomb).
	b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
	b.SendCoins(sub.Address(), p.target.Parent().Address(), balance)
	return nil
}

// newTreasuryFreezePropDefinition creates a proposal definition that
// freezes or unfreezes a descendant DAO's treasury.
func newTreasuryFreezePropDefinition(dao, target *commondao.CommonDAO, frozen bool) treasuryFreezePropDefinition {
	if dao == nil {
		panic("DAO is required")
	}
	if target == nil {
		panic("target DAO is required")
	}

	return treasuryFreezePropDefinition{dao, target, frozen}
}

// treasuryFreezePropDefinition defines a proposal type for an ancestor
// DAO to freeze or unfreeze a descendant's treasury. Freezing does not
// cascade: ancestors freeze each descendant explicitly. Only a proper
// ancestor can unfreeze — the frozen DAO's own council cannot.
type treasuryFreezePropDefinition struct {
	dao    *commondao.CommonDAO // proposing DAO, must be a proper ancestor
	target *commondao.CommonDAO
	frozen bool
}

func (p treasuryFreezePropDefinition) Title() string {
	if p.frozen {
		return "Treasury Freeze"
	}
	return "Treasury Unfreeze"
}

func (treasuryFreezePropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

// isTrustedMarkdownBody marks Body as self-assembled markdown (a DAO link).
func (treasuryFreezePropDefinition) isTrustedMarkdownBody() {}

// Threshold returns the tally threshold: simple majority, matching the
// clawback power it safeguards.
func (treasuryFreezePropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSimpleMajority
}

func (p treasuryFreezePropDefinition) Body() string {
	var b strings.Builder

	b.WriteString(md.Paragraph(md.Bold("Target DAO:") + "\n" + daoMDLink(p.target)))

	action := "frozen: no treasury spend can execute until a proper ancestor unfreezes it."
	if !p.frozen {
		action = "unfrozen: treasury spends can execute again."
	}
	b.WriteString(md.Paragraph(md.Bold("Effect:") + "\nThe target's treasury is " + action))

	return b.String()
}

func (p treasuryFreezePropDefinition) Validate() error {
	// Orphan rescue: when every proper ancestor is dissolved, the freezing
	// authority class is extinct, so the target's own council may restore
	// the constitutional default by unfreezing itself. Without this, a
	// frozen DAO orphaned by its ancestors' dissolution would hold its
	// funds locked forever. Freezing - and any self-targeting while a
	// live ancestor exists - stays ancestor-only.
	if !p.frozen && p.dao.ID() == p.target.ID() && !hasLiveProperAncestor(p.target) {
		return nil
	}
	return assertIsProperAncestor(p.dao, p.target)
}

func (p treasuryFreezePropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

func (p treasuryFreezePropDefinition) execute(_ int, sub realm) error {
	p.target.SetTreasuryFrozen(p.frozen)
	return nil
}

// daoMDLink returns a markdown link to a DAO's page.
func daoMDLink(dao *commondao.CommonDAO) string {
	return md.Link(dao.Name(), daoURL(dao.ID()))
}
