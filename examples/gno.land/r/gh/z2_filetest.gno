// z2_filetest_extended.gno

package main

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/gh"
)

const (
	adminAddr  = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	oracleAddr = std.Address("g1eunnckcl6r8ncwj0lrpxu9g5062xcvwxqlrf29")
	aliceAddr  = testutils.TestAddress("alice")
	bobAddr    = testutils.TestAddress("bob")
)

func main() {
	t := &testing.T{}

	testAdminSetOracleAddr(t)
	testOracleUpsertAccount(t)
	testOracleUpsertRepo(t)
	testRepoByID(t)
	testAccountByID(t)
	testRepoFields(t)
	testAlicePublicFunctions(t)

	if t.Failed() {
		panic("Some tests failed.")
	} else {
		println("All tests passed!")
	}
}

func testAdminSetOracleAddr(t *testing.T) {
	std.TestSetOrigCaller(adminAddr)
	newOracleAddr := std.Address("g1neworcl1234567890abcdef1234567890abcdef")
	gh.AdminSetOracleAddr(newOracleAddr)

	expected := newOracleAddr
	result := gh.AdminGetOracleAddr()
	if result != expected {
		t.Errorf("Expected oracle address to be set to %s, got %s", expected, result)
	}

	// Resetting back to original oracle for subsequent tests
	gh.AdminSetOracleAddr(oracleAddr)
}

func testOracleUpsertAccount(t *testing.T) {
	std.TestSetOrigCaller(oracleAddr)
	gh.OracleUpsertAccount("moul", "Manfred", "user")
	gh.OracleUpsertAccount("gnolang", "Gno Ecosystem", "org")

	moul := gh.AccountByID("moul")
	gnolang := gh.AccountByID("gnolang")

	{
		expected := "Manfred"
		result := moul.Name()
		if result != expected {
			t.Errorf("Expected moul's name to be %s, got %s", expected, result)
		}
	}

	{
		expected := "org"
		result := gnolang.Kind()
		if result != expected {
			t.Errorf("Expected gnolang's kind to be %s, got %s", expected, result)
		}
	}
}

func testOracleUpsertRepo(t *testing.T) {
	std.TestSetOrigCaller(oracleAddr)
	gh.OracleUpsertRepo("gnolang/gno", false, false)

	repo := gh.RepoByID("gnolang/gno")
	expected := "gno"
	result := repo.Name()
	if result != expected {
		t.Errorf("Expected repo name to be %s, got %s", expected, result)
	}
}

func testRepoByID(t *testing.T) {
	repo := gh.RepoByID("gnolang/gno")
	if repo == nil {
		t.Errorf("Expected to find a repo with ID gnolang/gno, got nil")
	}
}

func testAccountByID(t *testing.T) {
	acc := gh.AccountByID("moul")
	if acc == nil {
		t.Errorf("Expected to find an account with ID moul, got nil")
	}
}

func testRepoFields(t *testing.T) {
	repo := gh.RepoByID("gnolang/gno")
	if repo.IsFork() {
		t.Errorf("Expected repo to not be a fork")
	}

	if repo.IsPrivate() {
		t.Errorf("Expected repo to not be private")
	}

	expected := "https://github.com/gnolang/gno"
	result := repo.URL()
	if result != expected {
		t.Errorf("Expected repo URL to be %s, got %s", expected, result)
	}
}

func testAlicePublicFunctions(t *testing.T) {
	std.TestSetOrigCaller(aliceAddr)

	// Test AccountByID
	acc := gh.AccountByID("moul")
	if acc == nil || acc.Name() != "Manfred" {
		t.Errorf("Alice could not fetch the account 'moul'.")
	}

	// Test RepoByID
	repo := gh.RepoByID("gnolang/gno")
	if repo == nil || repo.Name() != "gno" {
		t.Errorf("Alice could not fetch the repo 'gnolang/gno'.")
	}
}

// Output:
// All tests passed!
