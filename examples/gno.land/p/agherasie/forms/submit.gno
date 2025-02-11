package forms

import (
	"std"
	"time"
)

// This function allows to submit a form
func (db *FormDB) SubmitForm(formID string, answers string) {
	// Check if form exists
	form, err := db.GetForm(formID)
	if err != nil {
		panic(err)
	}

	// Check if form was already submitted by this user
	previousAnswer, err := db.GetAnswer(formID, std.PrevRealm().Addr())
	if previousAnswer != nil {
		panic(errAlreadySubmitted)
	}

	// Check time restrictions
	if !form.IsOpen() {
		panic(errFormClosed)
	}

	// Check if answers are formatted correctly
	if ValidateAnswers(answers, form.Fields) == false {
		panic(errInvalidAnswers)
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
