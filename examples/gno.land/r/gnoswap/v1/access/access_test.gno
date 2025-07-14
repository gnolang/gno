package access

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/rbac"
)

var (
	testAdminAddr  = testutils.TestAddress("test_admin")
	testRouterAddr = testutils.TestAddress("test_router")

	testRealm1Addr = testutils.TestAddress("test_realm1")
	testRealm2Addr = testutils.TestAddress("test_realm2")

	adminAddr, _ = GetAddress(ROLE_ADMIN)
	adminRealm   = std.NewUserRealm(adminAddr)
)

func TestInitialize(t *testing.T) {
	t.Run("success with valid config", func(t *testing.T) {
		cfg := newConfig()
		cfg.roles["test_admin"] = testAdminAddr
		cfg.roles["test_router"] = testRouterAddr

		err := initialize(cfg)
		uassert.NoError(t, err)

		actual := GetCurrentConfig()

		// retrieve the config informations
		adminAddr := actual.roles["test_admin"]
		routerAddr := actual.roles["test_router"]

		uassert.Equal(t, testAdminAddr, adminAddr)
		uassert.Equal(t, testRouterAddr, routerAddr)
	})

	t.Run("initialize with nil config", func(t *testing.T) {
		err := initialize(nil)
		uassert.Error(t, err)
	})
}

func TestUpdateRoleAddress(t *testing.T) {
	testing.SetRealm(adminRealm)

	t.Run("update success", func(t *testing.T) {
		cfg := newConfig()
		cfg.roles["another_test_admin"] = testAdminAddr

		func(cur realm) {
			err := initialize(cfg)
			uassert.NoError(t, err)

			err = UpdateRoleAddress(cur, "another_test_admin", testRealm1Addr)
			uassert.NoError(t, err)

			actual := GetCurrentConfig()
			adminAddr := actual.roles["another_test_admin"]
			uassert.Equal(t, testRealm1Addr, adminAddr)

			err = rbac.CheckPermission("another_test_admin", PERM_ACCESS, adminAddr)
			uassert.NoError(t, err)
		}(cross)
	})

	t.Run("update fails while uninitialized", func(t *testing.T) {
		func(cur realm) {
			testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/access"))

			currentConfig = nil
			err := UpdateRoleAddress(cur, "another_test_admin", testRealm1Addr)
			uassert.Error(t, err)
		}(cross)
	})

	t.Run("update fails while update non-existent role", func(t *testing.T) {
		func(cur realm) {
			testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/access"))

			cfg := newConfig()
			err := initialize(cfg)
			uassert.NoError(t, err)

			err = UpdateRoleAddress(cur, "non_existent_role", testRealm1Addr)
			uassert.Error(t, err)
		}(cross)
	})
}

func TestUserRealm(t *testing.T) {
	testAdminAddr2 := testutils.TestAddress("test_admin2")
	testGovAddr := testutils.TestAddress("test_governance")
	otherAddr := testutils.TestAddress("other")
	testOperatorAddr := testutils.TestAddress("test_operator")
	testValidatorAddr := testutils.TestAddress("test_validator")
	testUserAddr := testutils.TestAddress("test_user")

	cfg := newConfig()
	cfg.roles["test_admin2"] = testAdminAddr2
	cfg.roles["test_governance"] = testGovAddr
	cfg.roles["test_operator"] = testOperatorAddr
	cfg.roles["test_validator"] = testValidatorAddr

	err := initialize(cfg)
	uassert.NoError(t, err)

	t.Run("Role Permission Tests", func(t *testing.T) {
		tests := []struct {
			name      string
			role      string
			addr      std.Address
			shouldErr bool
		}{
			{
				name:      "test_admin2 address accessing test_admin2 role",
				role:      "test_admin2",
				addr:      testAdminAddr2,
				shouldErr: false,
			},
			{
				name:      "non-test_admin2 address accessing test_admin2 role",
				role:      "test_admin2",
				addr:      otherAddr,
				shouldErr: true,
			},
			{
				name:      "test_governance address accessing test_governance role",
				role:      "test_governance",
				addr:      testGovAddr,
				shouldErr: false,
			},
			{
				name:      "non-test_governance address accessing test_governance role",
				role:      "test_governance",
				addr:      otherAddr,
				shouldErr: true,
			},
			{
				name:      "empty address accessing admin role",
				role:      "test_admin2",
				addr:      std.Address(""),
				shouldErr: true,
			},
			{
				name:      "operator address accessing validator role",
				role:      "test_validator",
				addr:      testOperatorAddr,
				shouldErr: true,
			},
			{
				name:      "regular user address accessing operator role",
				role:      "test_operator",
				addr:      testUserAddr,
				shouldErr: true,
			},
			{
				name:      "accessing non-existent role",
				role:      "non_existent_role",
				addr:      testUserAddr,
				shouldErr: true,
			},
			{
				name:      "validator address accessing own role",
				role:      "test_validator",
				addr:      testValidatorAddr,
				shouldErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := rbac.CheckPermission(tt.role, PERM_ACCESS, tt.addr)
				if tt.shouldErr {
					uassert.Error(t, err)
				} else {
					uassert.NoError(t, err)
				}
			})
		}
	})
}

func TestMultiRoleValidation(t *testing.T) {
	testLiquidityAddr := testutils.TestAddress("test_liquidity")
	testStakerAddr := testutils.TestAddress("test_staker")
	testMultiRoleAddr := testutils.TestAddress("test_multi_role")

	cfg := newConfig()
	cfg.roles["test_liquidity"] = testLiquidityAddr
	cfg.roles["test_staker"] = testStakerAddr

	// add multiple roles to one address
	cfg.roles["test_liquidity_multi"] = testMultiRoleAddr
	cfg.roles["test_staker_multi"] = testMultiRoleAddr

	err := initialize(cfg)
	uassert.NoError(t, err)

	t.Run("Multiple Role Assignment Validation", func(t *testing.T) {
		tests := []struct {
			name      string
			addr      std.Address
			roles     []string
			perms     []string
			shouldErr bool
		}{
			{
				name:      "address with correct multiple roles",
				addr:      testMultiRoleAddr,
				roles:     []string{"test_liquidity_multi", "test_staker_multi"},
				perms:     []string{PERM_ACCESS, PERM_ACCESS},
				shouldErr: false,
			},
			{
				name:      "address with single role trying multiple access",
				addr:      testLiquidityAddr,
				roles:     []string{"test_liquidity", "test_staker"},
				perms:     []string{PERM_ACCESS, PERM_ACCESS},
				shouldErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var lastErr error
				for i, role := range tt.roles {
					err := rbac.CheckPermission(role, tt.perms[i], tt.addr)
					if err != nil {
						lastErr = err
						break
					}
				}

				if tt.shouldErr {
					uassert.Error(t, lastErr)
				} else {
					uassert.NoError(t, lastErr)
				}
			})
		}
	})
}

func TestCreateRole(t *testing.T) {
	testing.SetRealm(adminRealm)
	testNewRoleAddr := testutils.TestAddress("test_new_role")
	testAnotherAddr := testutils.TestAddress("test_another_role")

	t.Run("create new role success", func(t *testing.T) {
		cfg := newConfig()
		err := initialize(cfg)
		uassert.NoError(t, err)

		err = CreateRole(cross, "custom_role", testNewRoleAddr)
		uassert.NoError(t, err)

		// verify role was created correctly
		actual := GetCurrentConfig()
		roleAddr := actual.roles["custom_role"]
		uassert.Equal(t, testNewRoleAddr, roleAddr)

		// verify permission is set correctly
		err = rbac.CheckPermission("custom_role", PERM_ACCESS, testNewRoleAddr)
		uassert.NoError(t, err)

		// verify different address cannot access the role
		err = rbac.CheckPermission("custom_role", PERM_ACCESS, testAnotherAddr)
		uassert.Error(t, err)
	})

	t.Run("create role fails while uninitialized", func(t *testing.T) {
		currentConfig = nil
		err := CreateRole(cross, "failed_role", testNewRoleAddr)
		uassert.Error(t, err)
	})

	t.Run("create duplicate role fails", func(t *testing.T) {
		testing.SetRealm(adminRealm)

		cfg := newConfig()
		err := initialize(cfg)
		uassert.NoError(t, err)

		// create first role
		err = CreateRole(cross, "duplicate_role", testNewRoleAddr)
		uassert.NoError(t, err)

		// attempt to create same role again
		err = CreateRole(cross, "duplicate_role", testAnotherAddr)
		uassert.Error(t, err)
	})
}

func TestRoleExists(t *testing.T) {
	testRoleAddr := testutils.TestAddress("test_role_exists")

	t.Run("check existing role", func(t *testing.T) {
		cfg := newConfig()
		cfg.roles["existing_role"] = testRoleAddr
		err := initialize(cfg)
		uassert.NoError(t, err)

		exists := RoleExists("existing_role")
		uassert.Equal(t, true, exists)
	})

	t.Run("check non-existing role", func(t *testing.T) {
		cfg := newConfig()
		err := initialize(cfg)
		uassert.NoError(t, err)

		exists := RoleExists("non_existing_role")
		uassert.Equal(t, false, exists)
	})

	t.Run("check while uninitialized", func(t *testing.T) {
		currentConfig = nil
		exists := RoleExists("any_role")
		uassert.Equal(t, false, exists)
	})
}

func TestGetRoles(t *testing.T) {
	testRole1Addr := testutils.TestAddress("test_role1_get")
	testRole2Addr := testutils.TestAddress("test_role2_get")

	t.Run("get roles from initialized config", func(t *testing.T) {
		cfg := newConfig()
		cfg.roles["role1"] = testRole1Addr
		cfg.roles["role2"] = testRole2Addr
		err := initialize(cfg)
		uassert.NoError(t, err)

		roles := GetRoles()
		uassert.Equal(t, 2, len(roles))

		// check if both roles exist in the result
		foundRole1 := false
		foundRole2 := false
		for _, role := range roles {
			if role == "role1" {
				foundRole1 = true
			}
			if role == "role2" {
				foundRole2 = true
			}
		}
		uassert.Equal(t, true, foundRole1)
		uassert.Equal(t, true, foundRole2)
	})

	t.Run("get roles from empty config", func(t *testing.T) {
		cfg := newConfig()
		err := initialize(cfg)
		uassert.NoError(t, err)

		roles := GetRoles()
		uassert.Equal(t, 0, len(roles))
	})

	t.Run("get roles while uninitialized", func(t *testing.T) {
		currentConfig = nil
		roles := GetRoles()
		uassert.Equal(t, 0, len(roles))
	})
}

func TestGetAddress(t *testing.T) {
	testRoleFooAddr := testutils.TestAddress("role_foo")
	testRoleBarAddr := testutils.TestAddress("role_bar")

	t.Run("get address from initialized config", func(t *testing.T) {
		cfg := newConfig()
		cfg.roles["role_foo"] = testRoleFooAddr
		cfg.roles["role_bar"] = testRoleBarAddr
		err := initialize(cfg)
		uassert.NoError(t, err)

		addr, err := GetAddress("role_foo")
		uassert.NoError(t, err)
		uassert.Equal(t, testRoleFooAddr, addr)
	})

	t.Run("get address from uninitialized config", func(t *testing.T) {
		currentConfig = nil
		_, err := GetAddress("role_foo")
		uassert.Error(t, err)
	})
}

func TestSetConfig(t *testing.T) {
	testManagerAddr := testutils.TestAddress("test_manager")
	testAnalystAddr := testutils.TestAddress("test_analyst")

	t.Run("success with valid config", func(t *testing.T) {
		cfg := newConfig()
		cfg.roles["test_manager"] = testManagerAddr
		cfg.roles["test_analyst"] = testAnalystAddr

		err := setConfig(cfg)
		uassert.NoError(t, err)

		actual := GetCurrentConfig()
		uassert.Equal(t, testManagerAddr, actual.roles["test_manager"])
		uassert.Equal(t, testAnalystAddr, actual.roles["test_analyst"])
	})

	t.Run("fail with nil config", func(t *testing.T) {
		err := setConfig(nil)
		uassert.Error(t, err)
	})
}

func TestSetRole(t *testing.T) {
	testing.SetRealm(adminRealm)
	testNewRoleAddr := testutils.TestAddress("test_new_role")
	testAnotherAddr := testutils.TestAddress("test_another_role")

	testAuditorAddr := testutils.TestAddress("test_auditor")
	testReviewerAddr := testutils.TestAddress("test_reviewer")

	t.Run("success with new role", func(t *testing.T) {
		func(cur realm) {
			cfg := newConfig()
			err := initialize(cfg)
			uassert.NoError(t, err)

			err = SetRole(cur, "test_auditor", testAuditorAddr)
			uassert.NoError(t, err)

			actual := GetCurrentConfig()
			uassert.Equal(t, testAuditorAddr, actual.roles["test_auditor"])

			// verify permission is set correctly
			err = rbac.CheckPermission("test_auditor", PERM_ACCESS, testAuditorAddr)
			uassert.NoError(t, err)
		}(cross)
	})

	t.Run("set new role success", func(t *testing.T) {
		func(cur realm) {
			cfg := newConfig()
			err := initialize(cfg)
			uassert.NoError(t, err)

			err = SetRole(cur, "custom_role_2", testNewRoleAddr)
			uassert.NoError(t, err)

			// verify role was set correctly
			actual := GetCurrentConfig()
			roleAddr := actual.roles["custom_role_2"]
			uassert.Equal(t, testNewRoleAddr, roleAddr)

			// verify permission is set correctly
			err = rbac.CheckPermission("custom_role_2", PERM_ACCESS, testNewRoleAddr)
			uassert.NoError(t, err)

			// verify different address cannot access the role
			err = rbac.CheckPermission("custom_role_2", PERM_ACCESS, testAnotherAddr)
			uassert.Error(t, err)
		}(cross)
	})

	t.Run("success updating existing role", func(t *testing.T) {
		func(cur realm) {
			cfg := newConfig()
			cfg.roles["test_reviewer"] = testReviewerAddr
			err := initialize(cfg)
			uassert.NoError(t, err)

			newReviewerAddr := testutils.TestAddress("test_reviewer_new")
			err = SetRole(cur, "test_reviewer", newReviewerAddr)
			uassert.NoError(t, err)

			// verify address was updated
			actual := GetCurrentConfig()
			uassert.Equal(t, newReviewerAddr, actual.roles["test_reviewer"])

			// verify old address cannot access anymore
			err = rbac.CheckPermission("test_reviewer", PERM_ACCESS, testReviewerAddr)
			uassert.Error(t, err)

			// verify new address can access
			err = rbac.CheckPermission("test_reviewer", PERM_ACCESS, newReviewerAddr)
			uassert.NoError(t, err)
		}(cross)
	})

	t.Run("set role fails while uninitialized", func(t *testing.T) {
		func(cur realm) {
			testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/access"))

			currentConfig = nil
			err := SetRole(cur, "failed_role", testNewRoleAddr)
			uassert.Error(t, err)
		}(cross)
	})

	t.Run("fail with non-admin address", func(t *testing.T) {
		func(cur realm) {
			testing.SetRealm(std.NewUserRealm(testutils.TestAddress("non_admin")))
			err := SetRole(cur, "test_auditor", testAuditorAddr)
			uassert.Error(t, err)
		}(cross)
	})
}

func TestDeleteRole(t *testing.T) {
	testing.SetRealm(adminRealm)
	testRoleAddr := testutils.TestAddress("test_role_to_delete")

	t.Run("success delete existing role", func(t *testing.T) {
		func(cur realm) {
			cfg := newConfig()
			cfg.roles["test_role_to_delete"] = testRoleAddr
			err := initialize(cfg)
			uassert.NoError(t, err)

			// Verify role exists before deletion
			exists := RoleExists("test_role_to_delete")
			uassert.Equal(t, true, exists)

			// Delete the role
			err = DeleteRole(cur, "test_role_to_delete")
			uassert.NoError(t, err)

			// Verify role no longer exists
			exists = RoleExists("test_role_to_delete")
			uassert.Equal(t, false, exists)

			// Verify permission is removed
			err = rbac.CheckPermission("test_role_to_delete", PERM_ACCESS, testRoleAddr)
			uassert.Error(t, err)
		}(cross)
	})

	t.Run("fail to delete non-existent role", func(t *testing.T) {
		func(cur realm) {
			cfg := newConfig()
			err := initialize(cfg)
			uassert.NoError(t, err)

			err = DeleteRole(cur, "non_existent_role")
			uassert.Error(t, err)
		}(cross)
	})

	t.Run("fail to delete admin role", func(t *testing.T) {
		func(cur realm) {
			err := DeleteRole(cur, ROLE_ADMIN)
			uassert.Error(t, err)
		}(cross)
	})

	t.Run("fail while uninitialized", func(t *testing.T) {
		currentConfig = nil
		func(cur realm) {
			err := DeleteRole(cur, "any_role")
			uassert.Error(t, err)
		}(cross)
	})

	t.Run("fail with empty role name", func(t *testing.T) {
		func(cur realm) {
			cfg := newConfig()
			err := initialize(cfg)
			uassert.NoError(t, err)

			err = DeleteRole(cur, "")
			uassert.Error(t, err)
		}(cross)
	})
}

// The below tests are affected by the state of global variable.
// so, need to run separately for now.
// Once the test environment of gno is improved, the tests should be modified to call the setupRoles function in each test.

func TestInitializeDefaultRoles(t *testing.T) {
	t.Skip()
	t.Run("success initialize default roles", func(t *testing.T) {
		actual := GetCurrentConfig()

		expectedRoles := map[string]std.Address{
			ROLE_ADMIN:      consts.ADMIN,
			ROLE_GOVERNANCE: consts.GOV_GOVERNANCE_ADDR,
			ROLE_GOV_STAKER: consts.GOV_STAKER_ADDR,
			ROLE_ROUTER:     consts.ROUTER_ADDR,
			ROLE_POOL:       consts.POOL_ADDR,
			ROLE_POSITION:   consts.POSITION_ADDR,
			ROLE_STAKER:     consts.STAKER_ADDR,
			ROLE_LAUNCHPAD:  consts.LAUNCHPAD_ADDR,
			ROLE_EMISSION:   consts.EMISSION_ADDR,
			ROLE_GOV_XGNS:   consts.GOV_XGNS_ADDR,
		}

		for role, expectedAddr := range expectedRoles {
			actualAddr, exists := actual.roles[role]
			uassert.Equal(t, true, exists, ufmt.Sprintf("Role %s should exist", role))
			uassert.Equal(t, expectedAddr, actualAddr, ufmt.Sprintf("Role %s should have address %s, got %s", role, expectedAddr, actualAddr))
		}
	})
}

func setupRoles(t *testing.T) {
	_, err := initializeDefaultRoles()
	uassert.NoError(t, err)
}
