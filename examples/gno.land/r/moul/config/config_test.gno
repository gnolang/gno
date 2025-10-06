package config

import (
	"chain"
	"chain/runtime"
	"errors"
	"testing"

	"gno.land/p/moul/authz"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
)

var (
	originAddr     = testutils.TestAddress("origin")
	manager1Addr   = testutils.TestAddress("manager1")
	manager2Addr   = testutils.TestAddress("manager2")
	nonManagerAddr = testutils.TestAddress("nonManager")
)

// Helper to reset the Authorizer for each test, simulating initialization.
func setupTest(cur realm, t *testing.T) {
	t.Helper()

	// Set the initial caller context
	testing.SetRealm(testing.NewUserRealm(originAddr))
	// Initialize the Authorizer with the originAddr as the sole member,
	// simulating the state after NewWithOrigin() in a real deployment.
	Authorizer = authz.NewWithAuthority(authz.NewMemberAuthority(originAddr))
	// Ensure the origin address is the initial manager
	uassert.True(t, HasManager(cur, originAddr), "origin should be the initial manager")
}

func TestAddManager(cur realm, t *testing.T) {
	setupTest(cur, t)

	// Origin adds manager1 - Should succeed
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err := AddManager(cross, manager1Addr)
	uassert.NoError(t, err, "origin adding manager1 should succeed")
	uassert.True(t, HasManager(cur, manager1Addr), "manager1 should now be a manager")

	// Non-manager tries to add manager2 - Should fail
	testing.SetRealm(testing.NewUserRealm(nonManagerAddr))
	err = AddManager(cross, manager2Addr)
	uassert.Error(t, err, "non-manager adding manager2 should fail")
	uassert.False(t, HasManager(cur, manager2Addr), "manager2 should not have been added")

	// Manager1 adds manager2 - Should succeed
	testing.SetRealm(testing.NewUserRealm(manager1Addr))
	err = AddManager(cross, manager2Addr)
	uassert.NoError(t, err, "manager1 adding manager2 should succeed")
	uassert.True(t, HasManager(cur, manager2Addr), "manager2 should now be a manager")

	// Transfer authority away from MemberAuthority
	testing.SetRealm(testing.NewUserRealm(originAddr)) // Origin transfers
	err = TransferManagement(cross, authz.NewAutoAcceptAuthority())
	uassert.NoError(t, err, "transferring authority should succeed")

	// Try adding after transfer - Should fail (wrong authority type)
	testing.SetRealm(testing.NewUserRealm(manager1Addr))
	err = AddManager(cross, nonManagerAddr) // Try adding someone new
	uassert.ErrorContains(t, err, "current authority is not a MemberAuthority", "adding manager should fail after transfer")
}

func TestRemoveManager(cur realm, t *testing.T) {
	setupTest(cur, t)

	// Add manager1 first
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err := AddManager(cross, manager1Addr)
	uassert.NoError(t, err, "setup: failed to add manager1")
	uassert.True(t, HasManager(cur, manager1Addr), "setup: manager1 should be added")

	// Non-manager tries to remove manager1 - Should fail
	testing.SetRealm(testing.NewUserRealm(nonManagerAddr))
	err = RemoveManager(cross, manager1Addr)
	uassert.Error(t, err, "non-manager removing manager1 should fail")
	uassert.True(t, HasManager(cur, manager1Addr), "manager1 should still be a manager")

	// Origin removes manager1 - Should succeed
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err = RemoveManager(cross, manager1Addr)
	uassert.NoError(t, err, "origin removing manager1 should succeed")
	uassert.False(t, HasManager(cur, manager1Addr), "manager1 should now be removed")

	// Add manager1 again for next test case
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err = AddManager(cross, manager1Addr)
	uassert.NoError(t, err, "setup: failed to re-add manager1")

	// Transfer authority
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err = TransferManagement(cross, authz.NewAutoAcceptAuthority())
	uassert.NoError(t, err, "transferring authority should succeed")

	// Try removing after transfer - Should fail (wrong authority type)
	testing.SetRealm(testing.NewUserRealm(originAddr)) // Use origin, doesn't matter which user now
	err = RemoveManager(cross, manager1Addr)
	uassert.ErrorContains(t, err, "current authority is not a MemberAuthority", "removing manager should fail after transfer")
}

func TestListManagers(cur realm, t *testing.T) {
	setupTest(cur, t)
	initialList := ListManagers(cross)
	assertAddrSliceEqual(t, []address{originAddr}, initialList)
	// Add manager1 and manager2
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err := AddManager(cross, manager1Addr)
	uassert.NoError(t, err)
	err = AddManager(cross, manager2Addr)
	uassert.NoError(t, err)

	// List should contain origin, manager1, manager2
	list1 := ListManagers(cross)
	expected1 := []address{manager2Addr, manager1Addr, originAddr}
	assertAddrSliceEqual(t, expected1, list1)

	// Remove manager1
	testing.SetRealm(testing.NewUserRealm(originAddr)) // Can be origin or manager2
	err = RemoveManager(cross, manager1Addr)
	uassert.NoError(t, err)

	// List should contain origin, manager2
	list2 := ListManagers(cross)
	expected2 := []address{manager2Addr, originAddr}
	assertAddrSliceEqual(t, expected2, list2)

	// Transfer authority
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err = TransferManagement(cross, authz.NewAutoAcceptAuthority())
	uassert.NoError(t, err)

	// List should be empty after transfer
	list3 := ListManagers(cross)
	uassert.True(t, len(list3) == 0, "manager list should be empty after transfer")
}

func TestHasManager(cur realm, t *testing.T) {
	setupTest(cur, t)

	// Initially, only origin is manager
	uassert.True(t, HasManager(cross, originAddr), "origin should initially be a manager")
	uassert.False(t, HasManager(cross, manager1Addr), "manager1 should not initially be a manager")
	uassert.False(t, HasManager(cross, nonManagerAddr), "nonManager should not initially be a manager")

	// Add manager1
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err := AddManager(cross, manager1Addr)
	uassert.NoError(t, err)

	// Check again
	uassert.True(t, HasManager(cross, originAddr), "origin should still be a manager")
	uassert.True(t, HasManager(cross, manager1Addr), "manager1 should now be a manager")
	uassert.False(t, HasManager(cross, nonManagerAddr), "nonManager should still not be a manager")

	// Transfer authority
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err = TransferManagement(cross, authz.NewAutoAcceptAuthority())
	uassert.NoError(t, err)

	// After transfer, HasManager should always return false for MemberAuthority checks
	uassert.False(t, HasManager(cross, originAddr), "HasManager should be false after transfer")
	uassert.False(t, HasManager(cross, manager1Addr), "HasManager should be false after transfer")
	uassert.False(t, HasManager(cross, nonManagerAddr), "HasManager should be false after transfer")
}

func TestTransferManagement(cur realm, t *testing.T) {
	setupTest(cur, t)

	// Add manager1
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err := AddManager(cross, manager1Addr)
	uassert.NoError(t, err)

	// Create a new authority (MemberAuthority with manager2)
	newAuthority := authz.NewMemberAuthority(manager2Addr)

	// Non-manager tries to transfer - Should fail
	testing.SetRealm(testing.NewUserRealm(nonManagerAddr))
	err = TransferManagement(cross, newAuthority)
	uassert.Error(t, err, "non-manager transfer should fail")
	_, isMemberAuth := Authorizer.Authority().(*authz.MemberAuthority)
	uassert.True(t, isMemberAuth, "authority should still be MemberAuthority") // Verify it didn't change

	// Manager1 tries to transfer - Should succeed
	testing.SetRealm(testing.NewUserRealm(manager1Addr))
	err = TransferManagement(cross, newAuthority)
	uassert.NoError(t, err, "manager1 transfer should succeed")

	// Verify current authority is the new one
	currentAuth := Authorizer.Authority()
	uassert.True(t, currentAuth == newAuthority, "current authority should be the new one")

	// Verify origin is no longer a manager under the *new* authority
	testing.SetRealm(testing.NewUserRealm(manager2Addr)) // Need new manager to check
	uassert.False(t, HasManager(cross, originAddr), "origin should not be manager under new authority")
	uassert.False(t, HasManager(cross, manager1Addr), "manager1 should not be manager under new authority")
	uassert.True(t, HasManager(cross, manager2Addr), "manager2 should be manager under new authority")

	// Try adding a manager using the old origin - Should fail
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err = AddManager(cross, nonManagerAddr)
	uassert.Error(t, err, "origin should not be able to add manager after transfer")

	// Try adding a manager using the new manager (manager2) - Should succeed
	testing.SetRealm(testing.NewUserRealm(manager2Addr))
	err = AddManager(cross, nonManagerAddr)
	uassert.NoError(t, err, "new manager (manager2) should be able to add managers")
	uassert.True(t, HasManager(cross, nonManagerAddr), "nonManager should be added by manager2")

	// Try transferring to nil - Should fail
	testing.SetRealm(testing.NewUserRealm(manager2Addr))
	err = TransferManagement(cross, nil)
	uassert.ErrorContains(t, err, "new authority cannot be nil", "transferring to nil should fail")
}

func TestTransferToContractAuthority(cur realm, t *testing.T) {
	setupTest(cur, t) // Origin is the initial manager

	contractPath := "gno.land/r/testcontract"
	contractRealm := testing.NewCodeRealm(contractPath) // Simulate contract realm

	// Define a simple contract authority handler
	handlerExecuted := false // Track if the handler itself gets called
	contractAuth := authz.NewContractAuthority(contractPath, func(title string, action authz.PrivilegedAction) error {
		// Simulate contract checking the caller *before* executing
		caller := runtime.CurrentRealm().Address()
		expectedContractAddr := chain.PackageAddress(contractPath)
		if caller != expectedContractAddr {
			// Fail before marking executed or running action
			// Note: In a real scenario, this handler might just ignore the call
			// if the caller isn't right, rather than returning an error,
			// depending on the desired contract logic. Returning an error
			// here helps the test verify the handler wasn't improperly called.
			return errors.New("handler: caller is not the contract")
		}

		// Only mark executed and run action if caller is correct
		handlerExecuted = true
		return action()
	})

	// Origin transfers management to the contract authority
	testing.SetRealm(testing.NewUserRealm(originAddr))
	err := TransferManagement(cross, contractAuth)
	uassert.NoError(t, err, "transfer to contract authority failed")
	uassert.True(t, Authorizer.Authority() == contractAuth, "authority should now be the contract authority")

	// Now, actions like AddManager/RemoveManager should fail because the current
	// authority is no longer a MemberAuthority. The contract would need its own
	// logic executed via Authorizer.DoByCurrent() to manage members if desired.

	// Try adding a manager (will check authority type) - Should fail
	testing.SetRealm(testing.NewUserRealm(originAddr)) // Caller doesn't matter for this check
	err = AddManager(cross, manager1Addr)
	uassert.ErrorContains(t, err, "current authority is not a MemberAuthority", "AddManager should fail with ContractAuthority")

	// Simulate an action authorized *by the contract* using Authorizer.Do
	var contractActionExecuted bool
	handlerExecuted = false         // Reset tracker
	testing.SetRealm(contractRealm) // Call must originate from the contract now
	err = Authorizer.DoByCurrent("some_contract_action", func() error {
		contractActionExecuted = true
		// Imagine contract logic here
		return nil
	})
	uassert.NoError(t, err, "contract action via Authorizer.Do failed")
	uassert.True(t, handlerExecuted, "handler should have been executed by contract call") // Verify handler ran
	uassert.True(t, contractActionExecuted, "contract action should have been executed")

	// Simulate an action from a user - Should fail before handler is called
	var userActionExecuted bool
	handlerExecuted = false // Reset tracker
	testing.SetRealm(testing.NewUserRealm(nonManagerAddr))
	err = Authorizer.DoByCurrent("some_user_action", func() error {
		userActionExecuted = true
		return nil
	})
	// The ContractAuthority.Authorize method should return an error
	// because the handler now returns an error if the caller isn't the contract.
	uassert.Error(t, err, "user action via Authorizer.Do should fail when contract is authority")
	uassert.ErrorContains(t, err, "handler: caller is not the contract", "error should originate from handler check") // Check specific error
	uassert.False(t, handlerExecuted, "handler should NOT have been executed by user call")                           // Verify handler didn't run past the check
	uassert.False(t, userActionExecuted, "user action should not have been executed")
}

// Helper to check if a slice contains a specific address
func containsAddr(list []address, addr address) bool {
	for _, item := range list {
		if item == addr {
			return true
		}
	}
	return false
}

func assertAddrSliceEqual(t *testing.T, expected, actual []address) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("expected slice length %d, got %d. Expected: %v, Got: %v", len(expected), len(actual), expected, actual)
	}

	for i := range expected {
		if expected[i] != actual[i] {
			t.Fatalf("slices differ at index %d. Expected: %v, Got: %v", i, expected, actual)
		}
	}
}
