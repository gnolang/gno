package forms

import (
	"std"
	"time"

	"gno.land/p/demo/seqid"
)

type Field struct {
	Label string
	FieldType string
	Required bool
}

type Form struct {
	ID    string
	Owner std.Address
	Title       string
	Description string
	Fields      []Field
	CreatedAt time.Time
	openAt    *time.Time
	closeAt   *time.Time
}

type Submission struct {
	FormID string
	Author std.Address
	Answers string // json
	SubmittedAt time.Time
}

type FormDB struct {
	Forms []*Form
	Answers []*Submission
	IDCounter seqid.ID
}

func NewDB() *FormDB {
	return &FormDB{
		Forms:   make([]*Form, 0),
		Answers: make([]*Submission, 0),
	}
}

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

	return time.Now().After(openAt) && time.Now().Before(closedAt)
}

func (form *Form) OpenAt() (time.Time, error) {
	if form.openAt == nil {
		return time.Time{}, errNoOpenDate
	}

	return *form.openAt, nil
}

func (form *Form) CloseAt() (time.Time, error) {
	if form.closeAt == nil {
		return time.Time{}, errNoCloseDate
	}

	return *form.closeAt, nil
}

func (db *FormDB) GetForm(id string) (*Form, error) {
	for _, form := range db.Forms {
		if form.ID == id {
			return form, nil
		}
	}
	return nil, errFormNotFound
}

func (db *FormDB) GetAnswer(formID string, author std.Address) (*Submission, error) {
	for _, answer := range db.Answers {
		if answer.FormID == formID && answer.Author.String() == author.String() {
			return answer, nil
		}
	}
	return nil, errAnswerNotFound
}

func (db *FormDB) GetSubmissionsByFormID(formID string) []*Submission {
	submissions := make([]*Submission, 0)

	for _, answer := range db.Answers {
		if answer.FormID == formID {
			submissions = append(submissions, answer)
		}
	}

	return submissions
}
