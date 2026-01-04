package forms

import "errors"

var (
	errNoOpenDate       = errors.New("Form has no open date")
	errNoCloseDate      = errors.New("Form has no close date")
	errInvalidJson      = errors.New("Invalid JSON")
	errInvalidDate      = errors.New("Invalid date")
	errFormNotFound     = errors.New("Form not found")
	errAnswerNotFound   = errors.New("Answer not found")
	errAlreadySubmitted = errors.New("You already submitted this form")
	errFormClosed       = errors.New("Form is closed")
	errInvalidAnswers   = errors.New("Invalid answers")
)
