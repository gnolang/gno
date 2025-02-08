package config

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	mainAddr         = Address()
	backupAddr       = Backup()
	unauthorizedAddr = testutils.TestAddress("unauthorized")
)

func TestAddress(t *testing.T) {
	addr := Address()
	expected := std.Address("g1hrfvdh7jdvnlxpk2y20tp3scj9jqal3zzu7wjz")
	uassert.Equal(t, expected, addr, "Address() should return initialized main address")
}

func TestSetBackup(t *testing.T) {
	addr := Backup()
	expected := std.Address("g1wq2h93ppkf2gkgncz5unayrsmt7pl8npktnznd")
	uassert.Equal(t, expected, addr, "Backup() should return initialized main address")
}

func TestAddConfig(t *testing.T) {
	// Test unauthorized add
	std.TestSetOrigCaller(unauthorizedAddr)
	defer func() {
		if r := recover(); r != ErrUnauthorized {
			t.Errorf("Expected unauthorized panic")
		}
	}()
	AddConfig("test", "test content")
	t.Error("Expected panic")

	// Test successful add
	std.TestSetOrigCaller(mainAddr)
	AddConfig("config1", "content1")

	// Test duplicate name
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected duplicate name panic")
		}
	}()
	AddConfig("config1", "content2")
	t.Error("Expected panic")
}
func TestRemoveConfig(t *testing.T) {
	// Setup test config
	std.TestSetOrigCaller(mainAddr)
	AddConfig("test", "test content")

	// Test unauthorized remove
	std.TestSetOrigCaller(unauthorizedAddr)
	defer func() {
		if r := recover(); r != ErrUnauthorized {
			t.Errorf("Expected unauthorized panic")
		}
	}()
	RemoveConfig("test")
	t.Error("Expected panic")

	// Test successful remove
	std.TestSetOrigCaller(mainAddr)
	RemoveConfig("test")

	// Test remove non-existent
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected non-existent config panic")
		}
	}()
	RemoveConfig("test")
	t.Error("Expected panic")
}
func TestIsAuthorized(t *testing.T) {
	// Test main address
	auth := IsAuthorized(Address())
	uassert.Equal(t, true, auth, "main address should be authorized")

	// Test backup address
	auth = IsAuthorized(Backup())
	uassert.Equal(t, true, auth, "backup address should be authorized")

	// Test unauthorized address
	auth = IsAuthorized(unauthorizedAddr)
	uassert.Equal(t, false, auth, "random address should not be authorized")
}
