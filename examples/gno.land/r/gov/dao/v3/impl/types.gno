package impl

import (
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao/proxy"
	"gno.land/r/gov/dao/v3/memberstore"
)

type Law struct {
	Supermajority float64
}

// ProposalsStatuses contains the status of all the proposals indexed by the proposal ID.
type ProposalsStatuses struct {
	*avl.Tree // map[int]*proposalStatus
}

func NewProposalsStatuses() ProposalsStatuses{
	return ProposalsStatuses{avl.NewTree()}
}

func (pss ProposalsStatuses) GetStatus(id proxy.ProposalID) *proposalStatus {
	pids := ufmt.Sprintf("%v", int(id))
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

	Inactive bool
	Accepted bool
	Denied   bool

	DeniedReason string
}

func newProposalStatus() *proposalStatus {
	return &proposalStatus{
		YesVotes: memberstore.NewMembersByTier(),
		NoVotes:  memberstore.NewMembersByTier(),
		AllVotes: memberstore.NewMembersByTier(),
	}
}

func (ps *proposalStatus) YesPercent() float64 {
	tp := memberstore.Get().GetTotalPower()
	var yp float64

	memberstore.Tiers.Iterate("", "", func(tn string, ti interface{}) bool {
		tier, ok := ti.(memberstore.Tier)
		if !ok {
			panic("type must be memberstore.Tier")
		}

		power := tier.PowerHandler(memberstore.Get(), memberstore.Tiers)
		ts := ps.YesVotes.GetTierSize(tn)

		yp = yp + (power * float64(ts))

		return false
	})

	return (yp/tp) * 100
}

func (ps *proposalStatus) NoPercent() float64 {
	tp := memberstore.Get().GetTotalPower()
	var np float64

	memberstore.Tiers.Iterate("", "", func(tn string, ti interface{}) bool {
		tier, ok := ti.(memberstore.Tier)
		if !ok {
			panic("type must be memberstore.Tier")
		}

		power := tier.PowerHandler(memberstore.Get(), memberstore.Tiers)
		ts := ps.NoVotes.GetTierSize(tn)
		np = np + (power * float64(ts))

		return false
	})

	return (np/tp) * 100
}

func (ps *proposalStatus) String() string {
	var sb strings.Builder
	sb.WriteString("Proposal Status:\n")

	if ps.Accepted {
		sb.WriteString("\nPROPOSAL HAS BEEN ACCEPTED\n")
	}
	if ps.Denied {
		sb.WriteString("\nPROPOSAL HAS BEEN DENIED\n")
		if ps.DeniedReason != "" {
			sb.WriteString("REASON: ")
			sb.WriteString(ps.DeniedReason)
		}
	}

	sb.WriteString("\n")

	writeVotes(&sb, ps.YesVotes, "Voting YES")
	sb.WriteString("\n\n")
	writeVotes(&sb, ps.NoVotes, "Voting NO")

	sb.WriteString("\n")
	sb.WriteString(ufmt.Sprintf("YES PERCENT: %v%%", ps.YesPercent()))
	sb.WriteString("\n\n")
	sb.WriteString(ufmt.Sprintf("NO PERCENT: %v%%", ps.NoPercent()))
	
	return sb.String()
}

func writeVotes(sb *strings.Builder, t memberstore.MembersByTier, title string) {
	if t.Size() != 0 {
		t.Iterate("", "", func(tn string, value interface{}) bool {
			tier, ok := memberstore.Tiers.GetTier(tn)
			if !ok {
				panic("tier not found")
			}

			power := tier.PowerHandler(memberstore.Get(), memberstore.Tiers)

			sb.WriteString(ufmt.Sprintf("%v from %v (VPPM %v):\n", title, tn, power))
			ms, _ := value.(*avl.Tree)
			ms.Iterate("", "", func(addr string, _ interface{}) bool {
				sb.WriteString("\n")
				sb.WriteString("- " + string(addr) + "\n")

				return false
			})

			return false
		})
	}
}

func (p *proxy.Proposal) String() string {
	return ufmt.Sprintf(`
Title: %s

Proposed by: %s

%s

This proposal contains the following metadata:

%s`, p.Title, p.Author, p.Description, p.Executor.String())
}