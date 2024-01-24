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
	expResult := `[{"id":"deelawn","type":"0","value_type":"string","tasks":[{"gnoAddress":"` +
		string(userAddress) +
		`","githubHandle":"deelawn"}]}]`
	if result != expResult {
		t.Fatalf("expected request result %s, got %s", expResult, result)
	}

	// Try to trigger feed ingestion from the non-authorized user.
	std.TestSetOrigCaller(userAddress)
	var errMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				errMsg = r.(string)
			}
		}()
		GnorkleEntrypoint("ingest,deelawn,OK")
	}()
	if errMsg != "caller not whitelisted" {
		t.Fatalf("expected caller not whitelisted, got %s", errMsg)
	}

	// Set the caller back to the whitelisted user and transfer contract ownership.
	std.TestSetOrigCaller(defaultAddress)
	SetOwner(string(userAddress))

	// Now trigger the feed ingestion from the user and new owner and only whitelisted address.
	std.TestSetOrigCaller(userAddress)
	GnorkleEntrypoint("ingest,deelawn,OK")

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
}
