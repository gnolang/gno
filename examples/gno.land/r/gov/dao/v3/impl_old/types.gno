package govdao

import (
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao/v3/types"
)

// Proposals contains all the proposals indexed by their ID.
type Proposals struct {
	*avl.Tree // map[int]*Proposal
}

// ProposalsStatuses contains the status of all the proposals indexed by the proposal ID.
type ProposalsStatuses struct {
	*avl.Tree // map[int]*proposalStatus
}

func (pss ProposalsStatuses) GetStatus(id string) *proposalStatus {
	psv, ok := pss.Get(id)
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
	YesVotes types.MembersByTier
	NoVotes  types.MembersByTier
	AllVotes types.MembersByTier

	Inactive bool //TODO: maybe these flags don't belong here
	Accepted bool
	Denied   bool

	DeniedReason string
}

func newProposalStatus() *proposalStatus {
	return &proposalStatus{
		YesVotes: types.NewMembersByTier(),
		NoVotes:  types.NewMembersByTier(),
		AllVotes: types.NewMembersByTier(),
	}
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
	writeVotes(&sb, ps.NoVotes, "Voting NO")

	ts := theLaw.Tiers
	totalPower := members.GetTotalPower(ts)
	yesPower := ps.YesVotes.GetTotalPower(ts)
	noPower := ps.NoVotes.GetTotalPower(ts)

	sb.WriteString("\n")
	sb.WriteString(ufmt.Sprintf("YES PERCENT: %v%%", (yesPower/totalPower)*100))
	sb.WriteString("\n\n")
	sb.WriteString(ufmt.Sprintf("NO PERCENT: %v%%", (noPower/totalPower)*100))

	return sb.String()
}

func writeVotes(sb *strings.Builder, t types.MembersByTier, title string) {
	if t.Size() != 0 {
		t.Iterate("", "", func(tn string, value interface{}) bool {
			sb.WriteString(title + " from " + tn + ":\n")
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
