package impl

import (
	"strings"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)

type Law struct {
	Supermajority float64
}

func (l *Law) String() string {
	return ufmt.Sprintf("This law contains the following data:\n\n- Supermajority: %v%%", l.Supermajority)
}

// ProposalsStatuses contains the status of all the proposals indexed by the proposal ID.
type ProposalsStatuses struct {
	*avl.Tree // map[int]*proposalStatus
}

func NewProposalsStatuses() ProposalsStatuses {
	return ProposalsStatuses{avl.NewTree()}
}

func (pss ProposalsStatuses) GetStatus(id dao.ProposalID) *proposalStatus {
	if pss.Tree == nil {
		return nil
	}

	pids := id.String()
	psv, ok := pss.Get(pids)
	if !ok {
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

	TotalPower float64 // TotalPower is the power of all the members existing when this proposal was created.
}

func getMembers(cur realm) memberstore.MembersByTier {
	return memberstore.Get()
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
		TotalPower:         getMembers(cross).GetTotalPower(),
	}
}

func (ps *proposalStatus) votePowerPercent(votes memberstore.MembersByTier) float64 {
	var vp float64
	memberstore.IterateTiers(func(tn string, tier memberstore.Tier) bool {
		power := memberstore.GetTierPower(tn, getMembers(cross))
		ts := votes.GetTierSize(tn)
		vp = vp + (power * float64(ts))
		return false
	})
	return (vp / ps.TotalPower) * 100
}

func (ps *proposalStatus) YesPercent() float64     { return ps.votePowerPercent(ps.YesVotes) }
func (ps *proposalStatus) NoPercent() float64      { return ps.votePowerPercent(ps.NoVotes) }
func (ps *proposalStatus) AbstainPercent() float64 { return ps.votePowerPercent(ps.AbstainVotes) }

func (ps *proposalStatus) IsAllowed(tier string) bool {
	for _, ta := range ps.TiersAllowedToVote {
		if ta == tier {
			return true
		}
	}

	return false
}

func (ps *proposalStatus) String() string {
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

	sb.WriteString(ufmt.Sprintf("- YES PERCENT: %v%%\n", ps.YesPercent()))
	sb.WriteString(ufmt.Sprintf("- NO PERCENT: %v%%\n", ps.NoPercent()))
	sb.WriteString(ufmt.Sprintf("- ABSTAIN PERCENT: %v%%\n", ps.AbstainPercent()))

	return sb.String()
}

func StringifyVotes(ps *proposalStatus) string {
	var sb strings.Builder

	writeVotes(&sb, ps.YesVotes, "YES")
	writeVotes(&sb, ps.NoVotes, "NO")
	writeVotes(&sb, ps.AbstainVotes, "ABSTAIN")

	if sb.String() == "" {
		return "No one voted yet."
	}

	return sb.String()
}

func writeVotes(sb *strings.Builder, t memberstore.MembersByTier, title string) {
	if t.Size() == 0 {
		return
	}
	t.Iterate("", "", func(tn string, value interface{}) bool {
		_, ok := memberstore.GetTier(tn)
		if !ok {
			panic("tier not found")
		}

		power := memberstore.GetTierPower(tn, getMembers(cross))

		sb.WriteString(ufmt.Sprintf("%v from %v (VPPM %v):\n\n", title, tn, power))
		ms, _ := value.(*avl.Tree)
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
