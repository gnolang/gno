// Package mdform provides a simple way to create Gno forms.
//
// Example usage:
//
//	import "gno.land/p/moul/mdform"
//
//	func Render(path string) string {
//	    table := mdtable.Table{
//	        Headers: []string{"ID", "Title", "Status", "Date"},
//	    }
//	    table.Append([]string{"#1", "Add a new validator", "succeed", "2024-01-01"})
//	    table.Append([]string{"#2", "Change parameter", "timed out", "2024-01-02"})
//	    return table.String()
//	}
//
// Output:
//
//	<gno-
//	| --- | --- | --- | --- |
//	| #1 | Add a new validator | succeed | 2024-01-01 |
//	| #2 | Change parameter | timed out | 2024-01-02 |
package mdform

import (
	"strings"
)

// The input type is the type for the forms.
type InputType uint

const (
	Text     InputType = iota // Plain text like "hello"
	Number                    // Any number like 42
	Email                     // String who follows the email regex: '^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$'
	Tel                       // A phone number: +33 6 12 34 56 78
	Password                  // A text that is hidden
)

var input2str = map[InputType]string {
	Text: "text",
	Number: "number",
	Email: "email",
	Tel: "tel",
	Password: "password",
}

type Input struct {
	Name        string
	PlaceHolder string
	Type        InputType
}

type Form struct {
	Path string
	Inputs []Input
}

func (f *Form) SetPath(path string) {
	f.Path = path
}

func (f *Form) addInput(t InputType, n, p string) {
	f.Inputs = append(f.Inputs, Input{
		Name: n,
		PlaceHolder: p,
		Type: t,
	})
}

func (f *Form) AddInput(name, placeholder string) {
	f.addInput(Text, name, placeholder)
}

func (f *Form) AddTextInput(name, placeholder string) {
	f.addInput(Text, name, placeholder)
}

func (f *Form) AddNumberInput(name, placeholder string) {
	f.addInput(Number, name, placeholder)
}

func (f *Form) AddEmailInput(name, placeholder string) {
	f.addInput(Email, name, placeholder)
}

func (f *Form) AddTelInput(name, placeholder string) {
	f.addInput(Tel, name, placeholder)
}

func (f *Form) AddPasswordInput(name, placeholder string) {
	f.addInput(Password, name, placeholder)
}

func (f Form) String() string {
	var sb strings.Builder

	// Write the begin of the form
	str := "<gno-form"
	if f.Path != "" {
		str += " path=\"" + strings.ReplaceAll(f.Path, "\"", "&#x22") + "\""
	}
	str += ">\n"
	sb.WriteString(str)

	// Generate input lines
	for _, v := range f.Inputs {
		// Use text by default
		x, ok := input2str[v.Type]
		if !ok {
			x = input2str[Text]
		}

		// Write the line
		sb.WriteString("  <gno-input name=\"" +
			strings.ReplaceAll(v.Name, "\"", "&#x22") +
			"\" type=\"" +
			strings.ReplaceAll(x, "\"", "&#x22") +
			"\" placeholder=\"" +
			strings.ReplaceAll(v.PlaceHolder, "\"", "&#x22") +
			"\" />\n")
	}

	// Write the end of the form
	sb.WriteString("</gno-form>\n")
	return sb.String()
}
