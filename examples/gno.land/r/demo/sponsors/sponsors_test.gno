package sponsors

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/testutils"
)

// TestSponsors tests the sponsor management functionalities
func TestSponsors(t *testing.T) {
	caller := testutils.TestAddress("sponsor1")

	std.TestSetRealm(std.NewUserRealm(caller))
	std.TestSetOrigCaller(caller) // TODO(bug): should not be needed

	// Test registering a new sponsor
	RegisterSponsor("Sponsor1", "v1", "http://example.com", "Test Sponsor", "test@example.com")

	// Check if the sponsor was registered
	sponsor := GetSponsorByName("Sponsor1")
	if sponsor == nil {
		t.Error("Sponsor1 not found")
	}

	if sponsor.Name != "Sponsor1" {
		t.Errorf("Expected name, got %s", sponsor.Name)
	}
	if sponsor.ApiVersion != "v1" {
		t.Errorf("Expected apiVersion v1, got %s", sponsor.ApiVersion)
	}
	if sponsor.Address != caller {
		t.Errorf("Expected address %s, got %s", caller.String(), sponsor.Address.String())
	}
	if sponsor.DiscoverURL != "http://example.com" {
		t.Errorf("Expected url http://example.com, got %s", sponsor.DiscoverURL)
	}
	if sponsor.ContactInfo != "test@example.com" {
		t.Errorf("Expected contactInfo test@example.com, got %s", sponsor.ContactInfo)
	}

	// Retrieve the sponsor by address
	sponsor = GetSponsorByAddress(caller)
	if sponsor == nil {
		t.Error("Sponsor1 not found by address")
	}

	if sponsor.Name != "Sponsor1" {
		t.Errorf("Expected name, got %s", sponsor.Name)
	}
	if sponsor.ApiVersion != "v1" {
		t.Errorf("Expected apiVersion v1, got %s", sponsor.ApiVersion)
	}
	if sponsor.Address != caller {
		t.Errorf("Expected address %s, got %s", caller.String(), sponsor.Address.String())
	}
	if sponsor.DiscoverURL != "http://example.com" {
		t.Errorf("Expected url http://example.com, got %s", sponsor.DiscoverURL)
	}
	if sponsor.ContactInfo != "test@example.com" {
		t.Errorf("Expected contactInfo test@example.com, got %s", sponsor.ContactInfo)
	}

	// Retrieve the sponsor's URL by name
	url := GetURLByName("Sponsor1")
	if url != "http://example.com" {
		t.Error("Sponsor1 not found")
	}

	// Retrieve the sponsor's URL by address
	url = GetURLByAddress(caller)
	if url != "http://example.com" {
		t.Error("Sponsor1 not found by address")
	}

	// Test duplicate registration by name
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for duplicate name registration")
		}
	}()
	RegisterSponsor("Sponsor1", "v1", "http://example.com", "Test Sponsor", "test@example.com")

	// Retrieve the sponsor by address
	sponsor = GetSponsorByAddress(caller)
	if sponsor == nil || sponsor.Name != "Sponsor1" {
		t.Errorf("Sponsor1 not found by address")
	}

	// Test the GetTotalSponsors function
	RegisterSponsor("Sponsor2", "v2", "http://example2.com", "Test Sponsor 2", "test2@example.com")

	if GetTotalSponsors() != 2 {
		t.Errorf("Expected 2 sponsors, got %d", GetTotalSponsors())
	}

	// Test the GetList function
	s := GetList()
	if len(s) != 4 {
		t.Errorf("Expected 4 sponsors, got %d", len(s))
	}

	if s[0] != "Sponsor1" {
		t.Errorf("Expected Sponsor1, got %s", s[0])
	}
	if s[1] != "http://example.com" {
		t.Errorf("Expected http://example.com, got %s", s[0])
	}
	if s[2] != "Sponsor2" {
		t.Errorf("Expected Sponsor2, got %s", s[0])
	}
	if s[3] != "http://example2.com" {
		t.Errorf("Expected http://example2.com, got %s", s[0])
	}

	// Render sponsor by name
	rendered := Render("Sponsor1")
	expected := "## Sponsor: Sponsor1\n\n * API Version: v1\n * Discover URL: http://example.com\n * Description: Test Sponsor 1\n * Contact Info: test1@example.com\n * Height: " + strconv.FormatInt(std.GetHeight(), 10) + "\n"
	if rendered != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, rendered)
	}

	// Render sponsor by address
	rendered = Render(caller.String())
	if rendered != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, rendered)
	}

	// Render home
	expectedHome := " * [Sponsor1](/r/demo/sponsors:Sponsor1)\n"
	if Render("") != expectedHome {
		t.Errorf("Expected:\n%s\nGot:\n%s", expectedHome, Render(""))
	}
}
