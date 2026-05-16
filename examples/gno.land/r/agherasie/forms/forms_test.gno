package forms

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
func TestGetFormByID(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	title := "Simple Form"
	description := "This is a form"
	openAt := "2021-01-01T00:00:00Z"
	closeAt := "2021-01-02T00:00:00Z"
	data := `[{"label":"Name","fieldType":"string","required":true},{"label":"Age","fieldType":"number","required":false},{"label":"Is this a test?","fieldType":"boolean","required":false},{"label":"Favorite Food","fieldType":"['Pizza', 'Schnitzel', 'Burger']","required":true},{"label":"Favorite Foods","fieldType":"{'Pizza', 'Schnitzel', 'Burger'}","required":true}]`

	urequire.NotPanics(t, cur, func() {
		id := CreateForm(cross2(cur), title, description, openAt, closeAt, data)

		form := GetFormByID(id)

		urequire.True(t, strings.Contains(form, data), "Form JSON was not rebuilt properly")
	})
}

func TestGetForms(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	description := "This is a form"
	openAt := "2021-01-01T00:00:00Z"
	closeAt := "2021-01-02T00:00:00Z"

	urequire.NotPanics(t, cur, func() {
		data1 := `[{"label":"Name","fieldType":"string","required":true}]`
		CreateForm(cross2(cur), "NameForm", description, openAt, closeAt, data1)
		data2 := `[{"label":"Age","fieldType":"number","required":false}]`
		CreateForm(cross2(cur), "AgeForm", description, openAt, closeAt, data2)

		forms := GetForms()

		urequire.True(t, strings.Contains(forms, data1) && strings.Contains(forms, data2), "Forms JSON were not rebuilt properly")
	})

}
