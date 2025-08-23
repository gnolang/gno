package oo

import (
	"time"
	"std"
	"strconv"

	"crypto/sha256"
	"encoding/hex"
)

type Vote struct {
	RequestId string
	Voter std.Address
	TokenAmount int64
	Hash string
	Value int64
	Revealed bool
}

type Voter struct {
	HasVoted bool
	VoteIndex int64
}

type Dispute struct {
	RequestId string
	Votes []Vote
	NbResolvedVotes int64
	Voters map[std.Address]Voter
	IsResolved bool
	WinningValue int64
	EndTime time.Time
	EndRevealTime time.Time
}

var (
	Disputes map[string]Dispute
	DisputeDuration int64 = 2 * int64(time.Minute.Seconds())
	RevealDuration int64 = 2 * int64(time.Minute.Seconds())
	VotePrice int64 = 1 * int64(1_000_000) // in GNOT
)

func init() {
	Disputes = make(map[string]Dispute)
}

func initiateDispute(id string) {
	if _, exists := Disputes[id]; exists {
		panic("Error: Dispute for this request already exists.")
	}
	dispute := Dispute{
		RequestId: id,
		Votes:     []Vote{},
		Voters:    make(map[std.Address]Voter),
		IsResolved: false,
		EndTime: time.Now().Add(time.Duration(DisputeDuration) * time.Second),
		EndRevealTime: time.Now().Add(time.Duration(DisputeDuration + RevealDuration) * time.Second),
	}
	Disputes[id] = dispute
	std.Emit("DisputeInitiated", "id", id)
}

// -- PUBLIC FUNCTIONS --

// VoteOnDispute allows a user to commit a vote during a dispute.
// @param id - The ID of the dispute (same as Request ID).
// @param hash - The SHA256 hash of the vote value and a secret salt.
func VoteOnDispute(cur realm, id string, hash string) {
	dispute, exists := Disputes[id]
	if !exists {
		panic("Error: No dispute with this ID exists.")
	}
	r, _ := Requests[id]
	if r.Proposer == std.OriginCaller() || r.Disputer == std.OriginCaller() {
		panic("Error: Proposer and Disputer cannot vote in this dispute.")
	}
	if dispute.IsResolved {
		panic("Error: Dispute is already resolved.")
	}
	if time.Now().After(dispute.EndTime) {
		panic("Error: Vote period has ended.")
	}
	// TODO: check that voter has at least VotePrice GNOT to stake
	// get voter token balance
	vote := Vote{
		RequestId:  id,
		Voter:      std.OriginCaller(),
		TokenAmount: 1, // TODO: actually get voter's token balance
		Hash:       hash,
		Revealed:   false,
	}
	if dispute.Voters[vote.Voter].HasVoted {
		panic("Error: Voter has already voted in this dispute.")
	}
	dispute.Votes = append(dispute.Votes, vote)
	dispute.Voters[vote.Voter] = Voter{HasVoted: true, VoteIndex: int64(len(dispute.Votes) - 1)}
	Disputes[id] = dispute
	std.Emit("VoteSubmitted", "id", id, "voter", vote.Voter.String())
}

// RevealVote allows a user to reveal their vote after the voting period has ended.
// @param id - The ID of the dispute (same as Request ID).
// @param value - The actual value of the vote.
// @param salt - The secret salt used to generate the vote hash.
func RevealVote(cur realm, id string, value int64, salt string) {
	dispute, exists := Disputes[id]
	if !exists {
		panic("Error: No dispute with this ID exists.")
	}
	if dispute.IsResolved {
		panic("Error: Dispute is resolved.")
	}
	if time.Now().Before(dispute.EndTime) {
		panic("Error: Vote period has not ended yet.")
	}
	if time.Now().After(dispute.EndRevealTime) {
		panic("Error: Reveal period has ended.")
	}
	voter := dispute.Voters[std.OriginCaller()]
	if !voter.HasVoted {
		panic("Error: Voter did not participate in this dispute.")
	}
	vote := dispute.Votes[voter.VoteIndex]
	if vote.Revealed {
		panic("Error: Vote already revealed.")
	}

	// Verify the hash
	res := sha256.Sum256([]byte(strconv.FormatInt(value, 10) + salt))
	expectedHash := hex.EncodeToString(res[:])
	if (vote.Hash != expectedHash) {
		panic("Error: Hash does not match the revealed value and salt.")
	}
	vote.Value = value
	vote.Revealed = true
	dispute.NbResolvedVotes += 1
	dispute.Votes[voter.VoteIndex] = vote
	Disputes[id] = dispute
	std.Emit("VoteRevealed", "id", id, "voter", vote.Voter.String(), "value", strconv.Itoa(int(value)))
}

// ResolveDispute finalizes a dispute after the reveal period, tallying votes and setting the winning value.
// @param id - The ID of the dispute to resolve (same as Request ID).
func ResolveDispute(cur realm, id string) {
	dispute, exists := Disputes[id]
	if !exists {
		panic("Error: No dispute with this ID exists.")
	}
	if dispute.IsResolved {
		panic("Error: Dispute is already resolved.")
	}
	if time.Now().Before(dispute.EndTime) {
		panic("Error: Dispute period has not ended yet.")
	}
	val := resolve(id)
	dispute.WinningValue = val
	dispute.IsResolved = true
	Disputes[id] = dispute
	// Update the original request with the winning value
	request, exists := Requests[id]

	request.ProposedValue = val
	request.State = "Resolved"
	Requests[id] = request
	std.Emit("DisputeResolved", "id", id, "winningValue", strconv.Itoa(int(val)))
	std.Emit("RequestResolved", "id", id, "winningValue", strconv.Itoa(int(val)))
	// TODO: Distribuate bond + reward between winning voters and if disputer win, give a part to him
}

// -- admin functions --

// SetDisputeDuration sets the duration (in seconds) for the voting period.
// @param duration - The new dispute duration in seconds.
func SetDisputeDuration(_ realm, duration int64) {
	if std.OriginCaller() == admin {
		DisputeDuration = duration * int64(time.Second)
	} else {
		panic("Error: Only admin can set dispute duration.")
	}
}

// SetRevealDuration sets the duration (in seconds) for the reveal period.
// @param duration - The new reveal duration in seconds.
func SetRevealDuration(_ realm, duration int64) {
	if std.OriginCaller() == admin {
		RevealDuration = duration * int64(time.Second)
	} else {
		panic("Error: Only admin can set reveal duration.")
	}
}

// SetVotePrice sets the price (in ugnot) to cast a vote.
// @param price - The new vote price.
func SetVotePrice(_ realm, price int64) {
	if std.OriginCaller() == admin {
		VotePrice = price
	} else {
		panic("Error: Only admin can set vote price.")
	}
}

// -- view functions --

// GetDispute returns the details of a specific dispute.
// @param id - The ID of the dispute (same as Request ID).
func GetDispute(_ realm, id string) Dispute {
	dispute, exists := Disputes[id]
	if !exists {
		panic("Error: No dispute with this ID exists.")
	}
	return dispute
}

// GetDisputeDuration returns the current dispute duration.
func GetDisputeDuration(_ realm) int64 {
	return DisputeDuration
}

// GetVotePrice returns the current vote price.
func GetVotePrice(_ realm) int64 {
	return VotePrice
}

// GetDisputeEndTime returns the end time of the voting period for a specific dispute.
// @param id - The ID of the dispute (same as Request ID).
func GetDisputeEndTime(_ realm, id string) time.Time {
	dispute, exists := Disputes[id]
	if !exists {
		panic("Error: No dispute with this ID exists.")
	}
	return dispute.EndTime
}

// GetDisputeVotesAmount returns the total number of votes cast in a dispute.
// @param id - The ID of the dispute (same as Request ID).
func GetDisputeVotesAmount(_ realm, id string) int64 {
	dispute, exists := Disputes[id]
	if !exists {
		panic("Error: No dispute with this ID exists.")
	}
	return int64(len(dispute.Votes))
}

// GetRevealEndTime returns the end time of the reveal period for a specific dispute.
// @param id - The ID of the dispute (same as Request ID).
func GetRevealEndTime(_ realm, id string) time.Time {
	dispute, exists := Disputes[id]
	if !exists {
		panic("Error: No dispute with this ID exists.")
	}
	return dispute.EndRevealTime
}

// GetRevealDuration returns the current reveal duration.
func GetRevealDuration(_ realm) int64 {
	return RevealDuration
}