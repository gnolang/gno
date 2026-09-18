package forms

import (
	"time"
)

// SubmitForm records an answer to a form. rlm must be the caller's own
// captured cur; the Author is set to rlm.Previous().Address() —
// unforgeable, vs. the prior runtime walk a non-crossing wrapper could
// spoof.
func (db *FormDB) SubmitForm(_ int, rlm realm, formID string, answers string) {
	// Check if form exists
	form, err := db.GetForm(formID)
	if err != nil {
		panic(err)
	}

	if !rlm.IsCurrent() {
		panic(errSpoofedRealm)
	}
	author := rlm.Previous().Address()

	// Check if form was already submitted by this user
	previousAnswer, _ := db.GetAnswer(formID, author)
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
		Author:      author,
		SubmittedAt: time.Now(),
	}
	db.Answers = append(db.Answers, &answer)
}
