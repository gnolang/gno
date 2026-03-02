package group

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/r/chaegnal/organizer"
)

func Test_ParseQueryKeyword(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	url := "https://example.com/page?user=Chaegnal&groupId=5&address=Ghhv"

	user := parseQueryKeyword(url, "user")
	if user != "Chaegnal" {
		t.Errorf("Expected 'Chaegnal', got '%s'", user)
	}

	id := parseQueryKeyword(url, "groupId")
	if id != "5" {
		t.Errorf("Expected '5', got '%s'", id)
	}

	token := parseQueryKeyword(url, "address")
	if token != "Ghhv" {
		t.Errorf("Expected 'Ghhv', got '%s'", token)
	}

	nonexistent := parseQueryKeyword(url, "abc")
	if nonexistent != "" {
		t.Errorf("Expected empty string for missing keyword, got '%s'", nonexistent)
	}
}

// Test getGroupFromUrl
func Test_getGroupFromUrl_Success(t *testing.T) {
	testing.SetRealm(std.NewUserRealm("Chaegnal"))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	uassert.NotPanics(t, func() { getGroupFromUrl("http://test.com?groupId=1") })
}

func Test_getGroupFromUrl_Failure(t *testing.T) {
	uassert.PanicsWithMessage(t, "no group id provided", func() { getGroupFromUrl("http://test.com?groupId=") })
}
