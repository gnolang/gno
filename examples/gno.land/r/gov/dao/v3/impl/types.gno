package impl

import (
	"strings"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
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
	YesVotes memberstore.MembersByTier
	NoVotes  memberstore.MembersByTier
	AllVotes memberstore.MembersByTier

	Accepted bool
	Denied   bool

	DeniedReason string

	TiersAllowedToVote []string

	TotalPower float64 // TotalPower is the power of all the members existing when this proposal was created.
}

func getMembers(cur realm) memberstore.MembersByTier {
	return memberstore.Get()
}

func newProposalStatus(allowedToVote []string) *proposalStatus {
	yv := memberstore.NewMembersByTier()
	yv.SetTier(memberstore.T1)
	yv.SetTier(memberstore.T2)
	yv.SetTier(memberstore.T3)
	nv := memberstore.NewMembersByTier()
	nv.SetTier(memberstore.T1)
	nv.SetTier(memberstore.T2)
	nv.SetTier(memberstore.T3)
	av := memberstore.NewMembersByTier()
	av.SetTier(memberstore.T1)
	av.SetTier(memberstore.T2)
	av.SetTier(memberstore.T3)

	return &proposalStatus{
		YesVotes: yv,
		NoVotes:  nv,
		AllVotes: av,

		TiersAllowedToVote: allowedToVote,

		TotalPower: getMembers(cross).GetTotalPower(),
	}
}

func (ps *proposalStatus) YesPercent() float64 {
	var yp float64

	memberstore.IterateTiers(func(tn string, tier memberstore.Tier) bool {
		power := memberstore.GetTierPower(tn, getMembers(cross))
		ts := ps.YesVotes.GetTierSize(tn)

		yp = yp + (power * float64(ts))

		return false
	})

	return (yp / ps.TotalPower) * 100
}

func (ps *proposalStatus) NoPercent() float64 {
	var np float64

	memberstore.IterateTiers(func(tn string, tier memberstore.Tier) bool {
		power := memberstore.GetTierPower(tn, getMembers(cross))
		ts := ps.NoVotes.GetTierSize(tn)
		np = np + (power * float64(ts))

		return false
	})

	return (np / ps.TotalPower) * 100
}

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
		}
	} else {
		sb.WriteString("- **Proposal is open for votes**\n")
	}

	sb.WriteString("- Tiers eligible to vote: ")
	sb.WriteString(strings.Join(ps.TiersAllowedToVote, ", "))
	sb.WriteString("\n")

	sb.WriteString(ufmt.Sprintf("- YES PERCENT: %v%%\n", ps.YesPercent()))
	sb.WriteString(ufmt.Sprintf("- NO PERCENT: %v%%\n", ps.NoPercent()))

	return sb.String()
}

func StringifyVotes(ps *proposalStatus) string {
	var sb strings.Builder

	writeVotes(&sb, ps.YesVotes, "YES")
	writeVotes(&sb, ps.NoVotes, "NO")

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
			sb.WriteString("\n")
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

	return out
}
