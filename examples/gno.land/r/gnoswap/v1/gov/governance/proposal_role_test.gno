package governance

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestProposeRoleChange(t *testing.T) {
	// This test works properly when the xGNS balance check is commented out.
	t.Skip("The proposal needs to be modified to ensure it has sufficient xGNS.")
	resetGlobalStateProposal(t)

	testAddr1 := testutils.TestAddress("test1")
	testAddr2 := testutils.TestAddress("test2")

	tests := []struct {
		name           string
		proposer       std.Address
		title          string
		description    string
		roleName       string
		permissionName string
		action         RoleChangeAction
		checkerType    CheckerType
		addresses      []std.Address
		tokenPath      string
		minAmount      uint64
		expectError    bool
	}{
		{
			name:           "Valid role addition with single address checker",
			proposer:       proposalAddr,
			title:          "Add Admin Role",
			description:    "Create new admin role with single address permission",
			roleName:       "admin",
			permissionName: "can_manage",
			action:         AddRole,
			checkerType:    SingleAddressChecker,
			addresses:      []std.Address{testAddr1},
			expectError:    false,
		},
		{
			name:           "Valid role addition with multi address checker",
			proposer:       proposalAddr,
			title:          "Add Moderator Role",
			description:    "Create new moderator role with multiple addresses",
			roleName:       "moderator",
			permissionName: "can_moderate",
			action:         AddRole,
			checkerType:    MultiAddressChecker,
			addresses:      []std.Address{testAddr1, testAddr2},
			expectError:    false,
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
						t.Errorf("proposeRoleChange panicked: %v", r)
					}
				}()
				pid = proposeRoleChange(
					tt.title,
					tt.description,
					tt.roleName,
					tt.permissionName,
					tt.action,
					tt.checkerType,
					tt.addresses,
					tt.tokenPath,
					tt.minAmount,
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
			uassert.Equal(t, proposal.ProposalType.String(), RoleChange.String())
			uassert.Equal(t, proposal.RoleChange.RoleName, tt.roleName)
			uassert.Equal(t, proposal.RoleChange.PermissionName, tt.permissionName)
			uassert.Equal(t, proposal.RoleChange.Action.String(), tt.action.String())
			uassert.Equal(t, proposal.RoleChange.CheckerType.String(), tt.checkerType.String())

			if tt.checkerType == SingleAddressChecker || tt.checkerType == MultiAddressChecker {
				uassert.Equal(t, len(proposal.RoleChange.Addresses), len(tt.addresses))
				for i, addr := range tt.addresses {
					uassert.Equal(t, proposal.RoleChange.Addresses[i], addr)
				}
			}

			uassert.True(t, proposal.State.Created)
			uassert.True(t, proposal.State.Upcoming)
			uassert.False(t, proposal.State.Active)
			uassert.True(t, proposal.Yea.IsZero())
			uassert.True(t, proposal.Nay.IsZero())
		})
	}
}

func TestProposeRoleChangeWithStringParams(t *testing.T) {
	t.Skip("The proposal needs to be modified to ensure it has sufficient xGNS.")
	resetGlobalStateProposal(t)

	testAddr1 := testutils.TestAddress("test1")
	testAddr2 := testutils.TestAddress("test2")

	tests := []struct {
		name           string
		title          string
		description    string
		roleName       string
		permissionName string
		action         string
		checkerType    string
		addresses      string
		tokenPath      string
		minAmount      uint64
		expectError    bool
		expectedAddrs  []std.Address
	}{
		{
			name:           "Add role with single address checker",
			title:          "Add Admin Role",
			description:    "Create new admin role with single address permission",
			roleName:       "admin",
			permissionName: "can_manage",
			action:         "AddRole",
			checkerType:    "SingleAddressChecker",
			addresses:      testAddr1.String(),
			tokenPath:      "",
			minAmount:      0,
			expectError:    false,
			expectedAddrs:  []std.Address{testAddr1},
		},
		{
			name:           "Add role with multiple address checker",
			title:          "Add Moderator Role",
			description:    "Create new moderator role with multiple addresses",
			roleName:       "moderator",
			permissionName: "can_moderate",
			action:         "AddRole",
			checkerType:    "MultiAddressChecker",
			addresses:      testAddr1.String() + ":" + testAddr2.String(),
			tokenPath:      "",
			minAmount:      0,
			expectError:    false,
			expectedAddrs:  []std.Address{testAddr1, testAddr2},
		},
		{
			name:           "Invalid action type",
			title:          "Invalid Action",
			description:    "Test with invalid action type",
			roleName:       "test",
			permissionName: "test_perm",
			action:         "InvalidAction",
			checkerType:    "AlwaysAllowChecker",
			addresses:      "",
			tokenPath:      "",
			minAmount:      0,
			expectError:    true,
			expectedAddrs:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalStateProposal(t)

			var pid uint64
			didPanic := false

			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic = true
					}
				}()
				pid = ProposeRoleChange(
					tt.title,
					tt.description,
					tt.roleName,
					tt.permissionName,
					tt.action,
					tt.checkerType,
					tt.addresses,
					tt.tokenPath,
					tt.minAmount,
				)
			}()

			if tt.expectError {
				uassert.True(t, didPanic)
				return
			}

			uassert.False(t, didPanic)

			prop, exists := proposals.Get(formatUint(pid))
			uassert.True(t, exists)

			proposal := prop.(ProposalInfo)

			uassert.Equal(t, tt.title, proposal.Title)
			uassert.Equal(t, tt.description, proposal.Description)
			uassert.Equal(t, tt.roleName, proposal.RoleChange.RoleName)
			uassert.Equal(t, tt.permissionName, proposal.RoleChange.PermissionName)

			if tt.expectedAddrs != nil {
				uassert.Equal(t, len(tt.expectedAddrs), len(proposal.RoleChange.Addresses))
				for i, addr := range tt.expectedAddrs {
					uassert.Equal(t, addr, proposal.RoleChange.Addresses[i])
				}
			}
		})
	}
}

func TestGetPermissionChecker(t *testing.T) {
	testAddr1 := testutils.TestAddress("test1")
	testAddr2 := testutils.TestAddress("test2")
	unauthorizedAddr := testutils.TestAddress("unauthorized")

	tests := []struct {
		name        string
		info        RoleChangeInfo
		testAddr    std.Address
		expectError bool
	}{
		{
			name: "SingleAddressChecker - authorized",
			info: RoleChangeInfo{
				CheckerType: SingleAddressChecker,
				Addresses:   []std.Address{testAddr1},
			},
			testAddr:    testAddr1,
			expectError: false,
		},
		{
			name: "SingleAddressChecker - unauthorized",
			info: RoleChangeInfo{
				CheckerType: SingleAddressChecker,
				Addresses:   []std.Address{testAddr1},
			},
			testAddr:    unauthorizedAddr,
			expectError: true,
		},
		{
			name: "MultiAddressChecker - authorized",
			info: RoleChangeInfo{
				CheckerType: MultiAddressChecker,
				Addresses:   []std.Address{testAddr1, testAddr2},
			},
			testAddr:    testAddr2,
			expectError: false,
		},
		{
			name: "MultiAddressChecker - unauthorized",
			info: RoleChangeInfo{
				CheckerType: MultiAddressChecker,
				Addresses:   []std.Address{testAddr1, testAddr2},
			},
			testAddr:    unauthorizedAddr,
			expectError: true,
		},
		{
			name: "AlwaysAllowChecker",
			info: RoleChangeInfo{
				CheckerType: AlwaysAllowChecker,
			},
			testAddr:    unauthorizedAddr,
			expectError: false,
		},
		{
			name: "AlwaysDenyChecker",
			info: RoleChangeInfo{
				CheckerType: AlwaysDenyChecker,
			},
			testAddr:    testAddr1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := getPermissionChecker(tt.info)
			err := checker(tt.testAddr)

			if tt.expectError {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}
