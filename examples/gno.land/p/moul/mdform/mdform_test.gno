package mdform_test

import (
	"testing"

	"gno.land/p/demo/urequire"
	"gno.land/p/moul/mdform"
)

// XXX: switch to `func Example() {}` when supported.
func TestExample(t *testing.T) {
	form := mdform.Form{}
	form.AddInput("name", "Enter your name")
	form.AddInput("email", "Enter your email")

	got := form.String()
	expected := `<gno-form>
  <gno-input name="name" type="text" placeholder="Enter your name" />
  <gno-input name="email" type="text" placeholder="Enter your email" />
</gno-form>
`
	urequire.Equal(t, got, expected)
}

func TestFormString(t *testing.T) {
	tests := []struct {
		name     string
		form     mdform.Form
		expected string
	}{
		{
			name: "Default example with 2 texts",
			form: mdform.Form{
				Inputs: []mdform.Input{
					mdform.Input{Name: "name", PlaceHolder: "Enter your name", Type: mdform.Text},
					mdform.Input{Name: "email", PlaceHolder: "Enter your email", Type: mdform.Email},
				},
			},
			expected: `<gno-form>
  <gno-input name="name" type="text" placeholder="Enter your name" />
  <gno-input name="email" type="email" placeholder="Enter your email" />
</gno-form>
`,
		},
		{
			name: "Default example with 2 texts and path",
			form: mdform.Form{
				Path: "foobar",
				Inputs: []mdform.Input{
					mdform.Input{Name: "name", PlaceHolder: "Enter your name", Type: mdform.Text},
					mdform.Input{Name: "email", PlaceHolder: "Enter your email", Type: mdform.Email},
				},
			},
			expected: `<gno-form path="foobar">
  <gno-input name="name" type="text" placeholder="Enter your name" />
  <gno-input name="email" type="email" placeholder="Enter your email" />
</gno-form>
`,
		},
		{
			name: "Without Inputs",
			form: mdform.Form{},
			expected: `<gno-form>
</gno-form>
`,
		},
		{
			name: "With Quote Character",
			form: mdform.Form{
				Inputs: []mdform.Input{
					mdform.Input{Name: "name", PlaceHolder: "foo\" bar=\"toto", Type: mdform.Text},
				},
			},
			expected: `<gno-form>
  <gno-input name="name" type="text" placeholder="foo&#x22 bar=&#x22toto" />
</gno-form>
`,
		},
		{
			name: "With All types",
			form: mdform.Form{
				Inputs: []mdform.Input{
					mdform.Input{Name: "text", PlaceHolder: "", Type: mdform.Text},
					mdform.Input{Name: "number", PlaceHolder: "", Type: mdform.Number},
					mdform.Input{Name: "email", PlaceHolder: "", Type: mdform.Email},
					mdform.Input{Name: "password", PlaceHolder: "", Type: mdform.Password},
					mdform.Input{Name: "tel", PlaceHolder: "", Type: mdform.Tel},
				},
			},
			expected: `<gno-form>
  <gno-input name="text" type="text" placeholder="" />
  <gno-input name="number" type="number" placeholder="" />
  <gno-input name="email" type="email" placeholder="" />
  <gno-input name="password" type="password" placeholder="" />
  <gno-input name="tel" type="tel" placeholder="" />
</gno-form>
`,
		},
		{
			name: "With invalide input",
			form: mdform.Form{
				Inputs: []mdform.Input{
					mdform.Input{Name: "invalid", PlaceHolder: "", Type: mdform.InputType(42)},
				},
			},
			expected: `<gno-form>
  <gno-input name="invalid" type="text" placeholder="" />
</gno-form>
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.form.String()
			urequire.Equal(t, got, tt.expected)
		})
	}
}

// func TestFormAppend(t *testing.T) {
// 	table := mdtable.Form{
// 		Headers: []string{"ID", "Title", "Status", "Date"},
// 	}

// 	// Use the Append method to add rows to the table
// 	table.Append([]string{"#1", "Add a new validator", "succeed", "2024-01-01"})
// 	table.Append([]string{"#2", "Change parameter", "timed out", "2024-01-02"})
// 	table.Append([]string{"#3", "Fill pool", "active", "2024-01-03"})
// 	got := table.String()

// 	expected := `| ID | Title | Status | Date |
// | --- | --- | --- | --- |
// | #1 | Add a new validator | succeed | 2024-01-01 |
// | #2 | Change parameter | timed out | 2024-01-02 |
// | #3 | Fill pool | active | 2024-01-03 |
// `
// 	urequire.Equal(t, got, expected)
// }
