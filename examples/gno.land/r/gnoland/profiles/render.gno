package profiles

import (
	"bytes"
	"net/url"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
)

const (
	BaseURL           = "/r/gnoland/profiles"
	SetStringFieldURL = BaseURL + "$help&func=SetStringField&field=%s"
	SetIntFieldURL    = BaseURL + "$help&func=SetIntField&field=%s"
	SetBoolFieldURL   = BaseURL + "$help&func=SetBoolField&field=%s"
)

func homeHandler(res *mux.ResponseWriter, _ *mux.Request) {
	var b bytes.Buffer

	intro := "# Gno.land Profiles\n\n"

	intro += `This realm provides a standardized way to set and share basic
profile information on-chain, like your display name, bio, homepage, and 
more. Other realms can read your profile data from here, so users don’t have 
to set it up multiple times in multiple places, creating redundant and fragmented data. It’s a simple,
reusable profile layer meant to make your info easily accessible across the Gno ecosystem if you choose so.

Below are transaction links that allow for setting the specific fields for the caller address.

For a short doc on how to use this realm to get user data in your app, see [this](/r/docs/profiles).
`

	b.WriteString(intro)
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
	addr := address(req.GetVar("addr"))

	b.WriteString(ufmt.Sprintf("# Profile %s\n", addr))

	for field := range stringFields {
		value := GetStringField(addr, field)
		link := ufmt.Sprintf(SetStringFieldURL, field)
		b.WriteString(ufmt.Sprintf("- %s: %s [Edit](%s)\n", field, value, link))
	}

	for field := range intFields {
		value := GetIntField(addr, field, 0)
		link := ufmt.Sprintf(SetIntFieldURL, field)
		b.WriteString(ufmt.Sprintf("- %s: %d [Edit](%s)\n", field, value, link))
	}

	for field := range boolFields {
		value := GetBoolField(addr, field)
		link := ufmt.Sprintf(SetBoolFieldURL, field)
		b.WriteString(ufmt.Sprintf("- %s: %t [Edit](%s)\n", field, value, link))
	}

	res.Write(b.String())
}

func fieldHandler(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	addr := address(req.GetVar("addr"))
	field := req.GetVar("field")

	b.WriteString(ufmt.Sprintf("# Field %s for %s\n", field, addr))

	value := "n/a"
	var editLink string

	if _, ok := stringFields[field]; ok {
		value = ufmt.Sprintf("%s", GetStringField(addr, field))
		editLink = ufmt.Sprintf(SetStringFieldURL+"&addr=%s&value=%s", field, addr, url.QueryEscape(value))
	} else if _, ok := intFields[field]; ok {
		value = ufmt.Sprintf("%d", GetIntField(addr, field, 0))
		editLink = ufmt.Sprintf(SetIntFieldURL+"&addr=%s&value=%s", field, addr, value)
	} else if _, ok := boolFields[field]; ok {
		value = ufmt.Sprintf("%t", GetBoolField(addr, field))
		editLink = ufmt.Sprintf(SetBoolFieldURL+"&addr=%s&value=%s", field, addr, value)
	}

	b.WriteString(ufmt.Sprintf("- %s: %s [Edit](%s)\n", field, value, editLink))

	res.Write(b.String())
}

func Render(path string) string {
	return router.Render(path)
}
