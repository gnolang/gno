package forms

import (
	"testing"
	"time"

	"gno.land/p/nt/urequire/v0"
)

func TestAnswerForm(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm("g1user"))

	db := NewDB()

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

	formID, err := db.CreateForm("Test Form", "Test Description", "", "", data)
	if err != nil {
		t.Error(err)
	}
	answers := `["Alex", 21, true, 0, [0, 1]]`
	db.SubmitForm(formID, answers)

	urequire.True(t, len(db.Answers) == 1, "Expected 1 answer, got", string(len(db.Answers)))
	urequire.True(t, db.Answers[0].FormID == formID, "Expected form ID", formID, "got", db.Answers[0].FormID)
	urequire.True(t, db.Answers[0].Answers == answers, "Expected answers", answers, "got", db.Answers[0].Answers)
	urequire.True(t, err == nil, "Submit should not return an error")
}

func TestAnswerFormDates(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm("g1user"))
	db := NewDB()

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

	urequire.PanicsWithMessage(t, "Form is closed", func() {
		formID, err := db.CreateForm("Test Form", "Test Description", tomorrow, "", data)
		if err != nil {
			t.Error(err)
		}
		db.SubmitForm(formID, answers)
	})

	urequire.PanicsWithMessage(t, "Form is closed", func() {
		formID, err := db.CreateForm("Test Form", "Test Description", "", yesterday, data)
		if err != nil {
			t.Error(err)
		}
		db.SubmitForm(formID, answers)
	})

	urequire.NotPanics(t, func() {
		formID, err := db.CreateForm("Test Form", "Test Description", yesterday, tomorrow, data)
		if err != nil {
			panic(err)
		}
		db.SubmitForm(formID, answers)
	})
}
