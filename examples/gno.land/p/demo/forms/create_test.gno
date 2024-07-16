package forms

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestCreateForm(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetOrigCaller(alice)

	title := "Simple Form"
	description := "This is a form"
	openAt := "2021-01-01T00:00:00Z"
	closeAt := "2021-01-02T00:00:00Z"
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
			"fieldType": "['Pizza', 'Schnitzel', 'Burger']",
			"required": true
		},
		{
			"label": "Favorite Foods",
			"fieldType": "{'Pizza', 'Schnitzel', 'Burger'}",
			"required": true
		}
	]`
	db := NewDatabase()
	id := db.CreateForm(title, description, openAt, closeAt, data)
	if id == "" {
		t.Error("Form ID is empty")
	}
	form := db.GetForm(id)
	if form == nil {
		t.Error("Form is nil")
	}
	if form.Owner != alice {
		t.Error("Owner is not correct")
	}
	if form.Title != title {
		t.Error("Title is not correct")
	}
	if form.Description != description {
		t.Error("Description is not correct")
	}
	if len(form.Fields) != 5 {
		t.Error("Fields are not correct")
	}
	if form.Fields[0].Label != "Name" {
		t.Error("Field 0 label is not correct")
	}
	if form.Fields[0].FieldType != "string" {
		t.Error("Field 0 type is not correct")
	}
	if form.Fields[0].Required != true {
		t.Error("Field 0 required is not correct")
	}
	if form.Fields[1].Label != "Age" {
		t.Error("Field 1 label is not correct")
	}
	if form.Fields[1].FieldType != "number" {
		t.Error("Field 1 type is not correct")
	}
}
