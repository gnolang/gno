package impl

import (
	"strings"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
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
	YesVotes     memberstore.MembersByTier
	NoVotes      memberstore.MembersByTier
	AbstainVotes memberstore.MembersByTier
	AllVotes     memberstore.MembersByTier

	Accepted bool
	Denied   bool

	DeniedReason string

	TiersAllowedToVote []string
}

func getMembers(cur realm) memberstore.MembersByTier {
	return memberstore.Get(0, cur)
}

func newEmptyVoteStore() memberstore.MembersByTier {
	mbt := memberstore.NewMembersByTier()
	mbt.SetTier(memberstore.T1)
	mbt.SetTier(memberstore.T2)
	mbt.SetTier(memberstore.T3)
	return mbt
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

func (ps *proposalStatus) votePowerPercent(_ int, rlm realm, votes memberstore.MembersByTier) float64 {
	members := getMembers(cross(rlm))
	var vp float64
	memberstore.IterateTiers(func(tn string, tier memberstore.Tier) bool {
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
			sb.WriteString(ps.DeniedReason)
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

func writeVotes(_ int, rlm realm, sb *strings.Builder, t memberstore.MembersByTier, title string) {
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

Executor created in: %s
`, p.ExecutorString(), p.ExecutorCreationRealm())
	}

	return out
}
