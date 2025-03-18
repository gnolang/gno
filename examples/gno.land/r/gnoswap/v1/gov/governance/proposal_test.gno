package governance

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"
)

var (
	proposalAddr = testutils.TestAddress("proposal")
	toAddr       = testutils.TestAddress("to")

	voter1 = testutils.TestAddress("voter1")
	voter2 = testutils.TestAddress("voter2")

	insufficientProposer = testutils.TestAddress("insufficient")
)

func resetGlobalStateProposal(t *testing.T) {
	t.Helper()
	proposalId = 0
	proposals = avl.NewTree()
	proposalIdListByProposer = avl.NewTree()
}

func mockCheckEnoughXGnsToPropose(proposer std.Address) (bool, uint64, uint64) {
	if proposer == insufficientProposer {
		return false, 500, 1000
	}
	return true, 1000, 1000
}

func mockGetPossibleVotingAddressWithWeight(t *testing.T, timestamp uint64) (uint64, map[std.Address]uint64) {
	t.Helper()
	weights := make(map[std.Address]uint64)
	weights[voter1] = 100
	weights[voter2] = 200
	return 300, weights
}

func TestProposeText(t *testing.T) {
	// override for test
	origProposalCreationThreshold := config.ProposalCreationThreshold
	config.ProposalCreationThreshold = 0

	resetGlobalStateProposal(t)

	tests := []struct {
		name        string
		proposer    std.Address
		title       string
		description string
		expectError bool
	}{
		{
			name:        "Valid text proposal",
			proposer:    proposalAddr,
			title:       "Test Proposal",
			description: "This is a test proposal",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pid uint64
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = r.(error)
					}
				}()
				pid = ProposeText(tt.title, tt.description)
			}()

			uassert.NoError(t, err)

			prop, exists := proposals.Get(formatUint(pid))
			uassert.True(t, exists)

			proposal := prop.(ProposalInfo)

			uassert.Equal(t, proposal.Title, tt.title)
			uassert.Equal(t, proposal.Description, tt.description)
			uassert.Equal(t, proposal.ProposalType.String(), Text.String())
			uassert.True(t, proposal.State.Created)
			uassert.True(t, proposal.State.Upcoming)
		})
	}

	config.ProposalCreationThreshold = origProposalCreationThreshold
}

func TestProposeCommunityPoolSpend(t *testing.T) {
	t.Skip("")
	// checkEnoughXGnsToPropose = mockCheckEnoughXGnsToPropose

	resetGlobalStateProposal(t)

	tests := []struct {
		name        string
		proposer    std.Address
		title       string
		description string
		to          std.Address
		tokenPath   string
		amount      uint64
		expectError bool
	}{
		{
			name:        "Valid community pool spend proposal",
			proposer:    proposalAddr,
			title:       "Community Spend",
			description: "Fund community initiative",
			to:          toAddr,
			tokenPath:   "token/path",
			amount:      1000,
			expectError: false,
		},
		{
			name:        "Insufficient balance for proposal",
			proposer:    insufficientProposer,
			title:       "Invalid Spend",
			description: "Should fail",
			to:          toAddr,
			tokenPath:   "token/path",
			amount:      1000,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pid uint64
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = r.(error)
					}
				}()
				pid = ProposeCommunityPoolSpend(
					tt.title, tt.description, tt.to, tt.tokenPath, tt.amount)
			}()

			uassert.NoError(t, err)

			prop, exists := proposals.Get(formatUint(pid))
			uassert.True(t, exists)

			proposal := prop.(ProposalInfo)

			uassert.Equal(t, proposal.ProposalType.String(), CommunityPoolSpend.String())
			uassert.Equal(t, proposal.CommunityPoolSpend.To, tt.to)
			uassert.Equal(t, proposal.CommunityPoolSpend.Amount, tt.amount)
		})
	}
}

func TestUpdateProposalsState(t *testing.T) {
	t.Skip("")
	baseTime := uint64(1000)
	newConfig := Config{
		VotingStartDelay: 50,
		VotingPeriod:     100,
		ExecutionDelay:   50,
		ExecutionWindow:  100,
	}
	setConfigVersion(1, newConfig)

	tests := []struct {
		name          string
		currentTime   uint64
		setupProposal func(uint64) ProposalInfo
		validate      func(*testing.T, ProposalInfo)
	}{
		{
			name:        "Should reject proposal when voting ends with insufficient votes",
			currentTime: baseTime + 200,
			setupProposal: func(now uint64) ProposalInfo {
				return ProposalInfo{
					ConfigVersion: 1,
					Yea:           u256.NewUint(100),
					Nay:           u256.NewUint(200),
					QuorumAmount:  300,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime,
						Active:    true,
					},
				}
			},
			validate: func(t *testing.T, proposal ProposalInfo) {
				uassert.True(t, proposal.State.Rejected)
				uassert.False(t, proposal.State.Active)
				uassert.False(t, proposal.State.Upcoming)
				uassert.NotEqual(t, proposal.State.RejectedAt, uint64(0))
			},
		},
		{
			name:        "Should pass proposal when voting ends with sufficient votes",
			currentTime: baseTime + 200,
			setupProposal: func(now uint64) ProposalInfo {
				return ProposalInfo{
					ConfigVersion: 1,
					Yea:           u256.NewUint(400),
					Nay:           u256.NewUint(200),
					QuorumAmount:  300,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime,
						Active:    true,
					},
				}
			},
			validate: func(t *testing.T, proposal ProposalInfo) {
				uassert.True(t, proposal.State.Passed)
				uassert.False(t, proposal.State.Active)
				uassert.NotEqual(t, proposal.State.PassedAt, uint64(0))
			},
		},
		{
			name:        "Should expire non-text proposal when execution window ends",
			currentTime: baseTime + 400,
			setupProposal: func(now uint64) ProposalInfo {
				return ProposalInfo{
					ConfigVersion: 1,
					ProposalType:  ParameterChange,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime,
						Passed:    true,
						PassedAt:  baseTime + 200,
					},
				}
			},
			validate: func(t *testing.T, proposal ProposalInfo) {
				uassert.True(t, proposal.State.Expired)
				uassert.NotEqual(t, proposal.State.ExpiredAt, uint64(0))
			},
		},
		{
			name:        "Should not update canceled proposal",
			currentTime: baseTime + 60,
			setupProposal: func(now uint64) ProposalInfo {
				return ProposalInfo{
					ConfigVersion: 1,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime,
						Canceled:  true,
						Upcoming:  true,
					},
				}
			},
			validate: func(t *testing.T, proposal ProposalInfo) {
				uassert.True(t, proposal.State.Canceled)
				uassert.True(t, proposal.State.Upcoming)
			},
		},
		{
			name:        "Should not expire text proposal",
			currentTime: baseTime + 400,
			setupProposal: func(now uint64) ProposalInfo {
				return ProposalInfo{
					ConfigVersion: 1,
					ProposalType:  Text,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime,
						Passed:    true,
						PassedAt:  baseTime + 200,
					},
				}
			},
			validate: func(t *testing.T, proposal ProposalInfo) {
				uassert.False(t, proposal.State.Expired)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proposals = avl.NewTree()

			proposal := tt.setupProposal(tt.currentTime)
			proposals.Set(formatUint(uint32(1)), proposal)

			updateProposalsState()

			updatedProposal, exists := proposals.Get(formatUint(uint32(1)))
			uassert.True(t, exists)
			tt.validate(t, updatedProposal.(ProposalInfo))
		})
	}
}

func TestProposeParameterChange(t *testing.T) {
	t.Skip("")
	resetGlobalStateProposal(t)

	tests := []struct {
		name          string
		proposer      std.Address
		title         string
		description   string
		numToExecute  uint64
		executions    string
		expectError   bool
		errorContains string
	}{
		{
			name:         "Valid parameter change proposal",
			proposer:     proposalAddr,
			title:        "Change Voting Period",
			description:  "Update voting period to 14 days",
			numToExecute: 2,
			executions:   "gno.land/r/gnoswap/v1/gns*EXE*SetAvgBlockTimeInMs*EXE*123*GOV*gno.land/r/gnoswap/v1/community_pool*EXE*TransferToken*EXE*gno.land/r/gnoswap/v1/gns,g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d,905",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalStateProposal(t)

			var pid uint64
			var err error
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Unexpected error: %v", r)
				}
			}()
			pid = ProposeParameterChange(
				tt.title,
				tt.description,
				tt.numToExecute,
				tt.executions,
			)

			uassert.NoError(t, err)

			prop, exists := proposals.Get(formatUint(pid))
			uassert.True(t, exists)

			proposal := prop.(ProposalInfo)

			uassert.Equal(t, proposal.Title, tt.title)
			uassert.Equal(t, proposal.Description, tt.description)
			uassert.Equal(t, proposal.ProposalType.String(), ParameterChange.String())

			uassert.Equal(t, proposal.Execution.Num, tt.numToExecute)
			uassert.Equal(t, len(proposal.Execution.Msgs), int(tt.numToExecute))

			uassert.True(t, proposal.State.Created)
			uassert.True(t, proposal.State.Upcoming)
			uassert.False(t, proposal.State.Active)
			uassert.True(t, proposal.Yea.IsZero())
			uassert.True(t, proposal.Nay.IsZero())
		})
	}
}

func TestIntegratedProposeParameterChange(t *testing.T) {
	t.Skip("The proposal needs to be modified to ensure it has sufficient xGNS.")
	resetGlobalStateProposal(t)

	tests := []struct {
		name           string
		proposer       std.Address
		title          string
		description    string
		numToExecute   uint64
		executions     string
		expectError    bool
		errorContains  string
		isRoleChange   bool
		roleName       string
		permissionName string
		action         RoleChangeAction
	}{
		{
			name:         "Valid parameter change proposal",
			proposer:     proposalAddr,
			title:        "Change Voting Period",
			description:  "Update voting period to 14 days",
			numToExecute: 2,
			executions:   "gno.land/r/gnoswap/v1/gns*EXE*SetAvgBlockTimeInMs*EXE*123*GOV*gno.land/r/gnoswap/v1/community_pool*EXE*TransferToken*EXE*gno.land/r/gnoswap/v1/gns,g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d,905",
			expectError:  false,
			isRoleChange: false,
		},
		{
			name:           "Valid role change proposal",
			proposer:       proposalAddr,
			title:          "Add Admin Role",
			description:    "Create new admin role with single address permission",
			numToExecute:   1,
			executions:     "ROLE_CHANGE:admin:can_manage:AddRole:SingleAddressChecker:g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d::0",
			expectError:    false,
			isRoleChange:   true,
			roleName:       "admin",
			permissionName: "can_manage",
			action:         AddRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalStateProposal(t)

			var pid uint64
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Unexpected error: %v", r)
					}
				}()
				pid = ProposeParameterChange(
					tt.title,
					tt.description,
					tt.numToExecute,
					tt.executions,
				)
			}()

			if tt.expectError {
				uassert.Error(t, err)
				return
			}
			uassert.NoError(t, err)

			prop, exists := proposals.Get(formatUint(pid))
			uassert.True(t, exists)

			proposal := prop.(ProposalInfo)

			uassert.Equal(t, proposal.Title, tt.title)
			uassert.Equal(t, proposal.Description, tt.description)

			if tt.isRoleChange {
				uassert.Equal(t, proposal.ProposalType.String(), RoleChange.String())
				uassert.Equal(t, proposal.RoleChange.RoleName, tt.roleName)
				uassert.Equal(t, proposal.RoleChange.PermissionName, tt.permissionName)
				if tt.action != proposal.RoleChange.Action {
					t.Errorf("Expected action: %v, got: %v", tt.action, proposal.RoleChange.Action)
				}
			} else {
				uassert.Equal(t, proposal.ProposalType.String(), ParameterChange.String())
				uassert.Equal(t, proposal.Execution.Num, tt.numToExecute)
			}

			uassert.True(t, proposal.State.Created)
			uassert.True(t, proposal.State.Upcoming)
			uassert.False(t, proposal.State.Active)
			uassert.True(t, proposal.Yea.IsZero())
			uassert.True(t, proposal.Nay.IsZero())
		})
	}
}
