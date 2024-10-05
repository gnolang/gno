package markdown

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestNewTable(t *testing.T) {
	header := []string{"Name", "Age", "Country"}
	rows := [][]string{
		{"Alice", "30", "USA"},
		{"Bob", "25", "UK"},
	}

	table, err := NewTable(header, rows)
	urequire.NoError(t, err)

	uassert.Equal(t, len(header), len(table.header))
	uassert.Equal(t, len(rows), len(table.rows))
}

func Test_AddRow(t *testing.T) {
	header := []string{"Name", "Age"}
	rows := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	}

	table, err := NewTable(header, rows)
	urequire.NoError(t, err)

	// Add a valid row
	err = table.AddRow([]string{"Charlie", "28"})
	urequire.NoError(t, err)

	expectedRows := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
		{"Charlie", "28"},
	}
	uassert.Equal(t, len(expectedRows), len(table.rows))

	// Attempt to add a row with a different number of columns
	err = table.AddRow([]string{"David"})
	uassert.Error(t, err)
}

func Test_AddColumn(t *testing.T) {
	header := []string{"Name", "Age"}
	rows := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	}

	table, err := NewTable(header, rows)
	urequire.NoError(t, err)

	// Add a valid column
	err = table.AddColumn("Country", []string{"USA", "UK"})
	urequire.NoError(t, err)

	expectedHeader := []string{"Name", "Age", "Country"}
	expectedRows := [][]string{
		{"Alice", "30", "USA"},
		{"Bob", "25", "UK"},
	}
	uassert.Equal(t, len(expectedHeader), len(table.header))
	uassert.Equal(t, len(expectedRows), len(table.rows))

	// Attempt to add a column with a different number of values
	err = table.AddColumn("City", []string{"New York"})
	uassert.Error(t, err)
}

func Test_RemoveRow(t *testing.T) {
	header := []string{"Name", "Age", "Country"}
	rows := [][]string{
		{"Alice", "30", "USA"},
		{"Bob", "25", "UK"},
	}

	table, err := NewTable(header, rows)
	urequire.NoError(t, err)

	// Remove the first row
	err = table.RemoveRow(0)
	urequire.NoError(t, err)

	expectedRows := [][]string{
		{"Bob", "25", "UK"},
	}
	uassert.Equal(t, len(expectedRows), len(table.rows))

	// Attempt to remove a row out of range
	err = table.RemoveRow(5)
	uassert.Error(t, err)
}

func Test_RemoveColumn(t *testing.T) {
	header := []string{"Name", "Age", "Country"}
	rows := [][]string{
		{"Alice", "30", "USA"},
		{"Bob", "25", "UK"},
	}

	table, err := NewTable(header, rows)
	urequire.NoError(t, err)

	// Remove the second column (Age)
	err = table.RemoveColumn(1)
	urequire.NoError(t, err)

	expectedHeader := []string{"Name", "Country"}
	expectedRows := [][]string{
		{"Alice", "USA"},
		{"Bob", "UK"},
	}
	uassert.Equal(t, len(expectedHeader), len(table.header))
	uassert.Equal(t, len(expectedRows), len(table.rows))

	// Attempt to remove a column out of range
	err = table.RemoveColumn(5)
	uassert.Error(t, err)
}

func Test_ValidateColumns(t *testing.T) {
	header := []string{"Name", "Age", "Country"}
	rows := [][]string{
		{"Alice", "30", "USA"},
		{"Bob", "25"},
	}

	table, err := NewTable(header, rows[:1])
	urequire.NoError(t, err)

	// Validate columns should pass
	err = table.ValidateColumns()
	urequire.NoError(t, err)

	// Add an invalid row and validate again
	table.rows = append(table.rows, rows[1])
	err = table.ValidateColumns()
	uassert.Error(t, err)
}
