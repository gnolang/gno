package pollrealm

import (
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/avl"
)

// resetState is a helper function to reset global state for testing
func resetState() {
	polls = nil
	polls = avl.NewTree()
	pollIDCounter = 0
}

func TestCreatePoll(t *testing.T) {
	// Reset global state for test
	resetState()

	deadline := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	result := CreatePoll(
		"Test Poll",
		"Test Description",
		2,
		deadline,
	)

	if !strings.Contains(result, "Successfully created poll #0000001") {
		t.Errorf("unexpected create result: %s", result)
	}

	if polls.Size() != 1 {
		t.Errorf("expected 1 poll, got %d", polls.Size())
	}
}

func TestCreatePollValidation(t *testing.T) {
	// Reset global state
	resetState()

	pastDeadline := time.Now().Add(-time.Hour).Format(time.RFC3339)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for past deadline")
		}
	}()

	CreatePoll(
		"Test Poll",
		"Test Description",
		1,
		pastDeadline,
	)
}

func TestAddAndEditOptions(t *testing.T) {
	// Reset global state
	resetState()

	deadline := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	CreatePoll(
		"Test Poll",
		"Test Description",
		2,
		deadline,
	)

	// Add options
	result := AddOption("0000001", "Option 1")
	if !strings.Contains(result, "Added a new option to poll #0000001") {
		t.Errorf("unexpected add result: %s", result)
	}

	AddOption("0000001", "Option 2")

	// Edit option
	result = EditOption("0000001", 0, "Updated Option 1")
	if !strings.Contains(result, "Edited option #0 in poll 0000001") {
		t.Errorf("unexpected edit result: %s", result)
	}

	// Delete option
	result = DeleteOption("0000001", 0)
	if !strings.Contains(result, "Deleted option #0 from poll 0000001") {
		t.Errorf("unexpected delete result: %s", result)
	}
}

func TestVoting(t *testing.T) {
	// Reset global state
	resetState()

	deadline := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	CreatePoll(
		"Test Poll",
		"Test Description",
		2,
		deadline,
	)

	AddOption("0000001", "Option 1")
	AddOption("0000001", "Option 2")
	AddOption("0000001", "Option 3")

	// Test successful vote
	result := Vote("0000001", "0,1")
	if !strings.Contains(result, "Successfully voted on poll #0000001") {
		t.Errorf("unexpected vote result: %s", result)
	}

	// Test invalid vote (too many choices)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for too many choices")
		}
	}()
	Vote("1", "0,1,2")
}

func TestEditPoll(t *testing.T) {
	// Reset global state
	resetState()

	deadline := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	CreatePoll(
		"Original Title",
		"Original Description",
		1,
		deadline,
	)

	newDeadline := time.Now().Add(time.Hour * 48).Format(time.RFC3339)
	result := EditPoll(
		"0000001",
		"New Title",
		"New Description",
		newDeadline,
		2,
	)

	if !strings.Contains(result, "Poll #0000001 edited successfully") {
		t.Errorf("unexpected edit result: %s", result)
	}
}

func TestRender(t *testing.T) {
	// Reset global state
	resetState()

	deadline := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	CreatePoll(
		"Test Poll",
		"Test Description",
		2,
		deadline,
	)

	AddOption("0000001", "Option 1")
	AddOption("0000001", "Option 2")

	// Test poll list rendering
	listOutput := Render("")
	if !strings.Contains(listOutput, "# All Polls") {
		t.Errorf("poll list should contain title")
	}
	if !strings.Contains(listOutput, "Test Poll") {
		t.Errorf("poll list should contain poll title")
	}

	// Test single poll rendering
	pollOutput := Render("0000001")
	if !strings.Contains(pollOutput, "# Poll 0000001: Test Poll") {
		t.Errorf("single poll view should contain poll title")
	}
	if !strings.Contains(pollOutput, "Option 1") {
		t.Errorf("single poll view should contain options")
	}

	// Test pagination
	for i := 0; i < 6; i++ { // Create enough polls for pagination
		CreatePoll(
			"Test Poll "+string(rune('A'+i)),
			"Description",
			1,
			deadline,
		)
	}

	page2Output := Render("page=2")
	if !strings.Contains(page2Output, "Page 2") {
		t.Errorf("pagination not working correctly")
	}
}

func TestInvalidPollID(t *testing.T) {
	// Reset global state
	resetState()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid poll ID")
		}
	}()

	Vote("999", "0") // Should panic - poll doesn't exist
}

func TestOptionManagementAfterVoting(t *testing.T) {
	// Reset global state
	resetState()

	deadline := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	CreatePoll(
		"Test Poll",
		"Test Description",
		1,
		deadline,
	)

	AddOption("0000001", "Option 1")
	AddOption("0000001", "Option 2")

	Vote("0000001", "0")

	// Attempt to add option after voting
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when adding option after voting")
		}
	}()
	AddOption("0000001", "Option 3")
}
