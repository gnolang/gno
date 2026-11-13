package table

import (
	"strings"

	"gno.land/p/nt/ufmt/v0"
)

// Table defines the structure for a markdown table
type Table struct {
	header []string
	rows   [][]string
}

// Validate checks if the number of columns in each row matches the number of columns in the header
func (t *Table) Validate() error {
	numCols := len(t.header)
	for _, row := range t.rows {
		if len(row) != numCols {
			return ufmt.Errorf("row %v does not match header length %d", row, numCols)
		}
	}
	return nil
}

// New creates a new Table instance, ensuring the header and rows match in size
func New(header []string, rows [][]string) (*Table, error) {
	t := &Table{
		header: header,
		rows:   rows,
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t, nil
}

// Table returns a markdown string for the given Table
func (t *Table) String() string {
	if err := t.Validate(); err != nil {
		panic(err)
	}

	var sb strings.Builder

	sb.WriteString("| " + strings.Join(t.header, " | ") + " |\n")
	sb.WriteString("| " + strings.Repeat("---|", len(t.header)) + "\n")

	for _, row := range t.rows {
		sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}

	return sb.String()
}

// AddRow adds a new row to the table
func (t *Table) AddRow(row []string) error {
	if len(row) != len(t.header) {
		return ufmt.Errorf("row %v does not match header length %d", row, len(t.header))
	}
	t.rows = append(t.rows, row)
	return nil
}

// AddColumn adds a new column to the table with the specified values
func (t *Table) AddColumn(header string, values []string) error {
	if len(values) != len(t.rows) {
		return ufmt.Errorf("values length %d does not match the number of rows %d", len(values), len(t.rows))
	}

	// Add the new header
	t.header = append(t.header, header)

	// Add the new column values to each row
	for i, value := range values {
		t.rows[i] = append(t.rows[i], value)
	}
	return nil
}

// RemoveRow removes a row from the table by its index
func (t *Table) RemoveRow(index int) error {
	if index < 0 || index >= len(t.rows) {
		return ufmt.Errorf("index %d is out of range", index)
	}
	t.rows = append(t.rows[:index], t.rows[index+1:]...)
	return nil
}

// RemoveColumn removes a column from the table by its index
func (t *Table) RemoveColumn(index int) error {
	if index < 0 || index >= len(t.header) {
		return ufmt.Errorf("index %d is out of range", index)
	}

	// Remove the column from the header
	t.header = append(t.header[:index], t.header[index+1:]...)

	// Remove the corresponding column from each row
	for i := range t.rows {
		t.rows[i] = append(t.rows[i][:index], t.rows[i][index+1:]...)
	}
	return nil
}
