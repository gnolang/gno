package config

import (
	"testing"

	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/testutils/v0"
)

func TestAddBackupOwner(t *testing.T) {
	owner := address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	AddBackupOwner(cross, u1)
	b := BackupOwners()
	if b[1] != u1.String() {
		t.Error("failed to add u1 to backupowners")
	}
	testing.SetOriginCaller(u1)
	r := revive(func() {
		AddBackupOwner(cross, u2)
	})
	if r != ownable.ErrUnauthorized {
		t.Error("failed to catch unauthorized access")
	}

	testing.SetOriginCaller(owner)
	RemoveBackupOwner(cross, u1)
	RemoveBackupOwner(cross, u2)
}

func TestRemoveBackupOwner(t *testing.T) {
	owner := address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	AddBackupOwner(cross, u1)

	testing.SetOriginCaller(u2)
	r := revive(func() {
		RemoveBackupOwner(cross, u1)
	})
	if r != ownable.ErrUnauthorized {
		t.Error("failed to catch unauthorized access")
	}

	testing.SetOriginCaller(owner)
	RemoveBackupOwner(cross, u1)

	if len(BackupOwners()) != 1 {
		t.Error("BackupOwners should be length == 1 ")
	}
}

func TestClaimOwnership(t *testing.T) {
	owner := address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")

	if owner != Owner() {
		t.Errorf("expected: %v, got: %v", owner, Owner())
	}

	testing.SetOriginCaller(owner)
	AddBackupOwner(cross, u1)

	testing.SetOriginCaller(u1)
	ClaimOwnership(cross)

	if u1 != Owner() {
		t.Errorf("expected: %v, got: %v", owner, Owner())
	}

	testing.SetOriginCaller(owner)
	ClaimOwnership(cross)
}

func TestAddAdmin(t *testing.T) {
	owner := address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	AddAdmin(cross, u1)
	admins := Admins()
	if admins[1] != u1.String() {
		t.Error("failed to add u1 to admins group")
	}
	testing.SetOriginCaller(u1)
	r := revive(func() {
		AddAdmin(cross, u2)
	})
	if r != ownable.ErrUnauthorized {
		t.Error("failed to catch unauthorized access")
	}

	// cleanup
	testing.SetOriginCaller(owner)
	RemoveAdmin(cross, u1)
}

func TestRemoveAdmin(t *testing.T) {
	owner := address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	AddAdmin(cross, u1)

	testing.SetOriginCaller(u2)
	r := revive(func() {
		RemoveAdmin(cross, u1)
	})
	if r != ownable.ErrUnauthorized {
		t.Error("failed to catch unauthorized access")
	}

	testing.SetOriginCaller(owner)
	RemoveAdmin(cross, u1)

	if len(Admins()) != 1 {
		t.Error("Admin should be length == 1 ")
	}
}

func TestIsAdmin(t *testing.T) {
	owner := address("g1j39fhg29uehm7twwnhvnpz3ggrm6tprhq65t0t")
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")

	testing.SetOriginCaller(owner)
	AddAdmin(cross, u1)

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
