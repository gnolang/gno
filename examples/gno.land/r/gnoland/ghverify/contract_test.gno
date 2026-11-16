package ghverify

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// init whitelists the deployer alone, so any other address is refused
// as an agent. It runs before TestVerificationLifecycle reseeds the whitelist.
func TestInitWhitelistsOnlyTheDeployer(cur realm, t *testing.T) {
	stranger := address(testutils.TestAddress("stranger"))
	uassert.NotEqual(t, string(ownerAddress), string(stranger))

	testing.SetOriginCaller(stranger)
	RequestVerification(cross(cur), "stranger")

	uassert.AbortsWithMessage(t, cur, "caller not whitelisted", func() {
		GnorkleEntrypoint(cross(cur), "ingest,"+string(stranger)+",OK")
	})

	oracle.RemoveFeed(string(stranger))
}

func TestVerificationLifecycle(cur realm, t *testing.T) {
	defaultAddress := address(testutils.TestAddress("owner"))
	user1Address := address(testutils.TestAddress("user 1"))
	user2Address := address(testutils.TestAddress("user 2"))

	// init captures the owner from the deployer, which no test has, so the
	// owner and the oracle whitelist are seeded here.
	ownerAddress = defaultAddress
	oracle.ClearWhitelist("")
	oracle.AddToWhitelist("", []string{string(defaultAddress)})
	testing.SetOriginCaller(defaultAddress)

	// Verify request returns no feeds.
	result := GnorkleEntrypoint(cross(cur), "request")
	if result != "[]" {
		t.Fatalf("expected empty request result, got %s", result)
	}

	// Make a verification request with the created user.
	testing.SetOriginCaller(user1Address)
	RequestVerification(cross(cur), "deelawn")

	// A subsequent request from the same address aborts because there is
	// already a feed with an ID of this user's address.
	uassert.AbortsWithMessage(t, cur, "feed already exists", func() {
		RequestVerification(cross(cur), "deelawn")
	})

	// Verify the request returns no feeds for this non-whitelisted user.
	result = GnorkleEntrypoint(cross(cur), "request")
	if result != "[]" {
		t.Fatalf("expected empty request result, got %s", result)
	}

	// Make a verification request with the created user.
	testing.SetOriginCaller(user2Address)
	RequestVerification(cross(cur), "omarsy")

	// Set the caller back to the whitelisted user and verify that the feed data
	// returned matches what should have been created by the `RequestVerification`
	// invocation.
	testing.SetOriginCaller(defaultAddress)
	result = GnorkleEntrypoint(cross(cur), "request")
	expResult := `[{"id":"` + string(user1Address) + `","type":"0","value_type":"string","tasks":[{"gno_address":"` +
		string(user1Address) + `","github_handle":"deelawn"}]},` +
		`{"id":"` + string(user2Address) + `","type":"0","value_type":"string","tasks":[{"gno_address":"` +
		string(user2Address) + `","github_handle":"omarsy"}]}]`
	if result != expResult {
		t.Fatalf("expected request result %s, got %s", expResult, result)
	}

	// Try to trigger feed ingestion from the non-authorized user.
	testing.SetOriginCaller(user1Address)
	uassert.AbortsWithMessage(t, cur, "caller not whitelisted", func() {
		GnorkleEntrypoint(cross(cur), "ingest,"+string(user1Address)+",OK")
	})

	// Set the caller back to the whitelisted user and transfer contract ownership.
	testing.SetOriginCaller(defaultAddress)
	SetOwner(cross(cur), defaultAddress)

	// Now trigger the feed ingestion from the user and new owner and only whitelisted address.
	GnorkleEntrypoint(cross(cur), "ingest,"+string(user1Address)+",OK")
	GnorkleEntrypoint(cross(cur), "ingest,"+string(user2Address)+",OK")

	// Verify the ingestion autocommitted the value and triggered the post handler.
	data := Render("")
	expResult = `{"deelawn": "` + string(user1Address) + `","omarsy": "` + string(user2Address) + `"}`
	if data != expResult {
		t.Fatalf("expected render data %s, got %s", expResult, data)
	}

	// Finally make sure the feed was cleaned up after the data was committed.
	result = GnorkleEntrypoint(cross(cur), "request")
	if result != "[]" {
		t.Fatalf("expected empty request result, got %s", result)
	}

	// Check that the accessor functions are working as expected.
	if handle := GetHandleByAddress(cur, string(user1Address)); handle != "deelawn" {
		t.Fatalf("expected deelawn, got %s", handle)
	}
	if address_XXX := GetAddressByHandle(cur, "deelawn"); address_XXX != string(user1Address) {
		t.Fatalf("expected %s, got %s", string(user1Address), address_XXX)
	}
}

// A call arriving through a realm is authenticated as that realm, so the
// owner's signature does not carry into it.
func TestSetOwnerFromRealm(cur realm, t *testing.T) {
	owner := address(testutils.TestAddress("owner"))
	proposedOwner := address(testutils.TestAddress("proposedOwner"))

	ownerAddress = owner
	testing.SetOriginCaller(owner)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/intermediate"))

	uassert.AbortsWithMessage(t, cur, "only the owner can set a new owner", func() {
		SetOwner(cross(cur), proposedOwner)
	})

	uassert.Equal(t, string(owner), string(ownerAddress))
}
