package gnodao

import (
	"fmt"
	"std"
	"strings"
	"testing"
	"time"
)

// This test file is disabled until we port this code to new avl

var (
	caller     std.Address = "g1rel7980x4y257yh30umy3jx223efwakvnabcde"
	caller1    std.Address = "g1rel7980x4y257yh30umy3jx223efwakvnaaaaa"
	caller2    std.Address = "g1rel7980x4y257yh30umy3jx223efwakvnbbbbb"
	daoMembers             = []string{
		"g1rel7980x4y257yh30umy3jx223efwakvnaaaaa",
		"g1rel7980x4y257yh30umy3jx223efwakvnbbbbb",
		"g1rel7980x4y257yh30umy3jx223efwakvnccccc",
	}
)

var (
	daoMembersPacked   = strings.Join(daoMembers, ",")
	votingPowersPacked = "1,2,3"
)

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func TestIsDAOMember(t *testing.T) {
	daos = []DAO{}
	CreateDAO(
		"https://gnodao1.org",
		"https://metadata.gnodao1.org",
		daoMembersPacked,
		votingPowersPacked,
		86400*2,
		30,
		10,
		10,
	)

	// TODO: what package is ideal to use for checks?
	if IsDAOMember(0, caller) != false {
		t.Errorf("Should not be false")
	}
	if IsDAOMember(0, caller1) != true {
		t.Errorf("Should be true")
	}
	if IsDAOMember(0, caller2) != true {
		t.Errorf("Should be true")
	}
}

func TestCreateDAO(t *testing.T) {
	daos = []DAO{}
	proposals = [][]Proposal{}
	CreateDAO(
		"https://gnodao1.org",
		"https://metadata.gnodao1.org",
		daoMembersPacked,
		votingPowersPacked,
		86400*2,
		30,
		10,
		10,
	)
	if len(daos) != 1 {
		t.Errorf("Number of daos be 1")
	}
	if len(proposals) != 1 {
		t.Errorf("Number of daos be 1")
	}
	dao := daos[0]
	if dao.id != 0 {
		t.Errorf("first DAO id should be 0")
	}
	if dao.uri != "https://gnodao1.org" {
		t.Errorf("dao uri not set properly")
	}
	if dao.metadata != "https://metadata.gnodao1.org" {
		t.Errorf("dao metadata not set properly")
	}
	if dao.funds != 0 {
		t.Errorf("dao funds not set properly")
	}
	if len(dao.depositHistory) != 0 {
		t.Errorf("dao deposit history not set properly")
	}
	if len(dao.spendHistory) != 0 {
		t.Errorf("dao spend history not set properly")
	}
	if len(dao.permissions) != 0 {
		t.Errorf("dao permissions not set properly")
	}
	if dao.permMap == nil {
		t.Errorf("dao permission map not set properly")
	}
	if getDAOVotingPower(0, caller.String()) != 0 {
		t.Errorf("voting power not set properly")
	}
	if getDAOVotingPower(0, caller1.String()) != 1 {
		t.Errorf("voting power not set properly")
	}
	if getDAOVotingPower(0, caller2.String()) != 2 {
		t.Errorf("voting power not set properly")
	}
	if dao.totalVotingPower != 6 {
		t.Errorf("totalVotingPower not set properly")
	}
	if dao.votingPeriod != 86400*2 {
		t.Errorf("votingPeriod not set properly")
	}
	if dao.voteQuorum != 30 {
		t.Errorf("voteQuorum not set properly")
	}
	if dao.threshold != 10 {
		t.Errorf("threshold not set properly")
	}
	if dao.vetoThreshold != 10 {
		t.Errorf("vetoThreshold not set properly")
	}
}

func TestCreateProposal(t *testing.T) {
	daos = []DAO{}
	proposals = [][]Proposal{}
	CreateDAO("https://gnodao1.org", "https://metadata.gnodao1.org", daoMembersPacked, votingPowersPacked, 86400*2, 30, 10, 10)

	assertPanic(t, func() {
		std.TestSetOrigCaller(caller)
		CreateProposal(0, "DAO fund bootstrap proposal", "Proposal to bootstrap DAO fund.", 0, caller, "", "", "", "")
	})

	std.TestSetOrigCaller(caller1)
	CreateProposal(0, "DAO fund bootstrap proposal", "Proposal to bootstrap DAO fund.", 0, caller, "", "", "", "")
	proposal := proposals[0][0]
	if proposal.daoId != 0 {
		t.Errorf("proposal daoId should be 0")
	}
	if proposal.id != 0 {
		t.Errorf("proposal id should be 0")
	}
	if proposal.title != "DAO fund bootstrap proposal" {
		t.Errorf("proposal title not set properly")
	}
	if proposal.summary != "Proposal to bootstrap DAO fund." {
		t.Errorf("proposal summary not set properly")
	}
	if proposal.spendAmount != 0 {
		t.Errorf("proposal spendAmount not set properly")
	}
	if proposal.spender != caller {
		t.Errorf("proposal spender not set properly")
	}
	if len(proposal.vpUpdates) != 0 {
		t.Errorf("proposal vpUpdates not set properly")
	}
	if proposal.newMetadata != "" {
		t.Errorf("proposal newMetadata not set properly")
	}
	if proposal.newURI != "" {
		t.Errorf("proposal newURI not set properly")
	}
	if proposal.submitTime != uint64(time.Now().Unix()) {
		t.Errorf("proposal submitTime not set properly")
	}
	if proposal.voteEndTime != uint64(time.Now().Unix())+daos[0].votingPeriod {
		t.Errorf("proposal voteEndTime not set properly")
	}
	if proposal.status != VOTING_PERIOD {
		t.Errorf("proposal status not set properly")
	}
	if proposal.votes == nil {
		t.Errorf("proposal votes not set properly")
	}
	if len(proposal.votingPowers) != 4 {
		t.Errorf("proposal votingPowers not set properly")
	}
}

func TestVoteProposal(t *testing.T) {
	daos = []DAO{}
	proposals = [][]Proposal{}
	CreateDAO("https://gnodao1.org", "https://metadata.gnodao1.org", daoMembersPacked, votingPowersPacked, 86400*2, 30, 10, 10)
	std.TestSetOrigCaller(caller1)
	CreateProposal(0, "DAO fund bootstrap proposal", "Proposal to bootstrap DAO fund.", 0, caller, "", "", "", "")
	assertPanic(t, func() { // invalid dao id
		std.TestSetOrigCaller(caller1)
		VoteProposal(1, 0, YES)
	})
	assertPanic(t, func() { // invalid proposal id
		std.TestSetOrigCaller(caller1)
		VoteProposal(0, 1, YES)
	})
	assertPanic(t, func() { // not dao member
		std.TestSetOrigCaller(caller)
		VoteProposal(0, 0, YES)
	})

	// vote and check result is set properly
	std.TestSetOrigCaller(caller1)
	VoteProposal(0, 0, YES)
	vote, found := getVote(0, 0, caller1)
	if !found {
		t.Errorf("proposal vote not set")
	}
	if vote.address != caller1 {
		t.Errorf("vote address not set properly")
	}
	if vote.timestamp != uint64(time.Now().Unix()) {
		t.Errorf("vote timestamp not set properly")
	}
	if vote.option != YES {
		t.Errorf("vote option not set properly")
	}
	if proposals[0][0].votingPowers[int(YES)] != 1 {
		t.Errorf("votePowers by vote option not set properly")
	}

	// vote again with different option and check result
	VoteProposal(0, 0, NO)
	vote, found = getVote(0, 0, caller1)
	if vote.option != NO {
		t.Errorf("vote option not set properly")
	}
	if proposals[0][0].votingPowers[int(YES)] != 0 {
		t.Errorf("votePowers for YES not set properly")
	}
	if proposals[0][0].votingPowers[int(NO)] != 1 {
		t.Errorf("votePowers for NO not set properly")
	}

	// test vote end time already reached
	assertPanic(t, func() { // not dao member
		std.TestSetOrigCaller(caller)
		proposals[0][0].voteEndTime = uint64(time.Now().Unix()) - 1
		VoteProposal(0, 0, YES)
	})
}

func TestTallyAndExecute(t *testing.T) {
	daos = []DAO{}
	proposals = [][]Proposal{}
	CreateDAO("https://gnodao1.org", "https://metadata.gnodao1.org", daoMembersPacked, votingPowersPacked, 86400*2, 30, 10, 10)
	std.TestSetOrigCaller(caller1)
	CreateProposal(0, "DAO fund bootstrap proposal", "Proposal to bootstrap DAO fund.", 0, caller, caller.String(), "1", "newMetadata.com", "newURI.com")
	assertPanic(t, func() { // invalid dao id
		std.TestSetOrigCaller(caller1)
		TallyAndExecute(1, 0)
	})
	assertPanic(t, func() { // invalid proposal id
		std.TestSetOrigCaller(caller1)
		TallyAndExecute(0, 1)
	})
	assertPanic(t, func() { // not dao member
		std.TestSetOrigCaller(caller)
		TallyAndExecute(0, 0)
	})
	assertPanic(t, func() { // vote end time not pass
		std.TestSetOrigCaller(caller1)
		TallyAndExecute(0, 0)
	})

	// vote end time to be reached
	proposals[0][0].voteEndTime = uint64(time.Now().Unix()) - 1

	// quorum not reached
	std.TestSetOrigCaller(caller1)
	TallyAndExecute(0, 0)
	if proposals[0][0].status != REJECTED {
		t.Errorf("proposal should be REJECTED for vote quorum")
	}

	// everyone abstains
	proposals[0][0].votingPowers[ABSTAIN] = daos[0].totalVotingPower
	std.TestSetOrigCaller(caller1)
	TallyAndExecute(0, 0)
	if proposals[0][0].status != REJECTED {
		t.Errorf("proposal should be REJECTED for all abstains")
	}

	// more than 1/3 vote with NO_WITH_VETO
	proposals[0][0].votingPowers[ABSTAIN] = daos[0].totalVotingPower / 2
	proposals[0][0].votingPowers[NO_WITH_VETO] = daos[0].totalVotingPower / 2
	std.TestSetOrigCaller(caller1)
	TallyAndExecute(0, 0)
	if proposals[0][0].status != REJECTED {
		t.Errorf("proposal should be REJECTED for NO_WITH_VETO")
	}

	// all YES vote
	proposals[0][0].votingPowers[ABSTAIN] = 0
	proposals[0][0].votingPowers[NO_WITH_VETO] = 0
	proposals[0][0].votingPowers[YES] = daos[0].totalVotingPower
	std.TestSetOrigCaller(caller1)
	TallyAndExecute(0, 0)
	if proposals[0][0].status != PASSED {
		t.Errorf("proposal should be PASSED")
	}
	if getDAOVotingPower(0, caller.String()) != 1 {
		t.Errorf("voting power not set properly")
	}
	if daos[0].metadata != "newMetadata.com" {
		t.Errorf("metadata not set properly")
	}
	if daos[0].uri != "newURI.com" {
		t.Errorf("uri not set properly")
	}
}

func TestDepositDAO(t *testing.T) {
	daos = []DAO{}
	proposals = [][]Proposal{}
	CreateDAO("https://gnodao1.org", "https://metadata.gnodao1.org", daoMembersPacked, votingPowersPacked, 86400*2, 30, 10, 10)

	// panic when not a dao member
	assertPanic(t, func() {
		std.TestSetOrigCaller(caller)
		DepositDAO(0, 100)
	})

	// not panics
	std.TestSetOrigCaller(caller1)
	DepositDAO(0, 100)
}

func TestGetDAO(t *testing.T) {
	daos = []DAO{}
	proposals = [][]Proposal{}
	CreateDAO("https://gnodao1.org", "https://metadata.gnodao1.org", daoMembersPacked, votingPowersPacked, 86400*2, 30, 10, 10)

	// panic when invalid dao id
	assertPanic(t, func() {
		std.TestSetOrigCaller(caller)
		GetDAO(100)
	})

	// success when valid dao id
	dao := GetDAO(0)
	if dao.uri != "https://gnodao1.org" {
		t.Errorf("uri not set properly")
	}
}

func TestGetDAOs(t *testing.T) {
	daos = []DAO{}
	proposals = [][]Proposal{}
	gotDaos := GetDAOs(0, 10)
	if len(gotDaos) != 0 {
		t.Errorf("invalid number of daos")
	}
	CreateDAO("https://gnodao1.org", "https://metadata.gnodao1.org", daoMembersPacked, votingPowersPacked, 86400*2, 30, 10, 10)
	CreateDAO("https://gnodao2.org", "https://metadata.gnodao2.org", daoMembersPacked, votingPowersPacked, 86400*2, 30, 10, 10)

	gotDaos = GetDAOs(0, 0)
	if len(gotDaos) != 0 {
		t.Errorf("invalid number of daos")
	}

	gotDaos = GetDAOs(0, 10)
	if len(gotDaos) != 2 {
		t.Errorf("invalid number of daos")
	}

	gotDaos = GetDAOs(0, 1)
	if len(gotDaos) != 1 {
		t.Errorf("invalid number of daos")
	}
}

func TestGetProposal(t *testing.T) {
	daos = []DAO{}
	proposals = [][]Proposal{}
	CreateDAO("https://gnodao1.org", "https://metadata.gnodao1.org", daoMembersPacked, votingPowersPacked, 86400*2, 30, 10, 10)
	std.TestSetOrigCaller(caller1)
	CreateProposal(0, "DAO fund bootstrap proposal", "Proposal to bootstrap DAO fund.", 0, caller, "", "", "newMetadata.com", "newURI.com")

	// panic when invalid dao id
	assertPanic(t, func() {
		GetProposal(1, 0)
	})

	// panic when invalid proposal id
	assertPanic(t, func() {
		GetProposal(0, 1)
	})

	// success when valid dao id and proposal id
	proposal := GetProposal(0, 0)
	if proposal.title != "DAO fund bootstrap proposal" {
		t.Errorf("title not set properly")
	}
}

func TestGetProposals(t *testing.T) {
	daos = []DAO{}
	proposals = [][]Proposal{}
	assertPanic(t, func() { // invalid dao id
		GetProposals(0, 0, 10)
	})
	CreateDAO("https://gnodao1.org", "https://metadata.gnodao1.org", daoMembersPacked, votingPowersPacked, 86400*2, 30, 10, 10)
	std.TestSetOrigCaller(caller1)
	CreateProposal(0, "proposal #1", "Proposal to bootstrap DAO fund.", 0, caller, "", "", "newMetadata.com", "newURI.com")
	CreateProposal(0, "proposal #2", "Proposal to bootstrap DAO fund.", 0, caller, "", "", "newMetadata.com", "newURI.com")

	gotProposals := GetProposals(0, 0, 0)
	if len(gotProposals) != 0 {
		t.Errorf("invalid number of proposals")
	}

	gotProposals = GetProposals(0, 0, 10)
	if len(gotProposals) != 2 {
		t.Errorf("invalid number of proposals")
	}

	gotProposals = GetProposals(0, 0, 1)
	if len(gotProposals) != 1 {
		t.Errorf("invalid number of proposals")
	}

	renderedProposals := RenderProposals(0, 0, 1)
	expected := `[{
	"daoId": 0,
	"id": 0,
	"title": "proposal #1",
	"summary": "Proposal to bootstrap DAO fund.",
	"spendAmount": 0,
	"spender": "g1rel7980x4y257yh30umy3jx223efwakvnabcde",
	"newMetadata": "newMetadata.com",
	"newURI": "newURI.com",
	"submitTime": 1234567890,
	"voteEndTime": 1234740690,
	"status": 1,
	"vpUpdates": [],
	"votingPowers": [0, 0, 0, 0]
}]`
	if renderedProposals != expected {
		t.Errorf("invalid proposal rendering, expected:\n%s\n, got:\n%s\n", expected, renderedProposals)
	}
}

func TestRender(t *testing.T) {
	if Render("") != "" {
		t.Errorf("Render function should empty")
	}
}
