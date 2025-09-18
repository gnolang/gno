package commondao

import (
	"errors"
	"std"

	"gno.land/p/nt/avl"
)

// VotingContext contains current voting context, which includes a voting
// record for the votes that has been submitted for a proposal and also
// the list of DAO members.
type VotingContext struct {
	VotingRecord ReadonlyVotingRecord
	Members      ReadonlyMemberStorage
}

// NewVotingContext creates a new voting context.
func NewVotingContext(r *VotingRecord, s MemberStorage) (VotingContext, error) {
	if r == nil {
		return VotingContext{}, errors.New("voting record is required")
	}

	if s == nil {
		return VotingContext{}, errors.New("member storage is required")
	}

	members := MustNewReadonlyMemberStorage(s)
	return VotingContext{
		VotingRecord: r.Readonly(),
		Members:      *members,
	}, nil
}

// MustNewVotingContext creates a new voting context or panics on error.
func MustNewVotingContext(r *VotingRecord, s MemberStorage) VotingContext {
	ctx, err := NewVotingContext(r, s)
	if err != nil {
		panic(err)
	}
	return ctx
}

// CollectVotes returns an AVL tree containing votes of members from one or more groups.
// Returned tree uses member account address as key and `commondao.Vote` as value.
func CollectVotes(ctx VotingContext, groups ...string) (avl.Tree, error) {
	if len(groups) == 0 {
		return avl.Tree{}, errors.New("one or more group names are required to collect votes")
	}

	var (
		votes    avl.Tree
		grouping = ctx.Members.Grouping()
	)

	for _, name := range groups {
		group, found := grouping.Get(name)
		if !found {
			return avl.Tree{}, errors.New("member group not found: " + name)
		}

		group.Members().IterateByOffset(0, group.Members().Size(), func(member std.Address) bool {
			v, found := ctx.VotingRecord.GetVote(member)
			if found {
				votes.Set(member.String(), v)
			}
			return false
		})
	}
	return votes, nil
}
