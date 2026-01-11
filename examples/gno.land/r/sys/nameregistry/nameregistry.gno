package nameregistry

import (
	"strings"

	"gno.land/p/moul/md"
)

// TODO: Allow updating validators

// Validators contains a list of name validators.
var validators = []Validator{
	NewMinLengthValidator(4),
	NewMaxLengthValidator(14),
	NewLowercaseValidator(),
	NewForbiddenPrefixValidator("g1", "gi", "gl"),
	NewAlphaNumValidator(),
}

// IsNameValid checks is a name is a valid name.
func IsNameValid(name string) error {
	for _, v := range validators {
		if err := v.Validate(name); err != nil {
			return err
		}
	}
	return nil
}

func Render(string) string {
	var b strings.Builder

	b.WriteString(md.H1("Name Registry"))
	b.WriteString(md.H2("Validation"))
	b.WriteString(md.Paragraph("Names must match the following rules:"))

	for _, v := range validators {
		b.WriteString(md.BulletItem(v.Description()))
	}

	return b.String()
}
