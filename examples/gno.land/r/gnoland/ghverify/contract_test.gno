package ghverify

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils/v0"
)

func TestVerificationLifecycle(cur realm, t *testing.T) {
	defaultAddress := runtime.OriginCaller()
	user1Address := address(testutils.TestAddress("user 1"))
	user2Address := address(testutils.TestAddress("user 2"))

	// Verify request returns no feeds.
	result := GnorkleEntrypoint(cur, "request")
	if result != "[]" {
		t.Fatalf("expected empty request result, got %s", result)
	}

	// Make a verification request with the created user.
	testing.SetOriginCaller(user1Address)
	RequestVerification(cur, "deelawn")

	// A subsequent request from the same address should panic because there is
	// already a feed with an ID of this user's address.
	var errMsg string
	func(cur realm) {
		defer func() {
			if r := recover(); r != nil {
				errMsg = r.(error).Error()
			}
		}()
		RequestVerification(cur, "deelawn")
	}(cur)
	if errMsg != "feed already exists" {
		t.Fatalf("expected feed already exists, got %s", errMsg)
	}

	// Verify the request returns no feeds for this non-whitelisted user.
	result = GnorkleEntrypoint(cur, "request")
	if result != "[]" {
		t.Fatalf("expected empty request result, got %s", result)
	}

	// Make a verification request with the created user.
	testing.SetOriginCaller(user2Address)
	RequestVerification(cur, "omarsy")

	// Set the caller back to the whitelisted user and verify that the feed data
	// returned matches what should have been created by the `RequestVerification`
	// invocation.
	testing.SetOriginCaller(defaultAddress)
	result = GnorkleEntrypoint(cur, "request")
	expResult := `[{"id":"` + string(user1Address) + `","type":"0","value_type":"string","tasks":[{"gno_address":"` +
		string(user1Address) + `","github_handle":"deelawn"}]},` +
		`{"id":"` + string(user2Address) + `","type":"0","value_type":"string","tasks":[{"gno_address":"` +
		string(user2Address) + `","github_handle":"omarsy"}]}]`
	if result != expResult {
		t.Fatalf("expected request result %s, got %s", expResult, result)
	}

	// Try to trigger feed ingestion from the non-authorized user.
	testing.SetOriginCaller(user1Address)
	func(cur realm) {
		defer func() {
			if r := recover(); r != nil {
				errMsg = r.(error).Error()
			}
		}()
		GnorkleEntrypoint(cur, "ingest,"+string(user1Address)+",OK")
	}(cur)
	if errMsg != "caller not whitelisted" {
		t.Fatalf("expected caller not whitelisted, got %s", errMsg)
	}

	// Set the caller back to the whitelisted user and transfer contract ownership.
	testing.SetOriginCaller(defaultAddress)
	SetOwner(cross, defaultAddress)

	// Now trigger the feed ingestion from the user and new owner and only whitelisted address.
	GnorkleEntrypoint(cur, "ingest,"+string(user1Address)+",OK")
	GnorkleEntrypoint(cur, "ingest,"+string(user2Address)+",OK")

	// Verify the ingestion autocommitted the value and triggered the post handler.
	data := Render("")
	expResult = `{"deelawn": "` + string(user1Address) + `","omarsy": "` + string(user2Address) + `"}`
	if data != expResult {
		t.Fatalf("expected render data %s, got %s", expResult, data)
	}

	// Finally make sure the feed was cleaned up after the data was committed.
	result = GnorkleEntrypoint(cur, "request")
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
