package json

import (
	"strings"
	"testing"
	"unicode"
)

func TestParseJSONPath(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{path: "$", expected: []string{"$"}},
		{path: "$.", expected: []string{"$"}},
		{path: "$..", expected: []string{"$", ".."}},
		{path: "$.*", expected: []string{"$", "*"}},
		{path: "$..*", expected: []string{"$", "..", "*"}},
		{path: "$.root.element", expected: []string{"$", "root", "element"}},
		{path: "$.root.*.element", expected: []string{"$", "root", "*", "element"}},
		{path: "$['root']['element']", expected: []string{"$", "root", "element"}},
		{path: "$['root'][*]['element']", expected: []string{"$", "root", "*", "element"}},
		{path: "$['store']['book'][0]['title']", expected: []string{"$", "store", "book", "0", "title"}},
		{path: "$['root'].*['element']", expected: []string{"$", "root", "*", "element"}},
		{path: "$.['root'].*.['element']", expected: []string{"$", "root", "*", "element"}},
		{path: "$['root'].*.['element']", expected: []string{"$", "root", "*", "element"}},
		{path: "$.phoneNumbers[*].type", expected: []string{"$", "phoneNumbers", "*", "type"}},
		// TODO: support filter expressions
		// {path: "$.store.book[?(@.price < 10)].title", expected: []string{"$", "store", "book", "?(@.price < 10)", "title"}},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result, err := parsePath(tt.path)
			if err != nil {
				t.Errorf("error on path %s: %s", tt.path, err.Error())
			} else if !sliceEqual(result, tt.expected) {
				t.Errorf("expected %s, got %s", sliceString(tt.expected), sliceString(result))
			}
		})
	}
}

func TestJsonPath(t *testing.T) {
	// JSON from: https://support.smartbear.com/alertsite/docs/monitors/api/endpoint/jsonpath.html
	data := []byte(`{
		"store": {
			"book": [
			{
				"category": "reference",
				"author": "Nigel Rees",
				"title": "Sayings of the Century",
				"price": 8.95
			},
			{
				"category": "fiction",
				"author": "Herman Melville",
				"title": "Moby Dick",
				"isbn": "0-553-21311-3",
				"price": 8.99
			},
			{
				"category": "fiction",
				"author": "J.R.R. Tolkien",
				"title": "The Lord of the Rings",
				"isbn": "0-395-19395-8",
				"price": 22.99
			}
			],
			"bicycle": {
			"color": "red",
			"price": 19.95
			}
		},
		"expensive": 10
	}`)

	tests := []struct {
		path     string
		expected []string
	}{
		{"$.store.*", []string{
			`{
   "color": "red",
   "price": 19.95
}`,
			`[
  {
   "category": "reference",
   "author": "Nigel Rees",
   "title": "Sayings of the Century",
   "price": 8.95
  },
  {
   "category": "fiction",
   "author": "Herman Melville",
   "title": "Moby Dick",
   "isbn": "0-553-21311-3",
   "price": 8.99
  },
  {
   "category": "fiction",
   "author": "J.R.R. Tolkien",
   "title": "The Lord of the Rings",
   "isbn": "0-395-19395-8",
   "price": 22.99
  }
]`,
		}},
		{"$.store.bicycle.color", []string{`"red"`}},
		{"$.store.book[*]", []string{
			`{
   "category": "reference",
   "author": "Nigel Rees",
   "title": "Sayings of the Century",
   "price": 8.95
}`,
			`{
   "category": "fiction",
   "author": "Herman Melville",
   "title": "Moby Dick",
   "isbn": "0-553-21311-3",
   "price": 8.99
}`,
			`{
   "category": "fiction",
   "author": "J.R.R. Tolkien",
   "title": "The Lord of the Rings",
   "isbn": "0-395-19395-8",
   "price": 22.99
}`,
		}},
		{"$.store.book[0].title", []string{`"Sayings of the Century"`}},
		{"$.store..price", []string{`19.95`, `8.95`, `8.99`, `22.99`}},
		{"$..price", []string{`19.95`, `8.95`, `8.99`, `22.99`}},
		{"$..book[*].title", []string{`"Sayings of the Century"`, `"Moby Dick"`, `"The Lord of the Rings"`}},
		{"$..book[0]", []string{
			`{
   "category": "reference",
   "author": "Nigel Rees",
   "title": "Sayings of the Century",
   "price": 8.95
}`,
		}},
	}

	for _, tt := range tests {
		result, err := Path(data, tt.path)
		if err != nil {
			t.Errorf("Unexpected error for path %q: %v", tt.path, err)
			continue
		}

		if len(result) != len(tt.expected) {
			t.Errorf("Path %q: expected %d results, got %d", tt.path, len(tt.expected), len(result))
			continue
		}

		for i, node := range result {
			expectedNorm := normalizeJSON(tt.expected[i])
			resultNorm := normalizeJSON(node.String())
			if resultNorm != expectedNorm {
				t.Errorf("Path %q: expected result %q, got %q", tt.path, expectedNorm, resultNorm)
			}
		}
	}
}

// normalizeJSON removes all whitespace outside of quoted text.
func normalizeJSON(s string) string {
	var sb strings.Builder
	inQuotes := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuotes = !inQuotes
			sb.WriteByte(c)
		} else if inQuotes {
			sb.WriteByte(c)
		} else if !inQuotes && !unicode.IsSpace(rune(c)) {
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func sliceString(array []string) string {
	return "[" + strings.Join(array, ", ") + "]"
}
