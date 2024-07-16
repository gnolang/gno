package forms

import (
	"testing"

	"gno.land/p/demo/testutils"
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

	formID := db.CreateForm("Test Form", "Test Description", "2020-01-01", "2020-01-02", data)
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
