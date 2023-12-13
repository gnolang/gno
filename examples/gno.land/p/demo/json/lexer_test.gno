package json

import (
	"errors"
	"testing"
)

type lexerTest struct {
	name     string
	data     []byte
	token    byte
	escaped  bool
	expected int
}

func TestFindTokenStart(t *testing.T) {
	tests := []lexerTest{
		{token: '{', expected: 5},
		{token: '[', expected: 0},
		{token: 'w', expected: 6},
		{token: 'l', expected: 9},
	}

	for _, test := range tests {
		l := New([]byte("hello{world}"))
		if got := l.findTokenStart(test.token); got != test.expected {
			t.Errorf("findTokenStart('%c') = %v, want %v", test.token, got, test.expected)
		}
	}
}

func TestTokenEnd(t *testing.T) {
	tests := []lexerTest{
		{name: "whitespace", data: []byte("hello world"), expected: 5},
		{name: "newline", data: []byte("hello\nworld"), expected: 5},
		{name: "tab", data: []byte("hello\tworld"), expected: 5},
		{name: "multiple words", data: []byte("hello world foo bar"), expected: 5},
		{name: "no space", data: []byte("helloworld"), expected: 10},
	}

	for _, test := range tests {
		if got := tokenEnd(test.data); got != test.expected {
			t.Errorf("tokenEnd() = %v, want %v", got, test.expected)
		}
	}
}

func TestLexer(t *testing.T) {
	tests := []struct {
		name  string
		data  string
		start int
		next  int
		last  int
	}{
		{
			name:  "Test 1",
			data:  "{\n\t\"key\": \"value\",\n\t\"array\": [1, 2, 3]\n}",
			start: 38,
			next:  1,
			last:  39,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lexer{data: []byte(tt.data)}

			if got := l.findLastTokenPosition(); got != tt.last {
				t.Errorf("findLastTokenPosition() = %v, want %v", got, tt.last)
			}
		})
	}
}


func TestSearchKeys(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		keys     []string
		expected int
		isErr    bool
	}{
		{
			name:     "Empty data and keys",
			data:     []byte{},
			keys:     []string{},
			expected: 0,
		},
		{
			name:     "Empty data with keys",
			data:     []byte{},
			keys:     []string{"key1", "key2"},
			expected: -1,
			isErr:    true,
		},
		{
			name:     "Data with no keys",
			data:     []byte(`{"foo": "bar"}`),
			keys:     []string{"key1", "key2"},
			expected: -1,
			isErr:    true,
		},
		{
			name:     "Data with single key",
			data:     []byte(`{"key1": "value1"}`),
			keys:     []string{"key1"},
			expected: 8,
		},
		{
			name:     "Data with multiple keys",
			data:     []byte(`{"key1": "value1", "key2": "value2"}`),
			keys:     []string{"key2"},
			expected: 26,
		},
		{
			name:     "Data with nested keys",
			data:     []byte(`{"outer": {"key1": "value1", "key2": "value2"}, "key3": "value3"}`),
			keys:     []string{"key3"},
			expected: 55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lexer{}
			got, err := l.searchKeys(tt.data, tt.keys...)

			if got != tt.expected {
				t.Errorf("searchKeys(%s) = %v, want %v", tt.name, got, tt.expected)
			}

			if (err != nil) != tt.isErr {
				t.Errorf("searchKeys(%s) error = %v, wantErr %v", tt.name, err, tt.isErr)
			}
		})
	}
}

func TestArrayEach(t *testing.T) {
	data := []byte(`{"a": { "b":[{"x": 1} ,{"x":2},{ "x":3}, {"x":4} ]}}`)
	count := 0

	l := &Lexer{}
	l.ArrayEach(data, func(value []byte, typ ValueType, offset int, err error) {
		count += 1

		switch count {
		case 1:
			if string(value) != `{"x": 1}` {
				t.Errorf("Wrong first item: %s", string(value))
			}
		case 2:
			if string(value) != `{"x":2}` {
				t.Errorf("Wrong second item: %s", string(value))
			}
		case 3:
			if string(value) != `{ "x":3}` {
				t.Errorf("Wrong third item: %s", string(value))
			}
		case 4:
			if string(value) != `{"x":4}` {
				t.Errorf("Wrong fourth item: %s", string(value))
			}
		default:
			t.Errorf("Too many items: %d", count)
		}
	}, "a", "b")
}