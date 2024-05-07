package ghverify

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestVerificationLifecycle(t *testing.T) {
	defaultAddress := std.GetOrigCaller()
	userAddress := std.Address(testutils.TestAddress("user"))

	// Verify request returns no feeds.
	result := GnorkleEntrypoint("request")
	if result != "[]" {
		t.Fatalf("expected empty request result, got %s", result)
	}

	// Make a verification request with the created user.
	std.TestSetOrigCaller(userAddress)
	RequestVerification("deelawn")

	// A subsequent request from the same address should panic because there is
	// already a feed with an ID of this user's address.
	var errMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				errMsg = r.(error).Error()
			}
		}()
		RequestVerification("deelawn")
	}()
	if errMsg != "feed already exists" {
		t.Fatalf("expected feed already exists, got %s", errMsg)
	}

	// Verify the request returns no feeds for this non-whitelisted user.
	result = GnorkleEntrypoint("request")
	if result != "[]" {
		t.Fatalf("expected empty request result, got %s", result)
	}

	// Set the caller back to the whitelisted user and verify that the feed data
	// returned matches what should have been created by the `RequestVerification`
	// invocation.
	std.TestSetOrigCaller(defaultAddress)
	result = GnorkleEntrypoint("request")
	expResult := `[{"id":"` + string(userAddress) + `","type":"0","value_type":"string","tasks":[{"gno_address":"` +
		string(userAddress) + `","github_handle":"deelawn"}]}]`
	if result != expResult {
		t.Fatalf("expected request result %s, got %s", expResult, result)
	}

	// Try to trigger feed ingestion from the non-authorized user.
	std.TestSetOrigCaller(userAddress)
	func() {
		defer func() {
			if r := recover(); r != nil {
				errMsg = r.(error).Error()
			}
		}()
		GnorkleEntrypoint("ingest," + string(userAddress) + ",OK")
	}()
	if errMsg != "caller not whitelisted" {
		t.Fatalf("expected caller not whitelisted, got %s", errMsg)
	}

	// Set the caller back to the whitelisted user and transfer contract ownership.
	std.TestSetOrigCaller(defaultAddress)
	SetOwner(userAddress)

	// Now trigger the feed ingestion from the user and new owner and only whitelisted address.
	std.TestSetOrigCaller(userAddress)
	GnorkleEntrypoint("ingest," + string(userAddress) + ",OK")

	// Verify the ingestion autocommitted the value and triggered the post handler.
	data := Render("")
	expResult = `{"deelawn": "` + string(userAddress) + `"}`
	if data != expResult {
		t.Fatalf("expected render data %s, got %s", expResult, data)
	}

	// Finally make sure the feed was cleaned up after the data was committed.
	result = GnorkleEntrypoint("request")
	if result != "[]" {
		t.Fatalf("expected empty request result, got %s", result)
	}

	// Check that the accessor functions are working as expected.
	if handle := GetHandleByAddress(string(userAddress)); handle != "deelawn" {
		t.Fatalf("expected deelawn, got %s", handle)
	}
	if address := GetAddressByHandle("deelawn"); address != string(userAddress) {
		t.Fatalf("expected %s, got %s", string(userAddress), address)
	}
}
