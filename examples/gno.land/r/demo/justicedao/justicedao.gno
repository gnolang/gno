package justicedao

import (
	"gno.land/p/demo/avl"
	fmt "gno.land/p/demo/ufmt"
	vrf "gno.land/r/demo/vrf_08"
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
	numJusticeDAO    uint64 // number of justice DAO members on justice proposal
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

type JusticeProposal struct {
	id          uint64         // unique id assigned for each proposal
	title       string         // proposal title
	summary     string         // proposal summary
	vrfId       uint64         // the vrf request id being used to determine governers
	governers   []string       // the governers of the proposal
	solution    string         // proposed result of justice DAO proposal
	submitTime  uint64         // solution submission time
	voteEndTime uint64         // vote end time for the proposal
	status      ProposalStatus // StatusNil | StatusVotingPeriod | StatusPassed | StatusRejected | StatusFailed
	votes       []Vote
}

// GNODAO STATE
var daoCreated bool
var dao DAO
var proposals []Proposal
var justiceProposals []JusticeProposal

func getDAOVotingPower(address string) uint64 {
	res, ok := dao.votingPowers.Get(address)
	if ok {
		return res.(uint64)
	}
	return 0
}

func IsDAOMember(address std.Address) bool {
	return getDAOVotingPower(address.String()) > 0
}

func getVote(proposalId uint64, address std.Address) (Vote, bool) {
	if int(proposalId) >= len(proposals) {
		return Vote{}, false
	}

	vote, ok := proposals[proposalId].votes.Get(address.String())
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
	numJusticeDAO uint64,
) {
	if daoCreated {
		panic("dao already created")
	}
	dao = DAO{
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
		numJusticeDAO:    numJusticeDAO,
	}

	parsedVPs := parseVotingPowers(daoMembers, votingPowers)
	totalVotingPower := uint64(0)
	for _, vp := range parsedVPs {
		dao.votingPowers.Set(vp.address, vp.power)
		totalVotingPower += vp.power
	}
	dao.totalVotingPower = totalVotingPower
	daoCreated = true
}

func CreateProposal(
	title, summary string,
	spendAmount uint64, spender std.Address,
	daoMembers string,
	vpUpdates string,
	newMetadata string,
	newURI string,
) {
	caller := std.GetOrigCaller()

	// if sender is not a dao member, revert
	isCallerDaoMember := IsDAOMember(caller)
	if !isCallerDaoMember {
		panic("caller is not a dao member")
	}

	parsedVPUpdates := parseVotingPowers(daoMembers, vpUpdates)
	proposals = append(proposals, Proposal{
		id:           uint64(len(proposals)),
		title:        title,
		summary:      summary,
		spendAmount:  spendAmount,
		spender:      spender,
		vpUpdates:    parsedVPUpdates,
		newMetadata:  newMetadata,
		newURI:       newURI,
		submitTime:   uint64(time.Now().Unix()),
		voteEndTime:  uint64(time.Now().Unix()) + dao.votingPeriod,
		status:       VOTING_PERIOD,
		votes:        avl.NewTree(),
		votingPowers: []uint64{0, 0, 0, 0}, // initiate as zero for 4 vote types
	})
}

func VoteProposal(proposalId uint64, option VoteOption) {
	caller := std.GetOrigCaller()

	// if sender is not a dao member, revert
	isCallerDaoMember := IsDAOMember(caller)
	if !isCallerDaoMember {
		panic("caller is not a gnodao member")
	}

	// if invalid proposal, panic
	if int(proposalId) >= len(proposals) {
		panic("invalid proposal id")
	}

	// if vote end time is reached panic
	if time.Now().Unix() > int64(proposals[proposalId].voteEndTime) {
		panic("vote end time reached")
	}

	// Original vote cancel
	callerVotingPower := getDAOVotingPower(caller.String())
	vote, ok := getVote(proposalId, caller)
	if ok {
		if proposals[proposalId].votingPowers[int(vote.option)] > callerVotingPower {
			proposals[proposalId].votingPowers[int(vote.option)] -= callerVotingPower
		} else {
			proposals[proposalId].votingPowers[int(vote.option)] = 0
		}
	}

	// Create a vote
	proposals[proposalId].votes.Set(caller.String(), Vote{
		address:   caller,
		timestamp: uint64(time.Now().Unix()),
		option:    option,
	})

	// Voting power by option update for new vote
	proposals[proposalId].votingPowers[int(option)] += callerVotingPower
}

func TallyAndExecute(proposalId uint64) {
	caller := std.GetOrigCaller()

	// if sender is not a dao member, revert
	isCallerDaoMember := IsDAOMember(caller)
	if !isCallerDaoMember {
		panic("caller is not a gnodao member")
	}

	// validation for proposalId
	if int(proposalId) >= len(proposals) {
		panic("invalid proposal id")
	}
	proposal := proposals[proposalId]
	votingPowers := proposal.votingPowers

	if time.Now().Unix() < int64(proposal.voteEndTime) {
		panic("proposal is in voting period")
	}

	// reference logic for tally - https://github.com/cosmos/cosmos-sdk/blob/main/x/gov/keeper/tally.go
	totalVotes := votingPowers[YES] + votingPowers[NO] + votingPowers[NO_WITH_VETO] + votingPowers[ABSTAIN]
	if totalVotes < dao.totalVotingPower*dao.voteQuorum/100 {
		proposals[proposalId].status = REJECTED
	}

	// If no one votes (everyone abstains), proposal rejected
	if totalVotes == votingPowers[ABSTAIN] {
		proposals[proposalId].status = REJECTED
	}

	// If more than 1/3 of voters veto, proposal rejected
	vetoThreshold := dao.vetoThreshold
	if votingPowers[NO_WITH_VETO] > totalVotes*vetoThreshold/100 {
		proposals[proposalId].status = REJECTED
	}

	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	threshold := dao.threshold
	if votingPowers[YES] > (totalVotes-votingPowers[ABSTAIN])*threshold/100 {
		proposals[proposalId].status = PASSED

		if proposal.spendAmount > 0 {
			if dao.funds >= proposal.spendAmount {
				dao.funds -= proposal.spendAmount
			} else {
				proposals[proposalId].status = FAILED
				return
			}
		}

		if proposal.newMetadata != "" {
			dao.metadata = proposal.newMetadata
		}

		if proposal.newURI != "" {
			dao.uri = proposal.newURI
		}

		for _, vp := range proposal.vpUpdates {
			dao.totalVotingPower -= getDAOVotingPower(vp.address)
			dao.votingPowers.Set(vp.address, vp.power)
			dao.totalVotingPower += vp.power
		}

		// TODO: contract does not own account that can hold coins - this is one of limitations
		// TODO: Adena Wallet from OnBloc - investigate on how they manage coins (swap - custody?)
		// Manual sending for funds (Address <-> Address) - Miloš Živković
		// https://github.com/gnolang/gno/blob/e392ab51bc05a5efbceaa8dbe395bac2e01ad808/tm2/pkg/crypto/keys/client/send.go#L109-L119
		return
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal rejected
	proposals[proposalId].status = REJECTED
}

func DepositDAO(amount uint64) {
	caller := std.GetOrigCaller()

	// if sender is not a dao member, revert
	isCallerDaoMember := IsDAOMember(caller)
	if !isCallerDaoMember {
		panic("caller is not a gnodao member")
	}

	// TODO: send coins from caller to DAO
	// TODO: verify received amount
	// dao.depositHistory = append(dao.depositHistory, Deposit{
	// 	address: caller,
	// 	amount:  amount,
	// })
}

func GetProposal(proposalId uint64) Proposal {
	if int(proposalId) >= len(proposals) {
		panic("invalid proposal id")
	}
	return proposals[proposalId]
}

func GetProposals(startAfter, limit uint64) []Proposal {
	max := uint64(len(proposals))
	if startAfter+limit < max {
		max = startAfter + limit
	}
	return proposals[startAfter:max]
}

func RenderVote(proposalId uint64, address std.Address) string {
	vote, found := getVote(proposalId, address)
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
	uri              string // DAO homepage link
	metadata         string // DAO metadata reference link
	funds            uint64 // DAO managing funds
	totalVotingPower uint64
	votingPeriod     uint64
	voteQuorum       uint64
	threshold        uint64
	vetoThreshold    uint64
	numJusticeDAO    uint64
}

type ProposalEncode struct {
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
		uri:              dao.uri,
		metadata:         dao.metadata,
		funds:            dao.funds,
		totalVotingPower: dao.totalVotingPower,
		votingPeriod:     dao.votingPeriod,
		voteQuorum:       dao.voteQuorum,
		threshold:        dao.threshold,
		vetoThreshold:    dao.vetoThreshold,
		numJusticeDAO:    dao.numJusticeDAO,
	}
}

func GetProposalEncodeObject(p Proposal) ProposalEncode {
	return ProposalEncode{
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

func RenderDAO() string {
	daoEncode := GetDAOEncodeObject(dao)

	return fmt.Sprintf(`{
	"uri": "%s",
	"metadata": "%s",
	"funds" %d,
	"totalVotingPower": %d,
	"votingPeriod": %d,
	"voteQuorum": %d,
	"threshold": %d,
	"vetoThreshold": %d,
	"numJusticeDAO": %d
}`, daoEncode.uri, daoEncode.metadata, daoEncode.funds, daoEncode.totalVotingPower, daoEncode.votingPeriod, daoEncode.voteQuorum, daoEncode.threshold, daoEncode.vetoThreshold, daoEncode.numJusticeDAO)
}

func RenderDAOMembers(start string, end string) string {
	votingPowers := []VotingPower{}
	dao.votingPowers.Iterate(start, end, func(key string, value interface{}) bool {
		power := value.(uint64)
		votingPowers = append(votingPowers, VotingPower{
			address: key,
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

func RenderProposal(proposalId uint64) string {
	p := GetProposalEncodeObject(GetProposal(proposalId))
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
}`, p.id, p.title, p.summary, p.spendAmount, p.spender.String(), p.newMetadata, p.newURI, p.submitTime, p.voteEndTime, int(p.status), vpUpdatesRendered, votingPowersBySumRendered)
}

func RenderProposals(startAfter, limit uint64) string {
	proposals := GetProposals(startAfter, limit)
	rendered := "["
	for index, proposal := range proposals {
		rendered += RenderProposal(proposal.id)
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

func CreateJusticeProposal(
	title, summary string,
) {
	reqId := vrf.RequestRandomWords(1)
	justiceProposals = append(justiceProposals, JusticeProposal{
		id:          uint64(len(justiceProposals)),
		title:       title,
		summary:     summary,
		vrfId:       reqId,
		governers:   []string{},
		solution:    "",
		voteEndTime: 0,
		status:      NIL,
		votes:       []Vote{},
	})
}

func GetDAOMembers() []string {
	members := []string{}
	dao.votingPowers.Iterate("", "", func(key string, value interface{}) bool {
		members = append(members, key)
		return false
	})
	return members
}

func DetermineJusticeDAOMembers(id uint64) {
	if int(id) >= len(justiceProposals) {
		panic("invalid justice DAO proposal id")
	}

	members := GetDAOMembers()
	for i := uint64(0); i < dao.numJusticeDAO; i++ {
		if len(members) == 0 {
			break
		}
		randomNum := vrf.RandomValueFromWordsWithIndex(id, i)
		index := int(randomNum) % len(members)
		member := members[index]

		// remove the index from the list
		members[index] = members[len(members)-1]
		members = members[:len(members)-1]
		justiceProposals[id].governers = append(justiceProposals[id].governers, member)
		justiceProposals[id].votes = append(justiceProposals[id].votes, Vote{})
	}
}

func ProposeJusticeDAOSolution(id uint64, solution string) {
	if int(id) >= len(justiceProposals) {
		panic("invalid justice DAO proposal id")
	}

	justiceProposals[id].solution = solution
	justiceProposals[id].submitTime = uint64(time.Now().Unix())
	justiceProposals[id].voteEndTime = uint64(time.Now().Unix()) + dao.votingPeriod
	justiceProposals[id].status = VOTING_PERIOD
}

func IsJusticeDAOMember(id uint64, member std.Address) bool {
	for _, governer := range justiceProposals[id].governers {
		if governer == member.String() {
			return true
		}
	}
	return false
}

func VoteJusticeSolutionProposal(id uint64, option VoteOption) {
	caller := std.GetOrigCaller()

	// if sender is not a justice dao member, revert
	isCallerJusticeDaoMember := IsJusticeDAOMember(id, caller)
	if !isCallerJusticeDaoMember {
		panic("caller is not a gnodao member")
	}

	// if invalid proposal, panic
	if int(id) >= len(justiceProposals) {
		panic("invalid proposal id")
	}

	// if vote end time is reached panic
	if time.Now().Unix() > int64(justiceProposals[id].voteEndTime) {
		panic("vote end time reached")
	}

	// Create a vote
	for i, govern := range justiceProposals[id].governers {
		if govern == caller.String() {
			justiceProposals[id].votes[i] = Vote{
				address:   caller,
				timestamp: uint64(time.Now().Unix()),
				option:    option,
			}
			break
		}
	}
}

func TallyAndExecuteJusticeSolution(proposalId uint64) {
	caller := std.GetOrigCaller()

	// if sender is not a dao member, revert
	isCallerJusticeDaoMember := IsJusticeDAOMember(proposalId, caller)
	if !isCallerJusticeDaoMember {
		panic("caller is not a justice DAO member")
	}

	// validation for proposalId
	if int(proposalId) >= len(justiceProposals) {
		panic("invalid justice DAO proposal id")
	}
	proposal := justiceProposals[proposalId]

	if time.Now().Unix() < int64(proposal.voteEndTime) {
		panic("justice DAO proposal is in voting period")
	}

	numYesVotes := uint64(0)
	for i, govern := range proposal.governers {
		vote := justiceProposals[proposalId].votes[i]
		if vote.option == YES {
			numYesVotes++
		}
	}

	// If more than 2/3 votes Yes, let it pass
	if numYesVotes > 0 && numYesVotes*3 >= uint64(len(proposal.governers))*2 {
		justiceProposals[proposalId].status = PASSED
	} else {
		justiceProposals[proposalId].status = REJECTED
	}
}

func RenderJusticeDAOProposal(proposalId uint64) string {
	// if invalid proposal, panic
	if int(proposalId) >= len(justiceProposals) {
		panic("invalid proposal id")
	}
	p := justiceProposals[proposalId]

	governers := strings.Join(p.governers, ",")

	voteEncodedObjects := []string{}
	for _, vote := range p.votes {
		voteEncodedObjects = append(voteEncodedObjects, fmt.Sprintf(`{
	"address": "%s",
	"timestamp": %d,
	"option": %d
}`, vote.address.String(), vote.timestamp, vote.option))
	}
	voteEncoded := strings.Join(voteEncodedObjects, ",\n")

	return fmt.Sprintf(`{
	"id": %d,
	"title": "%s",
	"summary": "%s",
	"vrfId": %d,
	"governers": [%s],
	"solution": "%s",
	"submitTime": %d,
	"voteEndTime": %d,
	"status": %d,
	"votes": [%s]
}`, p.id, p.title, p.summary, p.vrfId, governers, p.solution, p.submitTime, p.voteEndTime, int(p.status), voteEncoded)
}

func GetJusticeDAOProposals(startAfter, limit uint64) []JusticeProposal {
	max := uint64(len(justiceProposals))
	if startAfter+limit < max {
		max = startAfter + limit
	}
	return justiceProposals[startAfter:max]
}

func RenderJusticeDAOProposals(startAfter, limit uint64) string {
	proposals := GetJusticeDAOProposals(startAfter, limit)
	rendered := "["
	for index, proposal := range proposals {
		rendered += RenderJusticeDAOProposal(proposal.id)
		if index != len(proposals)-1 {
			rendered += ",\n"
		}
	}
	rendered += "]"
	return rendered
}
