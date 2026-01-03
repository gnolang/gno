package mdform

import (
	"html"
	"strings"
)

const (
	InputTypeText     = "text"
	InputTypeNumber   = "number"
	InputTypeEmail    = "email"
	InputTypePhone    = "tel"
	InputTypePassword = "password"
	InputTypeRadio    = "radio"
	InputTypeCheckbox = "checkbox"
)

var (
	formAttributes     = []string{"exec", "path"}
	inputAttributes    = []string{"checked", "description", "placeholder", "readonly", "type", "value"}
	textareaAttributes = []string{"placeholder", "readonly", "rows", "value"}
	selectAttributes   = []string{"description", "readonly", "selected"}
)

// New creates a new form.
func New(attributes ...string) *Form {
	assertEvenAttributes(attributes)

	form := &Form{}
	for i := 0; i < len(attributes); i += 2 {
		name, value := attributes[i], attributes[i+1]

		assertIsValidAttribute(name, formAttributes)

		form.attrs = append(form.attrs, formatAttribute(name, value))
	}
	return form
}

// Form is a form that can be rendered to Gno-Flavored Markdown.
type Form struct {
	attrs  []string
	fields []string
}

// Input appends a new input to form fields.
// Use `Form.Radio()` or `Form.Checkbox()` to append those types of inputs to the form.
// Method panics when appending inputs of type radio or checkbox, or when attributes are not valid.
func (f *Form) Input(name string, attributes ...string) *Form {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("form input name is required")
	}

	assertEvenAttributes(attributes)

	attrs := []string{formatAttribute("name", name)}
	for i := 0; i < len(attributes); i += 2 {
		name, value := attributes[i], attributes[i+1]
		if name == "type" {
			switch value {
			case InputTypeRadio:
				panic("use form.Radio() to create inputs of type radio")
			case InputTypeCheckbox:
				panic("use form.Checkbox() to create inputs of type checkbox")
			}
		}

		assertIsValidAttribute(name, inputAttributes)

		attrs = append(attrs, formatAttribute(name, value))
	}

	f.fields = append(f.fields, "<gno-input "+strings.Join(attrs, " ")+" />")
	return f
}

// Radio appends a new input of type radio to form fields.
// Method panics when attributes are not valid.
func (f *Form) Radio(name, value string, attributes ...string) *Form {
	return f.appendInputType(InputTypeRadio, name, value, attributes...)
}

// Checkbox appends a new input of type checkbox to form fields.
// Method panics when attributes are not valid.
func (f *Form) Checkbox(name, value string, attributes ...string) *Form {
	return f.appendInputType(InputTypeCheckbox, name, value, attributes...)
}

// Textarea appends a new textarea to form fields.
// Method panics when attributes are not valid.
func (f *Form) Textarea(name string, attributes ...string) *Form {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("form textarea name is required")
	}

	assertEvenAttributes(attributes)

	attrs := []string{formatAttribute("name", name)}
	for i := 0; i < len(attributes); i += 2 {
		name, value := attributes[i], attributes[i+1]

		assertIsValidAttribute(name, textareaAttributes)

		attrs = append(attrs, formatAttribute(name, value))
	}

	f.fields = append(f.fields, "<gno-textarea "+strings.Join(attrs, " ")+" />")
	return f
}

// Select appends a new select to form fields.
// Method panics when attributes are not valid.
func (f *Form) Select(name, value string, attributes ...string) *Form {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("form select name is required")
	}

	assertEvenAttributes(attributes)

	attrs := []string{
		formatAttribute("name", name),
		formatAttribute("value", value),
	}

	for i := 0; i < len(attributes); i += 2 {
		name, value := attributes[i], attributes[i+1]

		assertIsValidAttribute(name, selectAttributes)

		attrs = append(attrs, formatAttribute(name, value))
	}

	f.fields = append(f.fields, "<gno-select "+strings.Join(attrs, " ")+" />")
	return f
}

// String returns the form as Gno-Flavored Markdown.
func (f Form) String() string {
	fields := strings.Join(f.fields, "\n")
	attrs := strings.Join(f.attrs, " ")
	if len(attrs) > 0 {
		attrs = " " + attrs
	}

	return "<gno-form" + attrs + ">\n" + fields + "\n</gno-form>\n"
}

func (f *Form) appendInputType(typeName, name, value string, attributes ...string) *Form {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("form " + typeName + " input name is required")
	}

	assertEvenAttributes(attributes)

	attrs := []string{
		formatAttribute("type", typeName),
		formatAttribute("name", name),
		formatAttribute("value", value),
	}

	for i := 0; i < len(attributes); i += 2 {
		name, value := attributes[i], attributes[i+1]
		if name == "type" || name == "value" {
			continue
		}

		assertIsValidAttribute(name, inputAttributes)

		attrs = append(attrs, formatAttribute(name, value))
	}

	f.fields = append(f.fields, "<gno-input "+strings.Join(attrs, " ")+" />")
	return f
}

func formatAttribute(name, value string) string {
	return name + `="` + html.EscapeString(value) + `"`
}

func assertEvenAttributes(attrs []string) {
	if len(attrs)%2 != 0 {
		panic("expected an even number of attribute arguments")
	}
}

func assertIsValidAttribute(attr string, attrs []string) {
	for _, name := range attrs {
		if name == attr {
			return
		}
	}

	panic("invalid attribute: " + attr)
}
