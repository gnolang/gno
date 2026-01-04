package profile

import (
	"bytes"
	"net/url"

	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
)

const (
	BaseURL           = "/r/demo/profile"
	SetStringFieldURL = BaseURL + "$help&func=SetStringField&field=%s"
	SetIntFieldURL    = BaseURL + "$help&func=SetIntField&field=%s"
	SetBoolFieldURL   = BaseURL + "$help&func=SetBoolField&field=%s"
	ViewAllFieldsURL  = BaseURL + ":u/%s"
	ViewFieldURL      = BaseURL + ":f/%s/%s"
)

func homeHandler(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer

	b.WriteString("## Setters\n")
	for field := range stringFields {
		link := ufmt.Sprintf(SetStringFieldURL, field)
		b.WriteString(ufmt.Sprintf("- [Set %s](%s)\n", field, link))
	}

	for field := range intFields {
		link := ufmt.Sprintf(SetIntFieldURL, field)
		b.WriteString(ufmt.Sprintf("- [Set %s](%s)\n", field, link))
	}

	for field := range boolFields {
		link := ufmt.Sprintf(SetBoolFieldURL, field)
		b.WriteString(ufmt.Sprintf("- [Set %s Field](%s)\n", field, link))
	}

	b.WriteString("\n---\n\n")

	res.Write(b.String())
}

func profileHandler(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	addr := req.GetVar("addr")

	b.WriteString(ufmt.Sprintf("# Profile %s\n", addr))

	address_XXX := address(addr)

	for field := range stringFields {
		value := GetStringField(address_XXX, field, "n/a")
		link := ufmt.Sprintf(SetStringFieldURL, field)
		b.WriteString(ufmt.Sprintf("- %s: %s [Edit](%s)\n", field, value, link))
	}

	for field := range intFields {
		value := GetIntField(address_XXX, field, 0)
		link := ufmt.Sprintf(SetIntFieldURL, field)
		b.WriteString(ufmt.Sprintf("- %s: %d [Edit](%s)\n", field, value, link))
	}

	for field := range boolFields {
		value := GetBoolField(address_XXX, field, false)
		link := ufmt.Sprintf(SetBoolFieldURL, field)
		b.WriteString(ufmt.Sprintf("- %s: %t [Edit](%s)\n", field, value, link))
	}

	res.Write(b.String())
}

func fieldHandler(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	addr := req.GetVar("addr")
	field := req.GetVar("field")

	b.WriteString(ufmt.Sprintf("# Field %s for %s\n", field, addr))

	address_XXX := address(addr)
	value := "n/a"
	var editLink string

	if _, ok := stringFields[field]; ok {
		value = ufmt.Sprintf("%s", GetStringField(address_XXX, field, "n/a"))
		editLink = ufmt.Sprintf(SetStringFieldURL+"&addr=%s&value=%s", field, addr, url.QueryEscape(value))
	} else if _, ok := intFields[field]; ok {
		value = ufmt.Sprintf("%d", GetIntField(address_XXX, field, 0))
		editLink = ufmt.Sprintf(SetIntFieldURL+"&addr=%s&value=%s", field, addr, value)
	} else if _, ok := boolFields[field]; ok {
		value = ufmt.Sprintf("%t", GetBoolField(address_XXX, field, false))
		editLink = ufmt.Sprintf(SetBoolFieldURL+"&addr=%s&value=%s", field, addr, value)
	}

	b.WriteString(ufmt.Sprintf("- %s: %s [Edit](%s)\n", field, value, editLink))

	res.Write(b.String())
}

func Render(path string) string {
	return router.Render(path)
}
