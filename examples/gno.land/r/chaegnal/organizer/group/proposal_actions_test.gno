package group

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/r/chaegnal/organizer"
)

// Test VoteOnProposal
func Test_VoteOnProposal_Success(t *testing.T) {
	testing.SetOriginCaller("g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	CreateProposals(cross, "9", "test", "test", "", "", "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")

	uassert.NotPanics(t, func() { VoteOnProposal(cross, "9", "1", "1", "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g") })
}

func Test_VoteOnProposal_InvalidVoteId(t *testing.T) {
	testing.SetOriginCaller("g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	err := revive(func() {
		VoteOnProposal(cross, "7", "abc", "1", "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")
	})

	errorMsg := "invalid vote id"

	if err == nil || err.(string) != errorMsg {
		t.Fatalf("unexpected panic message: got %q, want %q", err.(string), errorMsg)
	}
}

func Test_VoteOnProposal_InvalidVoteChoice(t *testing.T) {
	testing.SetOriginCaller("g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	CreateProposals(cross, "10", "test", "test", "", "", "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")

	err := revive(func() {
		VoteOnProposal(cross, "10", "1", "abc", "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")
	})

	errorMsg := "invalid vote choice"

	if err == nil || err.(string) != errorMsg {
		t.Fatalf("unexpected panic message: got %q, want %q", err.(string), errorMsg)
	}
}

func Test_VoteOnProposal_NorYayNorNay(t *testing.T) {
	testing.SetOriginCaller("g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	CreateProposals(cross, "11", "test", "test", "", "", "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")

	err := revive(func() {
		VoteOnProposal(cross, "11", "1", "2", "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g")
	})

	errorMsg := "vote should be yay (1) or nay (0)"

	if err == nil || err.(string) != errorMsg {
		t.Fatalf("unexpected panic message: got %q, want %q", err.(string), errorMsg)
	}
}
