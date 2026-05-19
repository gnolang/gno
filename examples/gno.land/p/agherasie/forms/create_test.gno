package forms

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
func TestCreateForm(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))

	db := NewDB()
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

	urequire.NotPanics(t, cur, func() {
		var id string
		var err error
		func(cur realm) {
			id, err = db.CreateForm(0, cur, title, description, openAt, closeAt, data)
		}(cross(cur))
		if err != nil {
			panic(err)
		}
		urequire.True(t, id != "", "Form ID is empty")

		form, err := db.GetForm(id)
		if err != nil {
			panic(err)
		}

		urequire.True(t, form.ID == id, "Form ID is not correct")
		urequire.Equal(t, form.Owner, alice, "Owner is not correct")
		urequire.True(t, form.Title == title, "Title is not correct")
		urequire.True(t, form.Description == description, "Description is not correct")
		urequire.True(t, len(form.Fields) == 5, "Not enough fields were provided")
		urequire.True(t, form.Fields[0].Label == "Name", "Field 0 label is not correct")
		urequire.True(t, form.Fields[0].FieldType == "string", "Field 0 type is not correct")
		urequire.True(t, form.Fields[0].Required == true, "Field 0 required is not correct")
		urequire.True(t, form.Fields[1].Label == "Age", "Field 1 label is not correct")
		urequire.True(t, form.Fields[1].FieldType == "number", "Field 1 type is not correct")
	})
}
