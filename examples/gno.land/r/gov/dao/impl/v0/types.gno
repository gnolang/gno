package impl

import (
	"strings"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/markdown/sanitize/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/memberstore/v0"
)

type Law struct {
	Supermajority float64
}

func NewLaw(supermajority float64) Law {
	return Law{Supermajority: supermajority}
}

func (l *Law) String() string {
	return ufmt.Sprintf("This law contains the following data:\n\n- Supermajority: %v%%", l.Supermajority)
}

// ProposalsStatuses contains the status of all the proposals indexed by the proposal ID.
type ProposalsStatuses struct {
	*bptree.BPTree // map[int]*proposalStatus
}

func NewProposalsStatuses() ProposalsStatuses {
	return ProposalsStatuses{bptree.NewBPTree32()}
}

func (pss ProposalsStatuses) GetStatus(id dao.ProposalID) *proposalStatus {
	if pss.BPTree == nil {
		return nil
	}

	pids := id.String()
	psv := pss.Get(pids)
	if psv == nil {
		return nil
	}

	ps, ok := psv.(*proposalStatus)
	if !ok {
		panic("ProposalsStatuses must contains only proposalStatus types")
	}

	return ps
}

type proposalStatus struct {
	YesVotes     voteStore
	NoVotes      voteStore
	AbstainVotes voteStore
	AllVotes     voteStore

	Accepted bool
	Denied   bool

	DeniedReason string

	TiersAllowedToVote []string
}

func getMembers(cur realm) *memberstore.MembersByTier {
	return memberstore.Get(0, cur)
}

// voteStore records a proposal's votes: tier name -> address -> member. It is
// this realm's own type, so this realm's code writes it; memberstore's
// MembersByTier belongs to memberstore and changes only by crossing into it.
type voteStore struct {
	*bptree.BPTree
}

func newEmptyVoteStore() voteStore {
	vs := voteStore{BPTree: bptree.NewBPTree32()}
	for _, tn := range []string{memberstore.T1, memberstore.T2, memberstore.T3} {
		vs.Set(tn, bptree.NewBPTree32())
	}
	return vs
}

// SetMember records addr's vote under tier.
func (vs voteStore) SetMember(tier string, addr address, m *memberstore.Member) {
	if t, ok := vs.Get(tier).(*bptree.BPTree); ok {
		t.Set(string(addr), m)
	}
}

// GetMember returns addr's recorded member and tier, if any.
func (vs voteStore) GetMember(addr address) (m *memberstore.Member, tier string) {
	vs.Iterate("", "", func(tn string, v any) bool {
		mv := v.(*bptree.BPTree).Get(string(addr))
		if mv == nil {
			return false
		}
		m, tier = mv.(*memberstore.Member), tn
		return true
	})
	return
}

// GetTierSize returns how many votes tier holds.
func (vs voteStore) GetTierSize(tn string) int {
	t, ok := vs.Get(tn).(*bptree.BPTree)
	if !ok {
		return 0
	}
	return t.Size()
}

func newProposalStatus(allowedToVote []string) *proposalStatus {
	return &proposalStatus{
		YesVotes:           newEmptyVoteStore(),
		NoVotes:            newEmptyVoteStore(),
		AbstainVotes:       newEmptyVoteStore(),
		AllVotes:           newEmptyVoteStore(),
		TiersAllowedToVote: allowedToVote,
	}
}

// totalPower computes the total voting power dynamically from current members
// rather than using a snapshot. See https://github.com/gnolang/gno/pull/5271#discussion_r2952523023
//
// Non-crossing: rlm is threaded into the crossing getMembers(cross(rlm))
// call. The proposalStatus methods that compose this (votePowerPercent,
// YesPercent etc., String) all take `_ int, rlm realm` for the same reason.
func (ps *proposalStatus) totalPower(_ int, rlm realm) float64 {
	members := getMembers(cross(rlm))
	var tp float64
	for _, tn := range ps.TiersAllowedToVote {
		power := memberstore.GetTierPower(tn, members)
		tp += power * float64(members.GetTierSize(tn))
	}
	return tp
}

func (ps *proposalStatus) votePowerPercent(_ int, rlm realm, votes voteStore) float64 {
	members := getMembers(cross(rlm))
	var vp float64
	memberstore.IterateTiers(func(tn string, tier *memberstore.Tier) bool {
		power := memberstore.GetTierPower(tn, members)
		ts := votes.GetTierSize(tn)
		vp = vp + (power * float64(ts))
		return false
	})
	tp := ps.totalPower(0, rlm)
	if tp == 0 {
		return 0
	}
	return (vp / tp) * 100
}

func (ps *proposalStatus) YesPercent(_ int, rlm realm) float64 {
	return ps.votePowerPercent(0, rlm, ps.YesVotes)
}

func (ps *proposalStatus) NoPercent(_ int, rlm realm) float64 {
	return ps.votePowerPercent(0, rlm, ps.NoVotes)
}

func (ps *proposalStatus) AbstainPercent(_ int, rlm realm) float64 {
	return ps.votePowerPercent(0, rlm, ps.AbstainVotes)
}

func (ps *proposalStatus) IsAllowed(tier string) bool {
	for _, ta := range ps.TiersAllowedToVote {
		if ta == tier {
			return true
		}
	}

	return false
}

// XXX: can be optimized by passing down total power to Yes/No/Abstain percent fn to avoid re-computing
func (ps *proposalStatus) String(_ int, rlm realm) string {
	var sb strings.Builder
	sb.WriteString("### Stats\n")

	if ps.Accepted {
		sb.WriteString("- **PROPOSAL HAS BEEN ACCEPTED**\n")
	} else if ps.Denied {
		sb.WriteString("- **PROPOSAL HAS BEEN DENIED**\n")
		if ps.DeniedReason != "" {
			sb.WriteString("REASON: ")
			// DeniedReason is "execution failed: " + err.Error() from the
			// proposal's executor callback, which any third-party realm can
			// supply. Raw, it wrote attacker-controlled markdown directly beneath
			// the "PROPOSAL HAS BEEN DENIED" line. InlineText folds newlines, so
			// the reason cannot leave the REASON: line, and escapes emphasis and
			// raw HTML.
			sb.WriteString(sanitize.InlineText(clampField(ps.DeniedReason, maxRenderedReason)))
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("- **Proposal is open for votes**\n")
	}

	sb.WriteString("- Tiers eligible to vote: ")
	sb.WriteString(strings.Join(ps.TiersAllowedToVote, ", "))
	sb.WriteString("\n")

	sb.WriteString(ufmt.Sprintf("- YES PERCENT: %v%%\n", ps.YesPercent(0, rlm)))
	sb.WriteString(ufmt.Sprintf("- NO PERCENT: %v%%\n", ps.NoPercent(0, rlm)))
	sb.WriteString(ufmt.Sprintf("- ABSTAIN PERCENT: %v%%\n", ps.AbstainPercent(0, rlm)))

	return sb.String()
}

func StringifyVotes(_ int, rlm realm, ps *proposalStatus) string {
	var sb strings.Builder

	writeVotes(0, rlm, &sb, ps.YesVotes, "YES")
	writeVotes(0, rlm, &sb, ps.NoVotes, "NO")
	writeVotes(0, rlm, &sb, ps.AbstainVotes, "ABSTAIN")

	if sb.String() == "" {
		return "No one voted yet."
	}

	return sb.String()
}

func writeVotes(_ int, rlm realm, sb *strings.Builder, t voteStore, title string) {
	if t.Size() == 0 {
		return
	}
	members := getMembers(cross(rlm))
	t.Iterate("", "", func(tn string, value interface{}) bool {
		_, ok := memberstore.GetTier(tn)
		if !ok {
			panic("tier not found")
		}

		power := memberstore.GetTierPower(tn, members)

		sb.WriteString(ufmt.Sprintf("%v from %v (VPPM %v):\n\n", title, tn, power))
		ms, _ := value.(*bptree.BPTree)
		ms.Iterate("", "", func(addr string, _ interface{}) bool {
			sb.WriteString("- " + tryResolveAddr(address(addr)) + "\n")
			return false
		})

		sb.WriteString("\n")

		return false
	})
}

func StringifyProposal(p *dao.Proposal) string {
	out := ufmt.Sprintf(`
### Title: %s

### Proposed by: %s

%s
`, p.Title(), p.Author(), p.Description())

	if p.ExecutorString() != "" {
		out += ufmt.Sprintf(`
This proposal contains the following metadata:

%s
`, p.ExecutorString())
	}

	// Disclosed independently of the metadata string, and escaped, for the
	// same reasons as render.renderProposalPage.
	if cr := sanitize.InlineCode(strings.TrimSpace(clampField(p.ExecutorCreationRealm(), maxRenderedRealm))); cr != "" {
		out += ufmt.Sprintf("\nExecutor created in: %s\n", cr)
	}

	return out
}
