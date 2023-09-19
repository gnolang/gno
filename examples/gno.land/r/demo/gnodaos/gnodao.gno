package gnodao

import (
	"gno.land/p/demo/avl"
	fmt "gno.land/p/demo/ufmt"
	"std"
	"strconv"
	"strings"
	"time"
)

type VoteOption uint32

const (
	YES          VoteOption = 0 // Indicates approval of the proposal in its current form.
	NO           VoteOption = 1 // Indicates disapproval of the proposal in its current form.
	NO_WITH_VETO VoteOption = 2 // Indicates stronger opposition to the proposal than simply voting No. Not available for SuperMajority-typed proposals as a simple No of 1/3 out of total votes would result in the same outcome.
	ABSTAIN      VoteOption = 3 // Indicates that the voter is impartial to the outcome of the proposal. Although Abstain votes are counted towards the quorum, they're excluded when calculating the ratio of other voting options above.
)

// GNODAO VOTE
type Vote struct {
	address   std.Address // address of the voter
	timestamp uint64      // block timestamp of the vote
	option    VoteOption  // vote option
}

type DAO struct {
	id               uint64
	uri              string    // DAO homepage link
	metadata         string    // DAO metadata reference link
	funds            uint64    // DAO managing funds
	depositHistory   []string  // deposit history - reserved for later use
	spendHistory     []string  // spend history - reserved for later use
	permissions      []string  // permissions managed on DAO - reserved for later use
	permMap          *avl.Tree // permission map - reserved for later use
	votingPowers     *avl.Tree
	totalVotingPower uint64
	votingPeriod     uint64
	voteQuorum       uint64
	threshold        uint64
	vetoThreshold    uint64
}

type ProposalStatus uint32

const (
	NIL           ProposalStatus = 0
	VOTING_PERIOD ProposalStatus = 1
	PASSED        ProposalStatus = 2
	REJECTED      ProposalStatus = 3
	FAILED        ProposalStatus = 4
)

func (s ProposalStatus) String() string {
	switch s {
	case NIL:
		return "Nil"
	case VOTING_PERIOD:
		return "VotingPeriod"
	case PASSED:
		return "Passed"
	case REJECTED:
		return "Rejected"
	case FAILED:
		return "Failed"
	}
	return ""
}

type VotingPower struct {
	address string
	power   uint64
}

type Proposal struct {
	daoId        uint64         // dao id of the proposal
	id           uint64         // unique id assigned for each proposal
	title        string         // proposal title
	summary      string         // proposal summary
	spendAmount  uint64         // amount of tokens to spend as part the proposal
	spender      std.Address    // address to receive spending tokens
	vpUpdates    []VotingPower  // updates on voting power - optional
	newMetadata  string         // new metadata for the DAO - optional
	newURI       string         // new URI for the DAO - optional
	submitTime   uint64         // proposal submission time
	voteEndTime  uint64         // vote end time for the proposal
	status       ProposalStatus // StatusNil | StatusVotingPeriod | StatusPassed | StatusRejected | StatusFailed
	votes        *avl.Tree      // votes on the proposal
	votingPowers []uint64       // voting power sum per voting option
}

// GNODAO STATE
var daos []DAO
var proposals [][]Proposal

func getDAOVotingPower(daoId uint64, address string) uint64 {
	if len(daos) <= int(daoId) {
		return 0
	}
	res, ok := daos[daoId].votingPowers.Get(address)
	if ok {
		return res.(uint64)
	}
	return 0
}

func IsDAOMember(daoId uint64, address std.Address) bool {
	return getDAOVotingPower(daoId, address.String()) > 0
}

func getVote(daoId, proposalId uint64, address std.Address) (Vote, bool) {
	if int(daoId) >= len(daos) {
		return Vote{}, false
	}

	if int(proposalId) >= len(proposals[daoId]) {
		return Vote{}, false
	}

	vote, ok := proposals[daoId][proposalId].votes.Get(address.String())
	if ok {
		return vote.(Vote), true
	}
	return Vote{}, false
}

func parseVotingPowers(daoMembers, votingPowers string) []VotingPower {
	parsedVPs := []VotingPower{}
	if len(daoMembers) == 0 {
		return parsedVPs
	}
	memberAddrs := strings.Split(daoMembers, ",")
	memberPowers := strings.Split(votingPowers, ",")
	if len(memberAddrs) != len(memberPowers) {
		panic("mismatch between members and voting powers count")
	}
	for i, memberAddr := range memberAddrs {
		power, err := strconv.Atoi(memberPowers[i])
		if err != nil {
			panic(err)
		}
		parsedVPs = append(parsedVPs, VotingPower{
			address: memberAddr,
			power:   uint64(power),
		})
	}
	return parsedVPs
}

// GNODAO FUNCTIONS
func CreateDAO(
	uri string,
	metadata string,
	daoMembers string,
	votingPowers string,
	votingPeriod uint64,
	voteQuorum uint64,
	threshold uint64,
	vetoThreshold uint64,
) {
	daoId := uint64(len(daos))
	daos = append(daos, DAO{
		id:               daoId,
		uri:              uri,
		metadata:         metadata,
		funds:            0,
		depositHistory:   []string{},
		spendHistory:     []string{},
		permissions:      []string{},
		permMap:          avl.NewTree(),
		votingPowers:     avl.NewTree(),
		totalVotingPower: 0,
		votingPeriod:     votingPeriod,
		voteQuorum:       voteQuorum,
		threshold:        threshold,
		vetoThreshold:    vetoThreshold,
	})

	parsedVPs := parseVotingPowers(daoMembers, votingPowers)
	totalVotingPower := uint64(0)
	for _, vp := range parsedVPs {
		daos[daoId].votingPowers.Set(vp.address, vp.power)
		totalVotingPower += vp.power
	}
	daos[daoId].totalVotingPower = totalVotingPower
	proposals = append(proposals, []Proposal{})
	// TODO: emit events
}

func CreateProposal(
	daoId uint64,
	title, summary string,
	spendAmount uint64, spender std.Address,
	daoMembers string,
	vpUpdates string,
	newMetadata string,
	newURI string,
) {
	caller := std.GetOrigCaller()

	// if sender is not a dao member, revert
	isCallerDaoMember := IsDAOMember(daoId, caller)
	if !isCallerDaoMember {
		panic("caller is not a dao member")
	}

	parsedVPUpdates := parseVotingPowers(daoMembers, vpUpdates)
	proposals[daoId] = append(proposals[daoId], Proposal{
		daoId:        daoId,
		id:           uint64(len(proposals[daoId])),
		title:        title,
		summary:      summary,
		spendAmount:  spendAmount,
		spender:      spender,
		vpUpdates:    parsedVPUpdates,
		newMetadata:  newMetadata,
		newURI:       newURI,
		submitTime:   uint64(time.Now().Unix()),
		voteEndTime:  uint64(time.Now().Unix()) + daos[daoId].votingPeriod,
		status:       VOTING_PERIOD,
		votes:        avl.NewTree(),
		votingPowers: []uint64{0, 0, 0, 0}, // initiate as zero for 4 vote types
	})
}

func VoteProposal(daoId, proposalId uint64, option VoteOption) {
	caller := std.GetOrigCaller()

	// if sender is not a dao member, revert
	isCallerDaoMember := IsDAOMember(daoId, caller)
	if !isCallerDaoMember {
		panic("caller is not a gnodao member")
	}

	// if invalid proposal, panic
	if int(proposalId) >= len(proposals[daoId]) {
		panic("invalid proposal id")
	}

	// if vote end time is reached panic
	if time.Now().Unix() > int64(proposals[daoId][proposalId].voteEndTime) {
		panic("vote end time reached")
	}

	// Original vote cancel
	callerVotingPower := getDAOVotingPower(daoId, caller.String())
	vote, ok := getVote(daoId, proposalId, caller)
	if ok {
		if proposals[daoId][proposalId].votingPowers[int(vote.option)] > callerVotingPower {
			proposals[daoId][proposalId].votingPowers[int(vote.option)] -= callerVotingPower
		} else {
			proposals[daoId][proposalId].votingPowers[int(vote.option)] = 0
		}
	}

	// Create a vote
	proposals[daoId][proposalId].votes.Set(caller.String(), Vote{
		address:   caller,
		timestamp: uint64(time.Now().Unix()),
		option:    option,
	})

	// Voting power by option update for new vote
	proposals[daoId][proposalId].votingPowers[int(option)] += callerVotingPower
}

// TODO: handle voting power change during voting period for other proposal
// TODO: experiment with gas limit
func TallyAndExecute(daoId, proposalId uint64) {
	caller := std.GetOrigCaller()

	// if sender is not a dao member, revert
	isCallerDaoMember := IsDAOMember(daoId, caller)
	if !isCallerDaoMember {
		panic("caller is not a gnodao member")
	}

	// validation for proposalId
	if int(proposalId) >= len(proposals[daoId]) {
		panic("invalid proposal id")
	}
	dao := daos[daoId]
	proposal := proposals[daoId][proposalId]
	votingPowers := proposal.votingPowers

	if time.Now().Unix() < int64(proposal.voteEndTime) {
		panic("proposal is in voting period")
	}

	// reference logic for tally - https://github.com/cosmos/cosmos-sdk/blob/main/x/gov/keeper/tally.go
	totalVotes := votingPowers[YES] + votingPowers[NO] + votingPowers[NO_WITH_VETO] + votingPowers[ABSTAIN]
	if totalVotes < dao.totalVotingPower*dao.voteQuorum/100 {
		proposals[daoId][proposalId].status = REJECTED
	}

	// If no one votes (everyone abstains), proposal rejected
	if totalVotes == votingPowers[ABSTAIN] {
		proposals[daoId][proposalId].status = REJECTED
	}

	// If more than 1/3 of voters veto, proposal rejected
	vetoThreshold := dao.vetoThreshold
	if votingPowers[NO_WITH_VETO] > totalVotes*vetoThreshold/100 {
		proposals[daoId][proposalId].status = REJECTED
	}

	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	threshold := dao.threshold
	if votingPowers[YES] > (totalVotes-votingPowers[ABSTAIN])*threshold/100 {
		proposals[daoId][proposalId].status = PASSED

		// TODO: spend coins when spendAmount is positive & spender is a valid address
		if proposal.spendAmount > 0 {
			if daos[daoId].funds >= proposal.spendAmount {
				daos[daoId].funds -= proposal.spendAmount
			} else {
				proposals[daoId][proposalId].status = FAILED
				return
			}
		}

		if proposal.newMetadata != "" {
			daos[daoId].metadata = proposal.newMetadata
		}

		if proposal.newURI != "" {
			daos[daoId].uri = proposal.newURI
		}

		for _, vp := range proposal.vpUpdates {
			daos[daoId].totalVotingPower -= getDAOVotingPower(daoId, vp.address)
			daos[daoId].votingPowers.Set(vp.address, vp.power)
			daos[daoId].totalVotingPower += vp.power
		}

		// TODO: contract does not own account that can hold coins - this is one of limitations
		// TODO: Adena Wallet from OnBloc - investigate on how they manage coins (swap - custody?)
		// Manual sending for funds (Address <-> Address) - Miloš Živković
		// https://github.com/gnolang/gno/blob/e392ab51bc05a5efbceaa8dbe395bac2e01ad808/tm2/pkg/crypto/keys/client/send.go#L109-L119
		return
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal rejected
	proposals[daoId][proposalId].status = REJECTED
}

func DepositDAO(daoId uint64, amount uint64) {
	caller := std.GetOrigCaller()

	// if sender is not a dao member, revert
	isCallerDaoMember := IsDAOMember(daoId, caller)
	if !isCallerDaoMember {
		panic("caller is not a gnodao member")
	}

	// TODO: send coins from caller to DAO
	// TODO: verify received amount
	// daos[daoId].depositHistory = append(daos[daoId].depositHistory, Deposit{
	// 	address: caller,
	// 	amount:  amount,
	// })
}

func GetDAO(daoId uint64) DAO {
	if int(daoId) >= len(daos) {
		panic("invalid dao id")
	}
	return daos[daoId]
}

func GetDAOs(startAfter, limit uint64) []DAO {
	max := uint64(len(daos))
	if startAfter+limit < max {
		max = startAfter + limit
	}
	return daos[startAfter:max]
}

func GetProposal(daoId, proposalId uint64) Proposal {
	if int(daoId) >= len(daos) {
		panic("invalid dao id")
	}
	if int(proposalId) >= len(proposals[daoId]) {
		panic("invalid proposal id")
	}
	return proposals[daoId][proposalId]
}

func GetProposals(daoId, startAfter, limit uint64) []Proposal {
	if int(daoId) >= len(daos) {
		panic("invalid dao id")
	}
	max := uint64(len(proposals[daoId]))
	if startAfter+limit < max {
		max = startAfter + limit
	}
	return proposals[daoId][startAfter:max]
}

func RenderVote(daoId, proposalId uint64, address std.Address) string {
	vote, found := getVote(daoId, proposalId, address)
	if !found {
		return ""
	}

	return fmt.Sprintf(`{
	"address": "%s",
	"timestamp": %d,
	"option": %d
}`, vote.address.String(), vote.timestamp, vote.option)
}

type DAOEncode struct {
	id               uint64
	uri              string // DAO homepage link
	metadata         string // DAO metadata reference link
	funds            uint64 // DAO managing funds
	totalVotingPower uint64
	votingPeriod     uint64
	voteQuorum       uint64
	threshold        uint64
	vetoThreshold    uint64
}

type ProposalEncode struct {
	daoId        uint64
	id           uint64
	title        string
	summary      string
	spendAmount  uint64
	spender      std.Address
	vpUpdates    []VotingPower
	newMetadata  string
	newURI       string
	submitTime   uint64
	voteEndTime  uint64
	status       ProposalStatus
	votingPowers []uint64
}

func GetDAOEncodeObject(dao DAO) DAOEncode {
	return DAOEncode{
		id:               dao.id,
		uri:              dao.uri,
		metadata:         dao.metadata,
		funds:            dao.funds,
		totalVotingPower: dao.totalVotingPower,
		votingPeriod:     dao.votingPeriod,
		voteQuorum:       dao.voteQuorum,
		threshold:        dao.threshold,
		vetoThreshold:    dao.vetoThreshold,
	}
}

func GetProposalEncodeObject(p Proposal) ProposalEncode {
	return ProposalEncode{
		daoId:        p.daoId,
		id:           p.id,
		title:        p.title,
		summary:      p.summary,
		spendAmount:  p.spendAmount,
		spender:      p.spender,
		vpUpdates:    p.vpUpdates,
		newMetadata:  p.newMetadata,
		newURI:       p.newURI,
		submitTime:   p.submitTime,
		voteEndTime:  p.voteEndTime,
		status:       p.status,
		votingPowers: p.votingPowers,
	}
}

func RenderDAO(daoId uint64) string {
	daoEncode := GetDAOEncodeObject(GetDAO(daoId))

	return fmt.Sprintf(`{
	"id": %d,
	"uri": "%s",
	"metadata": "%s",
	"funds": %d,
	"totalVotingPower": %d,
	"votingPeriod": %d,
	"voteQuorum": %d,
	"threshold": %d,
	"vetoThreshold": %d
}`, daoEncode.id, daoEncode.uri, daoEncode.metadata, daoEncode.funds, daoEncode.totalVotingPower, daoEncode.votingPeriod, daoEncode.voteQuorum, daoEncode.threshold, daoEncode.vetoThreshold)
}

func RenderDAOMembers(daoId uint64, start string, end string) string {
	dao := GetDAO(daoId)
	votingPowers := []VotingPower{}
	dao.votingPowers.Iterate(start, end, func(tree *avl.Tree) bool {
		power := tree.Value().(uint64)
		votingPowers = append(votingPowers, VotingPower{
			address: tree.Key(),
			power:   power,
		})
		return false
	})

	rendered := "["
	for index, votingPower := range votingPowers {
		rendered += fmt.Sprintf(`{
	"address": "%s",
	"power": %d
}`, votingPower.address, votingPower.power)
		if index != len(votingPowers)-1 {
			rendered += ",\n"
		}
	}
	rendered += "]"
	return rendered
}

func RenderDAOs(startAfter, limit uint64) string {
	daos := GetDAOs(startAfter, limit)
	daoEncodes := []DAOEncode{}
	rendered := "["
	for index, dao := range daos {
		rendered += RenderDAO(dao.id)
		if index != len(daos)-1 {
			rendered += ",\n"
		}
	}
	rendered += "]"
	return rendered
}

func RenderProposal(daoId, proposalId uint64) string {
	p := GetProposalEncodeObject(GetProposal(daoId, proposalId))
	vpUpdatesRendered := "["
	for index, vpUpdate := range p.vpUpdates {
		vpUpdatesRendered += fmt.Sprintf(`{
	"address": "%s",
	"power": %d
}`, vpUpdate.address, vpUpdate.power)
		if index != len(p.vpUpdates)-1 {
			vpUpdatesRendered += ",\n"
		}
	}
	vpUpdatesRendered += "]"

	votingPowersBySumRendered := fmt.Sprintf(`[%d, %d, %d, %d]`, p.votingPowers[0], p.votingPowers[1], p.votingPowers[2], p.votingPowers[3])

	return fmt.Sprintf(`{
	"daoId": %d,
	"id": %d,
	"title": "%s",
	"summary": "%s",
	"spendAmount": %d,
	"spender": "%s",
	"newMetadata": "%s",
	"newURI": "%s",
	"submitTime": %d,
	"voteEndTime": %d,
	"status": %d,
	"vpUpdates": %s,
	"votingPowers": %s
}`, p.daoId, p.id, p.title, p.summary, p.spendAmount, p.spender.String(), p.newMetadata, p.newURI, p.submitTime, p.voteEndTime, int(p.status), vpUpdatesRendered, votingPowersBySumRendered)
}

func RenderProposals(daoId, startAfter, limit uint64) string {
	proposals := GetProposals(daoId, startAfter, limit)
	rendered := "["
	for index, proposal := range proposals {
		rendered += RenderProposal(proposal.daoId, proposal.id)
		if index != len(proposals)-1 {
			rendered += ",\n"
		}
	}
	rendered += "]"
	return rendered
}

func Render(path string) string {
	return ""
}
