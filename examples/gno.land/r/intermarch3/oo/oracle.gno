package oo

import (
	"crypto/sha256"
	"encoding/hex"
	"std"
	"strconv"
	"time"
)

type DataRequest struct {
	Id             string
	Creator        std.Address
	Timestamp      time.Time
	AncillaryData  string
	YesNoQuestion  bool
	ProposedValue  int64
	Proposer       std.Address
	ProposerBond   int64
	Disputer       std.Address
	DisputerBond   int64
	ResolutionTime time.Time
	WinningValue   int64
	State          string // "Requested", "Proposed", "Disputed", "Resolved"
	Deadline       time.Time
}

var (
	Requests map[string]DataRequest

	admin                 = std.Address("g1qdwemgqgqm7s42gy6xcj7f22uhfdzr82st3sy3")
	ResolutionTime  int64 = 2 * int64(time.Minute.Seconds())
	RequesterReward int64 = 1 * int64(1_000_000) // in GNOT
	Bond            int64 = 2 * int64(1_000_000) // in GNOT
	Bank            std.Banker
)

func init() {
	Requests = make(map[string]DataRequest)
	Bank = std.NewBanker(std.BankerTypeRealmSend)
}

// 6 first chars of sha256(ancillaryData + timestamp)
func getUniqueId(ancillaryData string, timestamp int64) string {
	data := ancillaryData + strconv.Itoa(int(timestamp)) + strconv.Itoa(int(std.ChainHeight()))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:6]
}

// -- PUBLIC FUNCTIONS --

// RequestData allows a user to request data from the oracle.
// `RequesterReward` value needs to be sent to the contract as a reward
// You need to ask a question that can be answered with a single number like a yes/no question (0 or 1) or a specific value (e.g. ETH/USD price).
// @param ancillaryData - The human-readable question for the oracle.
// @param yesNoQuestion - True if the question is a Yes/No question.
// @param deadline - A Unix timestamp for when the request should be fulfilled (need to be more than now + 24h).
func RequestData(cur realm, ancillaryData string, yesNoQuestion bool, deadline int64) string {
	id := getUniqueId(ancillaryData, deadline)
	if _, exists := Requests[id]; exists {
		panic("Error: Request with this ID already exists.")
	}
	if ancillaryData == "" {
		panic("Error: Ancillary data cannot be empty.")
	}
	if deadline < time.Now().Add(24*time.Hour).Unix() {
		panic("Error: Deadline must be at least 24 hours in the future.")
	}

	// TODO: check that x GNOT is sent as reward
	coins := std.OriginSend()
	expected := std.Coin{Denom: "ugnot", Amount: RequesterReward}
	if len(coins) == 0 || !coins[0].IsEqual(expected) {
		panic("Error: Incorrect reward amount sent. Required: " + strconv.FormatInt(RequesterReward, 10) + " ugnot.")
	}

	request := DataRequest{
		Id:            id,
		Timestamp:     time.Now(),
		AncillaryData: ancillaryData,
		YesNoQuestion: yesNoQuestion,
		State:         "Requested",
		Deadline:      time.Unix(deadline, 0),
		Creator:       std.PreviousRealm().Address(),
	}
	Requests[id] = request
	std.Emit("DataRequested", "id", id, "timestamp", request.Timestamp.String(), "ancillaryData", ancillaryData)
	return id
}

// ProposeValue allows a user to propose a value for a requested data point.
// `Bond` value needs to be sent to the contract as a bond.
// @param id - The ID of the data request.
// @param proposedValue - The proposed answer to the question.
func ProposeValue(cur realm, id string, proposedValue int64) {
	request, exists := Requests[id]
	if !exists {
		panic("Error: No request with this ID exists.")
	}
	if request.State != "Requested" {
		panic("Error: Request is not in 'Requested' state.")
	}
	if request.YesNoQuestion && proposedValue != 0 && proposedValue != 1 {
		panic("Error: Proposed value must be 0 or 1 for yes/no questions.")
	}
	if time.Now().After(request.Deadline) {
		panic("Error: Deadline for proposal has passed.")
	}

	// TODO: check that proposerBond is x GNOT
	coins := std.OriginSend()
	expected := std.Coin{Denom: "ugnot", Amount: Bond}
	if len(coins) == 0 || !coins[0].IsEqual(expected) {
		panic("Error: Incorrect bond amount sent. Required: " + strconv.FormatInt(Bond, 10) + " ugnot")
	}

	request.ProposedValue = proposedValue
	request.Proposer = std.PreviousRealm().Address()
	request.ProposerBond = Bond
	request.ResolutionTime = time.Now().Add(time.Duration(ResolutionTime) * time.Second)
	request.State = "Proposed"
	Requests[id] = request
	std.Emit("ValueProposed", "id", id, "proposedValue", strconv.Itoa(int(proposedValue)), "proposer", request.Proposer.String(), "resolutionTime", request.ResolutionTime.String())
}

// DisputeData allows a user to dispute a proposed value.
// `Bond` value needs to be sent to the contract as a bond.
// @param id - The ID of the data request to dispute.
func DisputeData(cur realm, id string) string {
	request, exists := Requests[id]
	if !exists {
		panic("Error: No request with this ID exists.")
	}
	if request.Proposer == std.PreviousRealm().Address() {
		panic("Error: Proposer cannot dispute their own proposal.")
	}
	if request.State != "Proposed" {
		panic("Error: Request is not in 'Proposed' state.")
	}
	if time.Now().After(request.ResolutionTime) {
		panic("Error: Dispute period has ended.")
	}

	// Vérifie que la caution a bien été envoyée.
	coins := std.OriginSend()
	expected := std.Coin{Denom: "ugnot", Amount: Bond}
	if len(coins) == 0 || !coins[0].IsEqual(expected) {
		panic("Error: Incorrect bond amount sent. Required: " + strconv.FormatInt(Bond, 10) + " ugnot")
	}

	request.Disputer = std.PreviousRealm().Address()
	request.DisputerBond = Bond
	request.State = "Disputed"
	Requests[id] = request
	initiateDispute(id)
	std.Emit("DataDisputed", "id", id, "disputer", request.Disputer.String())
	return "Time: " + time.Now().String()
}

// ResolveRequest finalizes an undisputed request after the resolution period has passed.
// @param id - The ID of the data request to resolve.
func ResolveRequest(cur realm, id string) {
	request, exists := Requests[id]
	if !exists {
		panic("Error: No request with this ID exists.")
	}
	if request.State == "Disputed" {
		panic("Error: Request is in 'Disputed' state.")
	}
	if request.State == "Proposed" && request.ResolutionTime.After(time.Now()) {
		panic("Error: Resolution period has not ended yet.")
	}
	if request.State == "Requested" {
		panic("Error: Request has not been proposed yet.")
	}
	if request.State == "Resolved" {
		panic("Error: Request is already resolved.")
	}
	request.State = "Resolved"
	request.WinningValue = request.ProposedValue
	Requests[id] = request

	// TODO: Return bonds + reward to proposer
	from := std.CurrentRealm().Address()
	to := request.Proposer
	totalPayout := request.ProposerBond + RequesterReward
	payout := std.Coins{std.Coin{Denom: "ugnot", Amount: totalPayout}}
	Bank.SendCoins(from, to, payout)

	std.Emit("RequestResolved", "id", id, "winningValue", strconv.Itoa(int(request.WinningValue)))
}

// RequestResult returns the winning value of a resolved request.
// @param id - The ID of the data request.
func RequestResult(cur realm, id string) int64 {
	request, exists := Requests[id]
	if !exists {
		panic("Error: No request with this ID exists.")
	}
	if request.State != "Resolved" {
		panic("Error: Request is not resolved.")
	}
	return request.WinningValue
}

// RequesterRetreiveFund allows the original requester to get their reward back if the deadline passed without a proposal.
// @param id - The ID of the data request.
func RequesterRetreiveFund(cur realm, id string) {
	request, exists := Requests[id]
	if !exists {
		panic("Error: No request with this ID exists.")
	}
	if request.State != "Requested" {
		panic("Error: cannot retreive fund as requests fulfilled.")
	}
	if request.Creator != std.PreviousRealm().Address() {
		panic("Error: Only the creator of the request can retreive the fund.")
	}
	if request.Deadline.After(time.Now()) {
		panic("Error: Cannot retreive fund before the deadline.")
	}

	// Rembourser la récompense au créateur de la requête
	from := std.CurrentRealm().Address()
	to := request.Creator
	refund := std.Coins{std.Coin{Denom: "ugnot", Amount: RequesterReward}}
	Bank.SendCoins(from, to, refund)

	// Marquer la requête comme expirée pour éviter un double remboursement
	request.State = "Expired"
	Requests[id] = request
	std.Emit("RequestExpired", "id", id)
}

// -- ADMIN FUNCTIONS --

// SetResolutionDuration sets the duration (in seconds) for the resolution period.
// @param duration - The new resolution duration in seconds.
func SetResolutionDuration(_ realm, duration int64) {
	if std.OriginCaller() == admin {
		ResolutionTime = duration
		std.Emit("ResolutionTimeSet", "duration", strconv.Itoa(int(duration)))
	} else {
		panic("Error: Only the admin can set the resolution time.")
	}
}

// SetRequesterReward sets the reward amount for a successful proposal.
// @param reward - The new reward amount in ugnot.
func SetRequesterReward(_ realm, reward int64) {
	if std.OriginCaller() == admin {
		RequesterReward = reward
		std.Emit("RequesterRewardSet", "reward", strconv.Itoa(int(reward)))
	} else {
		panic("Error: Only the admin can set the requester reward.")
	}
}

// SetBond sets the bond amount required for proposals and disputes.
// @param bond - The new bond amount in ugnot.
func SetBond(_ realm, bond int64) {
	if std.OriginCaller() == admin {
		Bond = bond
		std.Emit("BondSet", "bond", strconv.Itoa(int(bond)))
	} else {
		panic("Error: Only the admin can set the proposer bond.")
	}
}

// ChangeAdmin transfers admin privileges to a new address.
// @param newAdmin - The address of the new admin.
func ChangeAdmin(_ realm, newAdmin std.Address) {
	if std.OriginCaller() == admin {
		admin = newAdmin
		std.Emit("AdminChanged", "newAdmin", newAdmin.String())
	} else {
		panic("Error: Only the admin can change the admin.")
	}
}

// -- VIEW FUNCTIONS --

// GetRequest returns the details of a specific data request.
// @param id - The ID of the data request.
func GetRequest(_ realm, id string) DataRequest {
	request, exists := Requests[id]
	if !exists {
		panic("Error: No request with this ID exists.")
	}
	return request
}

// GetBond returns the current bond amount.
func GetBond(_ realm) int64 {
	return Bond
}

// GetResolutionTime returns the current resolution time duration.
func GetResolutionTime(_ realm) int64 {
	return ResolutionTime
}

// GetRequesterReward returns the current requester reward amount.
func GetRequesterReward(_ realm) int64 {
	return RequesterReward
}

// GetRequestState returns the current state of a specific data request.
// @param id - The ID of the data request.
func GetRequestState(_ realm, id string) string {
	request, exists := Requests[id]
	if !exists {
		panic("Error: No request with this ID exists.")
	}
	return request.State
}

// GetAdmin returns the current admin address.
func GetAdmin(_ realm) std.Address {
	return admin
}
