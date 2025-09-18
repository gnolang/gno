# `tablesort` - Sortable markdown tables

Generate sortable markdown tables with clickable column headers. Sorting state is managed via URL query parameters.

## Usage

```go
import "gno.land/p/samcrew/tablesort"

table := &tablesort.Table{
    Headings: []string{"Name", "Age", "City"},
    Rows: [][]string{
        {"Alice", "25", "New York"},
        {"Bob", "30", "London"},
        {"Charlie", "22", "Paris"},
    },
}

// Basic usage
u, _ := url.Parse("/users")
markdown := tablesort.Render(u, table, "")

// Multiple tables on same page (use prefix to avoid conflicts)
markdown1 := tablesort.Render(u, table, "table1-")
markdown2 := tablesort.Render(u, table, "table2-")
```

## On-chain Example

- [/r/gov/dao/v3/memberstore:members?filter=T1](/r/gov/dao/v3/memberstore:members?filter=T1)

## API

```go
type Table struct {
    Headings []string   // Column headers
    Rows     [][]string // Table data rows
}

// `u`: Current URL for generating sort links
// `table`: Table data structure
// `paramPrefix`: Prefix for URL params (use for multiple tables)
func Render(u *url.URL, table *Table, paramPrefix string) string
```

**URL Parameters:**
- `{prefix}sort-asc={column}`: Sort column ascending
- `{prefix}sort-desc={column}`: Sort column descending

**URL Examples:**
- `/users?sort-desc=Name` - Sort by Name descending
- `/page?users-sort-asc=Age&orders-sort-desc=Total` - Multiple tables

