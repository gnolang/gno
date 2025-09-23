package forms

import (
	"time"

	"gno.land/p/nt/seqid"
)

// FieldType examples :
// - string: "string";
// - number: "number";
// - boolean: "boolean";
// - choice: "['Pizza', 'Schnitzel', 'Burger']";
// - multi-choice: "{'Pizza', 'Schnitzel', 'Burger'}";
type Field struct {
	Label     string
	FieldType string
	Required  bool
}

type Form struct {
	ID          string
	Owner       address
	Title       string
	Description string
	Fields      []Field
	CreatedAt   time.Time
	openAt      time.Time
	closeAt     time.Time
}

// Answers example :
// - ["Alex", 21, true, 0, [0, 1]]
type Submission struct {
	FormID      string
	Author      address
	Answers     string // json
	SubmittedAt time.Time
}

type FormDB struct {
	Forms     []*Form
	Answers   []*Submission
	IDCounter seqid.ID
}

func NewDB() *FormDB {
	return &FormDB{
		Forms:   make([]*Form, 0),
		Answers: make([]*Submission, 0),
	}
}

// This function checks if the form is open by verifying the given dates
// - If a form doesn't have any dates, it's considered open
// - If a form has only an open date, it's considered open if the open date is in the past
// - If a form has only a close date, it's considered open if the close date is in the future
// - If a form has both open and close dates, it's considered open if the current date is between the open and close dates
func (form *Form) IsOpen() bool {
	openAt, errOpen := form.OpenAt()
	closedAt, errClose := form.CloseAt()

	noOpenDate := errOpen != nil
	noCloseDate := errClose != nil

	if noOpenDate && noCloseDate {
		return true
	}

	if noOpenDate && !noCloseDate {
		return time.Now().Before(closedAt)
	}

	if !noOpenDate && noCloseDate {
		return time.Now().After(openAt)
	}

	now := time.Now()
	return now.After(openAt) && now.Before(closedAt)
}

// OpenAt returns the open date of the form if it exists
func (form *Form) OpenAt() (time.Time, error) {
	if form.openAt.IsZero() {
		return time.Time{}, errNoOpenDate
	}

	return form.openAt, nil
}

// CloseAt returns the close date of the form if it exists
func (form *Form) CloseAt() (time.Time, error) {
	if form.closeAt.IsZero() {
		return time.Time{}, errNoCloseDate
	}

	return form.closeAt, nil
}

// GetForm returns a form by its ID if it exists
func (db *FormDB) GetForm(id string) (*Form, error) {
	for _, form := range db.Forms {
		if form.ID == id {
			return form, nil
		}
	}
	return nil, errFormNotFound
}

// GetAnswer returns an answer by its form - and author ids if it exists
func (db *FormDB) GetAnswer(formID string, author address) (*Submission, error) {
	for _, answer := range db.Answers {
		if answer.FormID == formID && answer.Author.String() == author.String() {
			return answer, nil
		}
	}
	return nil, errAnswerNotFound
}

// GetSubmissionsByFormID returns a list containing the existing form submissions by the form ID
func (db *FormDB) GetSubmissionsByFormID(formID string) []*Submission {
	submissions := make([]*Submission, 0)

	for _, answer := range db.Answers {
		if answer.FormID == formID {
			submissions = append(submissions, answer)
		}
	}

	return submissions
}
