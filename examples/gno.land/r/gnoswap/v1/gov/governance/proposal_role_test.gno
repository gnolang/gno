package governance

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

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
