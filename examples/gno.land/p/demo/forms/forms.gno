package forms

import (
	"std"

	"gno.land/p/demo/seqid"
)

type Field struct {
	Label string

	/*
		string: "string";
		number: "number";
		boolean: "boolean";
		choice: "['Pizza', 'Schnitzel', 'Burger']";
		multi-choice: "{'Pizza', 'Schnitzel', 'Burger'}";
	*/
	FieldType string

	Required bool
}

type Form struct {
	ID    string
	Owner std.Address

	Title       string
	Description string
	Fields      []Field

	CreatedAt string
	OpenAt    string
	CloseAt   string
}

type Submission struct {
	FormID string

	Author std.Address

	/* ["Alex", 21, true, 0, [0, 1]] */
	Answers string // json

	SubmittedAt string
}

type FormDatabase struct {
	Forms []*Form

	Answers []*Submission

	IDCounter seqid.ID
}

func NewDatabase() *FormDatabase {
	return &FormDatabase{
		Forms:   make([]*Form, 0),
		Answers: make([]*Submission, 0),
	}
}
