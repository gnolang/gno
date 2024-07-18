package forms

import (
	"std"
	"time"

	"gno.land/p/demo/ufmt"
)

func (db *FormDatabase) SubmitForm(formID string, answers string) {
	// Check if form exists
	form := db.GetForm(formID)
	if form == nil {
		panic(ufmt.Errorf("Form not found: %s", formID))
	}

	// Check if form was already submitted by this user
	previousAnswer := db.GetAnswer(formID, std.PrevRealm().Addr())
	if previousAnswer != nil {
		panic("You already submitted this form")
	}

	// Check time restrictions
	if !form.IsOpen() {
		panic("Form is closed")
	}

	// Check if answers are formatted correctly
	if ValidateAnswers(answers, form.Fields) == false {
		panic("Invalid answers")
	}

	// Save answers
	answer := Submission{
		FormID:      formID,
		Answers:     answers,
		Author:      std.PrevRealm().Addr(),
		SubmittedAt: time.Now(),
	}
	db.Answers = append(db.Answers, &answer)
}

