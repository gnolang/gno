package basedao

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	carol = testutils.TestAddress("carol")
	dave  = testutils.TestAddress("dave")
)

func TestNewDAO(cur realm, t *testing.T) {
	initialRoles := []RoleInfo{
		{Name: "admin", Description: "Admin is the superuser"},
	}
	initialMembers := []Member{
		{alice.String(), []string{"admin"}},
		{bob.String(), []string{}},
		{carol.String(), []string{}},
	}
	conf := &Config{
		Name:             "My DAO",
		Description:      "My DAO Description",
		Members:          NewMembersStore(initialRoles, initialMembers),
		GetProfileString: func(addr address, field, def string) string { return "" },
	}

	daoRealm := testing.NewCodeRealm("gno.land/r/testing/daorealm")
	testing.SetOriginCaller(alice)
	testing.SetRealm(daoRealm)
	wrapper := New(conf, cur)
	dao := wrapper.Private()
	roles := dao.Members.GetRoles()
	if len(roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(roles))
	}
	if roles[0] != "admin" {
		t.Errorf("Expected role 'admin', got %s", roles[0])
	}

	for _, member := range initialMembers {
		addr := member.Address
		if !dao.Members.IsMember(addr) {
			t.Errorf("Expected member %s to be a member", addr)
		}
		if len(member.Roles) == 1 && !dao.Members.HasRole(addr, member.Roles[0]) {
			t.Errorf("Expected member %s to have role %s", addr, member.Roles[0])
		}
	}

	urequire.Equal(t, 4, dao.Core.ResourcesCount(), "expected 4 resources")
	urequire.Equal(t, dao.RealmPkgPath, cur.PkgPath())

	// XXX: check realm and profile
}

func TestPropose(cur realm, t *testing.T) {
	members := []Member{
		{alice.String(), []string{"admin"}},
		{bob.String(), []string{}},
		{carol.String(), []string{}},
	}
	tdao := newTestingDAO(0, cur, t, 0.6, members)

	type testNewProposalInput struct {
		proposalReq daokit.ProposalRequest
		proposer    address
	}

	type tesNewProposalExpected struct {
		title        string
		description  string
		proposer     address
		messsageType string
		panic        bool
	}

	type testNewProposal struct {
		input    testNewProposalInput
		expected tesNewProposalExpected
	}

	type testNewProposalTable map[string]testNewProposal

	tests := testNewProposalTable{
		"Success": {
			input: testNewProposalInput{
				proposalReq: tdao.mockProposalRequest,
				proposer:    alice,
			},
			expected: tesNewProposalExpected{
				title:        "My Proposal",
				description:  "My Proposal Description",
				proposer:     alice,
				messsageType: "valid",
				panic:        false,
			},
		},
		"Non-member": {
			input: testNewProposalInput{
				proposalReq: tdao.mockProposalRequest,
				proposer:    dave,
			},
			expected: tesNewProposalExpected{
				panic: true,
			},
		},
		"Unknown action type": {
			input: testNewProposalInput{
				proposalReq: tdao.unknownProposalRequest,
				proposer:    alice,
			},
			expected: tesNewProposalExpected{
				panic: true,
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(cur realm, t *testing.T) {
			*tdao.mockHandlerCalled = false
			testing.SetRealm(testing.NewUserRealm(test.input.proposer))

			// tdao.propose crosses into basedao via an IIFE same-realm
			// cross; any panic inside the crossing frame is cross-realm
			// and cannot be caught by defer/recover here. Use revive
			// to capture the cross-boundary abort.
			var id uint64
			exc := revive(func() {
				id = tdao.propose(0, cur, test.input.proposalReq)
			})
			if test.expected.panic {
				if exc == nil {
					t.Errorf("Expected panic, got none")
				}
				return
			}
			if exc != nil {
				t.Errorf("Unexpected panic: %v", exc)
				return
			}

			urequire.False(t, *tdao.mockHandlerCalled, "execute should not be called")

			proposal := tdao.privdao.Core.Proposals.GetProposal(id)
			if proposal.Title != test.expected.title {
				t.Errorf("Expected title %s, got %s", test.expected.title, proposal.Title)
			}
			if proposal.Description != test.expected.description {
				t.Errorf("Expected description %s, got %s", test.expected.description, proposal.Description)
			}
			if proposal.ProposerID != test.expected.proposer.String() {
				t.Errorf("Expected proposer %s, got %s", test.expected.proposer, proposal.ProposerID)
			}
			if proposal.Action.Type() != test.expected.messsageType {
				t.Errorf("Expected action type %s, got %s", test.expected.messsageType, proposal.Action.Type())
			}
		})
	}
}

func TestVote(cur realm, t *testing.T) {
	members := []Member{
		{alice.String(), []string{"admin"}},
		{bob.String(), []string{}},
		{carol.String(), []string{}},
	}
	tdao := newTestingDAO(0, cur, t, 0.6, members)

	testing.SetRealm(testing.NewUserRealm(alice))
	tdao.propose(0, cur, tdao.mockProposalRequest)

	type testVoteInput struct {
		proposalID uint64
		vote       daocond.Vote
		voter      address
	}

	type testVoteExpected struct {
		eval  bool
		panic string
	}

	type testVote struct {
		input    testVoteInput
		expected testVoteExpected
	}

	type testVoteTable map[string]testVote

	tests := testVoteTable{
		"Success no": {
			input: testVoteInput{
				proposalID: 1,
				vote:       "no",
				voter:      alice,
			},
			expected: testVoteExpected{
				eval: false,
			},
		},
		"Success yes": {
			input: testVoteInput{
				proposalID: 1,
				vote:       "yes",
				voter:      alice,
			},
			expected: testVoteExpected{
				eval: true,
			},
		},
		"Unknown proposal": {
			input: testVoteInput{
				proposalID: 2,
				vote:       "yes",
				voter:      alice,
			},
			expected: testVoteExpected{
				panic: "proposal not found",
			},
		},
		"Non-member": {
			input: testVoteInput{
				proposalID: 1,
				vote:       "yes",
				voter:      dave,
			},
			expected: testVoteExpected{
				panic: "caller is not a member",
			},
		},
		"Invalid vote": {
			input: testVoteInput{
				proposalID: 1,
				vote:       "very-long-vote-very-long-vote-very-long-vote",
				voter:      alice,
			},
			expected: testVoteExpected{
				panic: "invalid vote",
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(cur realm, t *testing.T) {
			// SetRealm must run here, not inside `run`: SetRealm targets
			// the deepest non-testing frame's fr.Cur HIV. A non-crossing
			// `run` closure has no fr.Cur, so a SetRealm-from-inside-run
			// silently no-ops on the closure-captured cur.
			testing.SetRealm(testing.NewUserRealm(test.input.voter))
			run := func() {
				tdao.vote(0, cur, test.input.proposalID, test.input.vote)
			}

			*tdao.mockHandlerCalled = false

			if test.expected.panic != "" {
				// tdao.vote crosses into basedao; panic is cross-realm
				// and shows up as an abort, not a same-realm panic.
				uassert.AbortsWithMessage(t, cur, test.expected.panic, run)
			} else {
				urequire.NotPanics(t, cur, run)
				urequire.False(t, *tdao.mockHandlerCalled, "execute should not be called")
				proposal := tdao.privdao.Core.Proposals.GetProposal(test.input.proposalID)
				eval := proposal.Condition.Eval(proposal.Ballot)
				if eval != test.expected.eval {
					t.Errorf("Expected eval %t, got %t", test.expected.eval, eval)
				}
			}
		})
	}
}

func TestExecuteProposal(cur realm, t *testing.T) {
	members := []Member{
		{alice.String(), []string{"admin"}},
		{bob.String(), []string{}},
		{carol.String(), []string{}},
	}
	tdao := newTestingDAO(0, cur, t, 0.6, members)

	testing.SetRealm(testing.NewUserRealm(alice))
	tdao.propose(0, cur, tdao.mockProposalRequest)

	type testExecuteInput struct {
		proposalID uint64
		executor   address
		haveVote   bool
		voter      address
	}

	type testExecuteExpected struct {
		panic bool
	}

	type testExecute struct {
		input    testExecuteInput
		expected testExecuteExpected
	}

	type testExecuteTable map[string]testExecute

	tests := testExecuteTable{
		"Conditions not met": {
			input: testExecuteInput{
				proposalID: 1,
				executor:   alice,
				haveVote:   false,
				voter:      alice,
			},
			expected: testExecuteExpected{
				panic: true,
			},
		},
		"Success": {
			input: testExecuteInput{
				proposalID: 1,
				executor:   alice,
				haveVote:   true,
				voter:      alice,
			},
			expected: testExecuteExpected{
				panic: false,
			},
		},
		"Unknown proposal": {
			input: testExecuteInput{
				proposalID: 2,
				executor:   alice,
				haveVote:   false,
				voter:      alice,
			},
			expected: testExecuteExpected{
				panic: true,
			},
		},
		"Non-member": {
			input: testExecuteInput{
				proposalID: 1,
				executor:   dave,
				haveVote:   false,
				voter:      alice,
			},
			expected: testExecuteExpected{
				panic: true,
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(cur realm, t *testing.T) {
			if test.input.haveVote {
				testing.SetRealm(testing.NewUserRealm(test.input.voter))
				tdao.vote(0, cur, test.input.proposalID, "yes")
			}

			testing.SetRealm(testing.NewUserRealm(test.input.executor))

			// tdao.execute crosses into basedao; cross-realm panic is
			// not catchable by defer/recover, so use revive.
			exc := revive(func() {
				tdao.execute(0, cur, test.input.proposalID)
			})
			if test.expected.panic {
				if exc == nil {
					t.Errorf("Expected panic, got none")
				}
				return
			}
			if exc != nil {
				t.Errorf("Unexpected panic: %v", exc)
				return
			}

			proposal := tdao.privdao.Core.Proposals.GetProposal(test.input.proposalID)

			if proposal.Status != daokit.ProposalStatusExecuted {
				t.Errorf("Expected status %s, got %s", daokit.ProposalStatusExecuted, proposal.Status)
			}
		})
	}
}

func TestInstantExecute(cur realm, t *testing.T) {
	members := []Member{
		{alice.String(), []string{"admin"}},
		{bob.String(), []string{}},
		{carol.String(), []string{}},
	}
	tdao := newTestingDAO(0, cur, t, 0.6, members)

	type testInstantExecuteInput struct {
		proposalReq daokit.ProposalRequest
		executor    address
	}

	type testInstantExecuteExpected struct {
		panic bool
	}

	type testInstantExecute struct {
		input    testInstantExecuteInput
		expected testInstantExecuteExpected
	}

	type testInstantExecuteTable map[string]testInstantExecute

	tests := testInstantExecuteTable{
		"Success": {
			input: testInstantExecuteInput{
				proposalReq: tdao.mockProposalRequest,
				executor:    alice,
			},
			expected: testInstantExecuteExpected{
				panic: false,
			},
		},
		"Unknown action type": {
			input: testInstantExecuteInput{
				proposalReq: tdao.unknownProposalRequest,
				executor:    alice,
			},
			expected: testInstantExecuteExpected{
				panic: true,
			},
		},
		"Non-member": {
			input: testInstantExecuteInput{
				proposalReq: tdao.mockProposalRequest,
				executor:    dave,
			},
			expected: testInstantExecuteExpected{
				panic: true,
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(cur realm, t *testing.T) {
			testing.SetRealm(testing.NewUserRealm(test.input.executor))

			// tdao.propose crosses into basedao; cross-realm panic
			// is not catchable by defer/recover, so use revive.
			exc := revive(func() {
				tdao.propose(0, cur, test.input.proposalReq)
			})
			if test.expected.panic {
				if exc == nil {
					t.Errorf("Expected panic, got none")
				}
			} else if exc != nil {
				t.Errorf("Unexpected panic: %v", exc)
			}
		})
	}
}

func TestGetMembers(cur realm, t *testing.T) {
	members := []Member{
		{alice.String(), []string{"admin"}},
		{bob.String(), []string{}},
		{carol.String(), []string{}},
	}
	tdao := newTestingDAO(0, cur, t, 0.6, members)

	expectedMembers := []string{alice.String(), bob.String(), carol.String()}
	m := tdao.privdao.Members.GetMembers()
	if len(m) != len(expectedMembers) {
		t.Errorf("Expected %d members, got %d", len(expectedMembers), len(m))
	}

	for _, eMember := range expectedMembers {
		if !tdao.privdao.Members.IsMember(eMember) {
			t.Errorf("Expected member %s to be a member", eMember)
		}
	}
}

func TestAddMemberProposal(cur realm, t *testing.T) {
	members := []Member{
		{alice.String(), []string{"admin"}},
	}
	tdao := newTestingDAO(0, cur, t, 0.2, members)

	if tdao.privdao.Members.IsMember(bob.String()) {
		t.Errorf("Expected member %s to not be a member", bob.String())
	}

	testing.SetRealm(testing.NewUserRealm(alice))
	tdao.instantExecute(0, cur, daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action: NewAddMemberAction(&ActionAddMember{
			Address: bob,
			Roles:   []string{"admin"},
		}),
	})

	if !tdao.privdao.Members.IsMember(bob.String()) {
		t.Errorf("Expected member %s to be a member", bob.String())
	}

	if !tdao.privdao.Members.HasRole(bob.String(), "admin") {
		t.Errorf("Expected member %s to have role 'admin'", bob.String())
	}

	proposalWithUnknownRole := daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action: NewAddMemberAction(&ActionAddMember{
			Address: bob,
			Roles:   []string{"unknown"},
		}),
	}
	// Cross-realm panic from instantExecute is not catchable by
	// defer/recover; use revive.
	if revive(func() { tdao.instantExecute(0, cur, proposalWithUnknownRole) }) == nil {
		t.Errorf("Expected panic, got none")
	}
}

func TestRemoveMemberProposal(cur realm, t *testing.T) {
	members := []Member{
		{alice.String(), []string{"admin"}},
		{bob.String(), []string{"admin"}},
	}
	tdao := newTestingDAO(0, cur, t, 0.2, members)

	if !tdao.privdao.Members.IsMember(bob.String()) {
		t.Errorf("Expected member %s to be a member", bob.String())
	}

	if !tdao.privdao.Members.HasRole(bob.String(), "admin") {
		t.Errorf("Expected member %s to have role 'admin'", bob.String())
	}
	testing.SetRealm(testing.NewUserRealm(alice))
	tdao.instantExecute(0, cur, daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action:      NewRemoveMemberAction(bob),
	})

	if tdao.privdao.Members.IsMember(bob.String()) {
		t.Errorf("Expected user %s to not be a member", bob.String())
	}

	if tdao.privdao.Members.HasRole(bob.String(), "admin") {
		t.Errorf("Expected user %s to not have role 'admin'", bob.String())
	}
}

func TestAddRoleToUserProposal(cur realm, t *testing.T) {
	members := []Member{
		{alice.String(), []string{"admin"}},
		{bob.String(), []string{}},
	}
	tdao := newTestingDAO(0, cur, t, 0.2, members)

	if tdao.privdao.Members.HasRole(bob.String(), "admin") {
		t.Errorf("Expected member %s to not have role 'admin'", bob.String())
	}

	testing.SetRealm(testing.NewUserRealm(alice))
	tdao.instantExecute(0, cur, daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action:      NewAssignRoleAction(&ActionAssignRole{Address: bob, Role: "admin"}),
	})
	if !tdao.privdao.Members.HasRole(bob.String(), "admin") {
		t.Errorf("Expected member %s to have role 'admin'", bob.String())
	}

	// Cross-realm panics aren't catchable by defer/recover; use revive
	// per failing case.
	if revive(func() {
		tdao.instantExecute(0, cur, daokit.ProposalRequest{
			Title:       "My Proposal",
			Description: "My Proposal Description",
			Action:      NewAssignRoleAction(&ActionAssignRole{Address: alice, Role: "unknown"}),
		})
	}) == nil {
		t.Errorf("Expected panic on unknown role, got none")
	}

	if revive(func() {
		tdao.instantExecute(0, cur, daokit.ProposalRequest{
			Title:       "My Proposal",
			Description: "My Proposal Description",
			Action:      NewAssignRoleAction(&ActionAssignRole{Address: carol, Role: "admin"}),
		})
	}) == nil {
		t.Errorf("Expected panic on non-member, got none")
	}

	if revive(func() {
		tdao.instantExecute(0, cur, daokit.ProposalRequest{
			Title:       "My Proposal",
			Description: "My Proposal Description",
			Action:      NewAssignRoleAction(&ActionAssignRole{Address: bob, Role: "admin"}),
		})
	}) == nil {
		t.Errorf("Expected panic on already-assigned role, got none")
	}
}

func TestRemoveRoleFromUserProposal(cur realm, t *testing.T) {
	members := []Member{
		{
			alice.String(),
			[]string{"admin"},
		},
		{
			bob.String(),
			[]string{"admin"},
		},
	}
	tdao := newTestingDAO(0, cur, t, 0.2, members)

	if !tdao.privdao.Members.HasRole(bob.String(), "admin") {
		t.Errorf("Expected member %s to have role 'admin'", bob.String())
	}
	testing.SetRealm(testing.NewUserRealm(alice))
	tdao.instantExecute(0, cur, daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action:      NewUnassignRoleAction(&ActionUnassignRole{Address: bob, Role: "admin"}),
	})

	if tdao.privdao.Members.HasRole(bob.String(), "admin") {
		t.Errorf("Expected member %s to not have role 'admin'", bob.String())
	}

	proposalWithUnknowkRole := daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action:      NewUnassignRoleAction(&ActionUnassignRole{Address: alice, Role: "unknown"}),
	}
	testing.SetRealm(testing.NewUserRealm(alice))
	if revive(func() { tdao.instantExecute(0, cur, proposalWithUnknowkRole) }) == nil {
		t.Errorf("Expected panic on unknown role, got none")
	}

	proposalWithNonMember := daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action:      NewUnassignRoleAction(&ActionUnassignRole{Address: carol, Role: "admin"}),
	}
	testing.SetRealm(testing.NewUserRealm(alice))
	if revive(func() { tdao.instantExecute(0, cur, proposalWithNonMember) }) == nil {
		t.Errorf("Expected panic on non-member, got none")
	}

	proposalWithNonRole := daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action:      NewUnassignRoleAction(&ActionUnassignRole{Address: bob, Role: "admin"}),
	}
	testing.SetRealm(testing.NewUserRealm(alice))
	if revive(func() { tdao.instantExecute(0, cur, proposalWithNonRole) }) == nil {
		t.Errorf("Expected panic on already-unassigned role, got none")
	}
}

type testingDAOContext struct {
	dao                    *DAOWrapper
	privdao                *DAOPrivate
	mockHandlerCalled      *bool
	mockHandler            daokit.ActionHandler
	mockProposalRequest    daokit.ProposalRequest
	unknownProposalRequest daokit.ProposalRequest
}

func newTestingDAO(_ int, rlm realm, t *testing.T, threshold float64, members []Member) *testingDAOContext {
	roles := []RoleInfo{{Name: "admin", Description: "Admin is the superuser"}}
	membersStore := NewMembersStore(roles, members)
	initialCondition := daocond.MembersThreshold(threshold, membersStore.IsMember, membersStore.MembersCount)

	conf := &Config{
		Name:             "My DAO",
		Description:      "My DAO Description",
		Members:          membersStore,
		InitialCondition: initialCondition,
		GetProfileString: func(addr address, field, def string) string { return "" },
	}

	daoRealm := testing.NewCodeRealm("gno.land/r/testing/daorealm")
	testing.SetOriginCaller(alice)
	testing.SetRealm(daoRealm)
	pubdao := New(conf, rlm)
	dao := pubdao.Private()

	mockHandlerCalled := false
	mockHandler := daokit.NewActionHandler("valid", func(_ interface{}) { mockHandlerCalled = true })
	dao.Core.Resources.Set(&daokit.Resource{
		Handler:   mockHandler,
		Condition: daocond.MembersThreshold(0.2, dao.Members.IsMember, dao.Members.MembersCount),
	})

	mockProposalRequest := daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action:      daokit.NewAction("valid", nil),
	}

	unknownProposalRequest := daokit.ProposalRequest{
		Title:       "My Proposal",
		Description: "My Proposal Description",
		Action:      daokit.NewAction("unknown", nil),
	}

	return &testingDAOContext{
		dao:                    pubdao,
		privdao:                dao,
		mockHandlerCalled:      &mockHandlerCalled,
		mockHandler:            mockHandler,
		mockProposalRequest:    mockProposalRequest,
		unknownProposalRequest: unknownProposalRequest,
	}
}

// Crossing wrappers for dao.X calls.

func (c *testingDAOContext) propose(_ int, rlm realm, req daokit.ProposalRequest) (id uint64) {
	func(cur realm) { id = c.dao.Propose(0, cur, req) }(cross(rlm))
	return
}

func (c *testingDAOContext) vote(_ int, rlm realm, id uint64, v daocond.Vote) {
	func(cur realm) { c.dao.Vote(0, cur, id, v) }(cross(rlm))
}

func (c *testingDAOContext) execute(_ int, rlm realm, id uint64) {
	func(cur realm) { c.dao.Execute(0, cur, id) }(cross(rlm))
}

func (c *testingDAOContext) instantExecute(_ int, rlm realm, req daokit.ProposalRequest) (id uint64) {
	func(cur realm) { id = c.dao.InstantExecute(0, cur, req) }(cross(rlm))
	return
}
