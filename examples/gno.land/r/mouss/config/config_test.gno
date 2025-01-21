package config

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	mainAddr         = Address()
	backupAddr       = testutils.TestAddress("backup")
	unauthorizedAddr = testutils.TestAddress("unauthorized")
)

func TestAddress(t *testing.T) {
	addr := Address()
	expected := std.Address("g1hrfvdh7jdvnlxpk2y20tp3scj9jqal3zzu7wjz")
	uassert.Equal(t, expected, addr, "Address() should return initialized main address")
}

func TestSetBackup(t *testing.T) {
	// Test setting backup as main address
	std.TestSetOrigCaller(mainAddr)
	validAddr := testutils.TestAddress("validbackup")
	err := SetBackup(validAddr)
	uassert.NoError(t, err, "main address should be able to set backup")

	// Test setting invalid address format
	err = SetBackup(std.Address("invalid"))
	uassert.ErrorIs(t, err, ErrorInvalidAddr, "should reject invalid address format")

	// Test setting empty address
	err = SetBackup(std.Address(""))
	uassert.ErrorIs(t, err, ErrorInvalidAddr, "should reject empty address")

	// Test unauthorized caller
	std.TestSetOrigCaller(unauthorizedAddr)
	err = SetBackup(validAddr)
	uassert.ErrorIs(t, err, ErrorUnauthorized, "should reject unauthorized caller")
}

func TestCheckAuthorized(t *testing.T) {
	// Test main address authorization
	std.TestSetOrigCaller(mainAddr)
	err := checkAuthorized()
	uassert.NoError(t, err, "main address should be authorized")

	// Test unauthorized address
	std.TestSetOrigCaller(unauthorizedAddr)
	err = checkAuthorized()
	uassert.ErrorIs(t, err, ErrorUnauthorized, "random address should not be authorized")

	// Set and test backup address authorization
	std.TestSetOrigCaller(mainAddr)
	err = SetBackup(backupAddr)
	uassert.NoError(t, err, "setting backup address should succeed")

	std.TestSetOrigCaller(backupAddr)
	err = checkAuthorized()
	uassert.NoError(t, err, "backup address should be authorized")
}
