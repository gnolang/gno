package mdtable_test

import (
	"testing"

	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/urequire/v0"
)

// XXX: switch to `func Example() {}` when supported.
func TestExample(t *testing.T) {
	table := mdtable.Table{
		Headers: []string{"ID", "Title", "Status"},
		Rows: [][]string{
			{"#1", "Add a new validator", "succeed"},
			{"#2", "Change parameter", "timed out"},
			{"#3", "Fill pool", "active"},
		},
	}

	got := table.String()
	expected := `| ID | Title | Status |
| --- | --- | --- |
| #1 | Add a new validator | succeed |
| #2 | Change parameter | timed out |
| #3 | Fill pool | active |
`

	urequire.Equal(t, got, expected)
}

func TestTableString(t *testing.T) {
	tests := []struct {
		name     string
		table    mdtable.Table
		expected string
	}{
		{
			name: "With Headers and Rows",
			table: mdtable.Table{
				Headers: []string{"ID", "Title", "Status", "Date"},
				Rows: [][]string{
					{"#1", "Add a new validator", "succeed", "2024-01-01"},
					{"#2", "Change parameter", "timed out", "2024-01-02"},
				},
			},
			expected: `| ID | Title | Status | Date |
| --- | --- | --- | --- |
| #1 | Add a new validator | succeed | 2024-01-01 |
| #2 | Change parameter | timed out | 2024-01-02 |
`,
		},
		{
			name: "Without Headers",
			table: mdtable.Table{
				Rows: [][]string{
					{"#1", "Add a new validator", "succeed", "2024-01-01"},
					{"#2", "Change parameter", "timed out", "2024-01-02"},
				},
			},
			expected: `|  |  |  |  |
| --- | --- | --- | --- |
| #1 | Add a new validator | succeed | 2024-01-01 |
| #2 | Change parameter | timed out | 2024-01-02 |
`,
		},
		{
			name: "Without Rows",
			table: mdtable.Table{
				Headers: []string{"ID", "Title", "Status", "Date"},
			},
			expected: `| ID | Title | Status | Date |
| --- | --- | --- | --- |
`,
		},
		{
			name: "With Pipe Character in Content",
			table: mdtable.Table{
				Headers: []string{"ID", "Title", "Status", "Date"},
				Rows: [][]string{
					{"#1", "Add a new | validator", "succeed", "2024-01-01"},
					{"#2", "Change parameter", "timed out", "2024-01-02"},
				},
			},
			expected: `| ID | Title | Status | Date |
| --- | --- | --- | --- |
| #1 | Add a new &#124; validator | succeed | 2024-01-01 |
| #2 | Change parameter | timed out | 2024-01-02 |
`,
		},
		{
			name: "With Varying Row Sizes", // XXX: should we have a different behavior?
			table: mdtable.Table{
				Headers: []string{"ID", "Title"},
				Rows: [][]string{
					{"#1", "Add a new validator"},
					{"#2", "Change parameter", "Extra Column"},
					{"#3", "Fill pool"},
				},
			},
			expected: `| ID | Title |
| --- | --- |
| #1 | Add a new validator |
| #2 | Change parameter | Extra Column |
| #3 | Fill pool |
`,
		},
		{
			name: "With UTF-8 Characters",
			table: mdtable.Table{
				Headers: []string{"ID", "Title", "Status", "Date"},
				Rows: [][]string{
					{"#1", "Café", "succeed", "2024-01-01"},
					{"#2", "München", "timed out", "2024-01-02"},
					{"#3", "São Paulo", "active", "2024-01-03"},
				},
			},
			expected: `| ID | Title | Status | Date |
| --- | --- | --- | --- |
| #1 | Café | succeed | 2024-01-01 |
| #2 | München | timed out | 2024-01-02 |
| #3 | São Paulo | active | 2024-01-03 |
`,
		},
		{
			name:     "With no Headers and no Rows",
			table:    mdtable.Table{},
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.table.String()
			urequire.Equal(t, got, tt.expected)
		})
	}
}

func TestTableAppend(t *testing.T) {
	table := mdtable.Table{
		Headers: []string{"ID", "Title", "Status", "Date"},
	}

	// Use the Append method to add rows to the table
	table.Append([]string{"#1", "Add a new validator", "succeed", "2024-01-01"})
	table.Append([]string{"#2", "Change parameter", "timed out", "2024-01-02"})
	table.Append([]string{"#3", "Fill pool", "active", "2024-01-03"})
	got := table.String()

	expected := `| ID | Title | Status | Date |
| --- | --- | --- | --- |
| #1 | Add a new validator | succeed | 2024-01-01 |
| #2 | Change parameter | timed out | 2024-01-02 |
| #3 | Fill pool | active | 2024-01-03 |
`
	urequire.Equal(t, got, expected)
}
