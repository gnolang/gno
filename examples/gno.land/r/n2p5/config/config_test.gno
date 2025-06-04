package config

import (
	"std"
	"testing"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/testutils"
)

func TestAddBackupOwner(t *testing.T) {
	owner := std.Address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	cross(AddBackupOwner)(u1)
	b := BackupOwners()
	if b[1] != u1.String() {
		t.Error("failed to add u1 to backupowners")
	}
	testing.SetOriginCaller(u1)
	r := revive(func() {
		cross(AddBackupOwner)(u2)
	})
	if r != ownable.ErrUnauthorized {
		t.Error("failed to catch unauthorized access")
	}

	testing.SetOriginCaller(owner)
	cross(RemoveBackupOwner)(u1)
	cross(RemoveBackupOwner)(u2)
}

func TestRemoveBackupOwner(t *testing.T) {
	owner := std.Address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	cross(AddBackupOwner)(u1)

	testing.SetOriginCaller(u2)
	r := revive(func() {
		cross(RemoveBackupOwner)(u1)
	})
	if r != ownable.ErrUnauthorized {
		t.Error("failed to catch unauthorized access")
	}

	testing.SetOriginCaller(owner)
	cross(RemoveBackupOwner)(u1)

	if len(BackupOwners()) != 1 {
		t.Error("BackupOwners should be length == 1 ")
	}
}

func TestClaimOwnership(t *testing.T) {
	owner := std.Address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")

	if owner != Owner() {
		t.Errorf("expected: %v, got: %v", owner, Owner())
	}

	testing.SetOriginCaller(owner)
	cross(AddBackupOwner)(u1)

	testing.SetOriginCaller(u1)
	cross(ClaimOwnership)()

	if u1 != Owner() {
		t.Errorf("expected: %v, got: %v", owner, Owner())
	}

	testing.SetOriginCaller(owner)
	cross(ClaimOwnership)()
}

func TestAddAdmin(t *testing.T) {
	owner := std.Address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	cross(AddAdmin)(u1)
	admins := Admins()
	if admins[1] != u1.String() {
		t.Error("failed to add u1 to admins group")
	}
	testing.SetOriginCaller(u1)
	r := revive(func() {
		cross(AddAdmin)(u2)
	})
	if r != ownable.ErrUnauthorized {
		t.Error("failed to catch unauthorized access")
	}

	// cleanup
	testing.SetOriginCaller(owner)
	cross(RemoveAdmin)(u1)
}

func TestRemoveAdmin(t *testing.T) {
	owner := std.Address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	cross(AddAdmin)(u1)

	testing.SetOriginCaller(u2)
	r := revive(func() {
		cross(RemoveAdmin)(u1)
	})
	if r != ownable.ErrUnauthorized {
		t.Error("failed to catch unauthorized access")
	}

	testing.SetOriginCaller(owner)
	cross(RemoveAdmin)(u1)

	if len(Admins()) != 1 {
		t.Error("Admin should be length == 1 ")
	}
}

func TestIsAdmin(t *testing.T) {
	owner := std.Address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	cross(AddAdmin)(u1)

	if !IsAdmin(owner) {
		t.Error("owner should be admin")
	}
	if !IsAdmin(u1) {
		t.Error("u1 should be admin")
	}
	if IsAdmin(u2) {
		t.Error("u2 should not be admin")
	}
}
