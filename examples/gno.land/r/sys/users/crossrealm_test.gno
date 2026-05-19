package users

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// TestUpdateNameCallerIdentity verifies that the CurrentRealm() check in
// UpdateName correctly identifies the calling controller.
//
// Background: RegisterUser uses PreviousRealm() because it is a crossing
// function (cur realm). UpdateName uses CurrentRealm() because it is a
// non-crossing method on *UserData. Per the interrealm spec:
//
//   - Crossing function: CurrentRealm = this realm, PreviousRealm = caller
//   - Non-crossing method on external object, called from crossing context:
//     CurrentRealm = caller (unchanged from crossing context)
//
// This test verifies that a whitelisted controller can call UpdateName,
// and a non-whitelisted realm cannot.
func TestUpdateNameCallerIdentity(t *testing.T) {
	controllerPath := initControllerPath
	nonControllerPath := "gno.land/r/evil/attacker"

	t.Run("whitelisted_controller_can_update", func(t *testing.T) {
		cleanStore(t)

		// Register as whitelisted controller
		testing.SetRealm(testing.NewCodeRealm(controllerPath))
		urequire.NoError(t, RegisterUser(cross, "testuser1", testutils.TestAddress("testuser1")))

		// Resolve and update name — should succeed because controller is whitelisted
		data := ResolveAddress(testutils.TestAddress("testuser1"))
		urequire.NotEqual(t, nil, data)
		uassert.NoError(t, data.UpdateName("newname1"))
		uassert.Equal(t, "newname1", data.Name())
	})

	t.Run("non_whitelisted_realm_cannot_update", func(t *testing.T) {
		cleanStore(t)

		// Register as whitelisted controller
		testing.SetRealm(testing.NewCodeRealm(controllerPath))
		urequire.NoError(t, RegisterUser(cross, "testuser2", testutils.TestAddress("testuser2")))

		data := ResolveAddress(testutils.TestAddress("testuser2"))
		urequire.NotEqual(t, nil, data)

		// Switch to non-whitelisted realm — UpdateName should fail
		testing.SetRealm(testing.NewCodeRealm(nonControllerPath))
		err := data.UpdateName("hacked")
		uassert.ErrorContains(t, err, "does not exist in whitelist")
		uassert.Equal(t, "testuser2", data.Name()) // name unchanged
	})
}

// TestDeleteCallerIdentity verifies the same CurrentRealm() behavior for Delete.
func TestDeleteCallerIdentity(t *testing.T) {
	controllerPath := initControllerPath
	nonControllerPath := "gno.land/r/evil/attacker"

	t.Run("whitelisted_controller_can_delete", func(t *testing.T) {
		cleanStore(t)

		testing.SetRealm(testing.NewCodeRealm(controllerPath))
		urequire.NoError(t, RegisterUser(cross, "deluser1", testutils.TestAddress("deluser1")))

		data := ResolveAddress(testutils.TestAddress("deluser1"))
		urequire.NotEqual(t, nil, data)
		uassert.NoError(t, data.Delete())
		uassert.True(t, data.IsDeleted())
	})

	t.Run("non_whitelisted_realm_cannot_delete", func(t *testing.T) {
		cleanStore(t)

		testing.SetRealm(testing.NewCodeRealm(controllerPath))
		urequire.NoError(t, RegisterUser(cross, "deluser2", testutils.TestAddress("deluser2")))

		data := ResolveAddress(testutils.TestAddress("deluser2"))
		urequire.NotEqual(t, nil, data)

		// Switch to non-whitelisted realm — Delete should fail
		testing.SetRealm(testing.NewCodeRealm(nonControllerPath))
		err := data.Delete()
		uassert.ErrorContains(t, err, "does not exist in whitelist")
		uassert.False(t, data.IsDeleted()) // not deleted
	})
}
