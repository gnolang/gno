package forms

import (
	"time"
	"testing"
)

func TestAnswerForm(t *testing.T) {
	db := NewDatabase()

	data := `[
		{
			"label": "Name",
			"fieldType": "string",
			"required": true
		},
		{
			"label": "Age",
			"fieldType": "number",
			"required": false
		},
		{
			"label": "Is this a test?",
			"fieldType": "boolean",
			"required": false
		},
		{
			"label": "Favorite Food",
			"fieldType": "[Pizza|Schnitzel|Burger]",
			"required": true
		},
		{
			"label": "Favorite Foods",
			"fieldType": "{Pizza|Schnitzel|Burger}",
			"required": true
		}
	]`

	formID := db.CreateForm("Test Form", "Test Description", "", "", data)
	answers := `["Alex", 21, true, 0, [0, 1]]`

	db.SubmitForm(formID, answers)

	if len(db.Answers) != 1 {
		t.Errorf("Expected 1 answer, got %d", len(db.Answers))
	}

	if db.Answers[0].FormID != formID {
		t.Errorf("Expected form ID %s, got %s", formID, db.Answers[0].FormID)
	}

	if db.Answers[0].Answers != answers {
		t.Errorf("Expected answers %s, got %s", answers, db.Answers[0].Answers)
	}
}

func shouldPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should have panic")
		}
	}()
	f()
}

func TestAnswerFormDates(t *testing.T) {
	db := NewDatabase()

	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02T15:04:05Z")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02T15:04:05Z")

	data := `[
		{
			"label": "Name",
			"fieldType": "string",
			"required": true
		}
	]`
	answers := `["Test"]`

	shouldPanic(t, func() {
		formID := db.CreateForm("Test Form", "Test Description", tomorrow, "", data)
		db.SubmitForm(formID, answers)
	})

	shouldPanic(t, func() {
		formID := db.CreateForm("Test Form", "Test Description", "", yesterday, data)
		db.SubmitForm(formID, answers)
	})

	formID := db.CreateForm("Test Form", "Test Description", yesterday, tomorrow, data)
	db.SubmitForm(formID, answers)
}
