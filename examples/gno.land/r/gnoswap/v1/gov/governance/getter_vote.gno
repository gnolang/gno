package governance

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoswap/v1/common"
)

func GetVoteByVoteKey(voteKey string) bool {
	return mustGetVote(voteKey)
}

func GetVoteYesByVoteKey(voteKey string) bool {
	return mustGetVoteInfo(voteKey).Yes
}

func GetVoteWeightByVoteKey(voteKey string) uint64 {
	return mustGetVoteInfo(voteKey).Weight
}

func GetVotedHeightByVoteKey(voteKey string) uint64 {
	return mustGetVoteInfo(voteKey).VotedHeight
}

func GetVotedAtByVoteKey(voteKey string) uint64 {
	return mustGetVoteInfo(voteKey).VotedAt
}

func divideVoteKeyToProposalIdAndUser(voteKey string) (proposalId uint64, user std.Address) {
	parts, err := common.Split(voteKey, ":", 2)
	if err != nil {
		panic(addDetailToError(
			errInvalidInput,
			ufmt.Sprintf("voteKey(%s) is invalid", voteKey),
		))
	}

	return parseUint64(parts[0]), std.Address(parts[1])
}
