package govdao

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type MembersByTier struct {
	*avl.Tree // *avl.Tree[string]*avl.Tree[std.Address]*Member
}

func NewMembersByTier() MembersByTier {
	return MembersByTier{avl.NewTree()}
}

func (mbt MembersByTier) GetTierSize(tn string) int {
	tv, ok := mbt.Get(tn)
	if !ok {
		return 0
	}

	tree, ok := tv.(*avl.Tree)
	if !ok {
		return 0
	}

	return tree.Size()
}

func (mbt MembersByTier) SetMember(tier string, addr std.Address, member *Member) {
	if ok := mbt.Has(tier); !ok {
		mbt.Set(tier, avl.NewTree())
	}

	ms, _ := mbt.Get(tier)
	mst := ms.(*avl.Tree)

	mst.Set(string(addr), member)
}

func (mbt MembersByTier) GetMember(addr std.Address) (m *Member, t string) {
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		mst, ok := msv.(*avl.Tree)
		if !ok {
			panic("MembersByTier values can only be avl.Tree")
		}

		return mst.Iterate("", "", func(ma string, msv interface{}) bool {
			if ma != string(addr) {
				return false
			}

			mst, ok := msv.(*Member)
			if !ok {
				panic("MembersByTier values can only be *Member")
			}

			m = mst
			t = tn

			return true
		})
	})

	return
}

func (mbt MembersByTier) GetTotalPower(tbn TiersByName) float64 {
	var out float64
	mbt.Iterate("", "", func(tn string, msv interface{}) bool {
		mst, ok := msv.(*avl.Tree)
		if !ok {
			panic("MembersByTier values can only be avl.Tree")
		}

		tier, ok := tbn.GetTier(tn)
		if !ok {
			// tier does not exists, so we cannot count power from this tier
			return false
		}

		out = out + tier.PowerHandler(mbt, tbn)
	})

	return out
}

type TiersByName struct {
	*avl.Tree // *avl.Tree[string]Tier
}

func (tbn TiersByName) GetTier(tn string) (Tier, bool) {
	val, ok := tbn.Get(tn)
	if !ok {
		return Tier{}, false
	}

	t, ok := val.(Tier)
	if !ok {
		panic("TiersByName must contains only Tier types")
	}

	return t, true
}

type Member struct {
	// TODO IDEA: add here a tree to store possible future metadata for the user, like the age and number of invitation points as examples.
	// Possible data keys for members must be defined on the Law to make possible future new validations
}

type Tier struct {
	BasePower    float64
	MaxSize      func(membersByTier MembersByTier, tiersByName TiersByName) int
	MinSize      func(membersByTier MembersByTier, tiersByName TiersByName) int
	PowerHandler func(membersByTier MembersByTier, tiersByName TiersByName) float64
}

type Law struct {
	Tiers                        TiersByName
	VotingPercentageNeeded       int
	Payment                      int // The payment itself will be good, but not exceed 90th percentile of senior software architect roles in the highest paid city globally.
	CheckMemberAddProposal       func(membersByTier MembersByTier, proposer std.Address, proposedTier string, proposedAddr std.Address) error
	CheckMemberPromoteProposal   func(membersByTier MembersByTier, proposer std.Address, proposedTier string, proposedAddr std.Address) error
	CheckMemberWithdrawnProposal func(membersByTier MembersByTier, proposer std.Address, proposedAddr std.Address) error
	CheckProposalWithdrawn       func(membersByTier MembersByTier, proposer std.Address) error
}

type Executor struct {
	metadata   Metadata
	callbackFn func(m Metadata) error
}

func (e *Executor) String() string {
	return e.metadata.String()
}

func NewExecutor(fn func(m Metadata) error, metadata Metadata) *Executor {
	return &Executor{
		metadata:   metadata,
		callbackFn: fn,
	}
}

type Metadata interface {
	IsMetadata()
	String() string
}

type Proposals struct {
	*avl.Tree // map[int]*Proposal
}

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

type Proposal struct {
	Author std.Address

	Title       string
	Description string

	Executor *Executor
}

func (p *Proposal) String() string {
	return ufmt.Sprintf(`
Title: %s
Proposed by: %s

%s

This proposal contains the following metadata:
%s`, p.Title, p.Author, p.Description, p.Executor.String())
}

type proposalStatus struct {
	YesVotes MembersByTier
	NoVotes  MembersByTier
	AllVotes MembersByTier

	Inactive bool //TODO: maybe these flags don't belong here
	Accepted bool
	Denied   bool
}

func newProposalStatus() *proposalStatus {
	return &proposalStatus{
		YesVotes: NewMembersByTier(),
		NoVotes:  NewMembersByTier(),
		AllVotes: NewMembersByTier(),
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
	}

	sb.WriteString("\n")

	if ps.YesVotes.Size() == 0 &&
		ps.NoVotes.Size() == 0 {
		sb.WriteString("No one voted yet")
	}

	writeVotes(&sb, ps.YesVotes, "Voting YES:")
	writeVotes(&sb, ps.NoVotes, "Voting NO:")

	ts := theLaw.Tiers
	totalPower := members.GetTotalPower(ts)
	yesPower := ps.YesVotes.GetTotalPower(ts)
	noPower := ps.NoVotes.GetTotalPower(ts)

	sb.WriteString("\n")
	sb.WriteString(ufmt.Sprintf("YES PERCENT: %v%%", (yesPower/totalPower)*100))
	sb.WriteString("\n")
	sb.WriteString(ufmt.Sprintf("NO PERCENT: %v%%", (noPower/totalPower)*100))

	return sb.String()
}

func writeVotes(sb *strings.Builder, t MembersByTier, title string) {
	if t.Size() != 0 {
		sb.WriteString(title)
		sb.WriteString("\n")
		t.Iterate("", "", func(tn string, value interface{}) bool {
			sb.WriteString("From " + tn + ":\n")
			ms, _ := value.(*avl.Tree)
			ms.Iterate("", "", func(addr string, _ interface{}) bool {
				sb.WriteString("\t-" + string(addr) + "\n")

				return false
			})

			return false
		})
	}
}
