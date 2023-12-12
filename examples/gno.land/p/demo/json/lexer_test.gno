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
		l := New(test.data)
		if got := l.tokenEnd(); got != test.expected {
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

			if got := l.tokenStart(); got != tt.start {
				t.Errorf("tokenStart() = %v, want %v", got, tt.start)
			}

			if got := l.findLastTokenPosition(); got != tt.last {
				t.Errorf("findLastTokenPosition() = %v, want %v", got, tt.last)
			}
		})
	}
}

func TestFindKeyPositionRange(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		key     string
		want    []int
		wantErr bool
	}{
		{
			name: "Simple key",
			data: `{"key1": "value1", "key2": "value2"}`,
			key:  "key1",
			want: []int{2, 7},
		},
		{
			name: "Nested key",
			data: `{"outer": {"key1": "value1"}, "key2": "value2"}`,
			key:  "key1",
			want: []int{12, 17},
		},
		{
			name: "Nested key 2",
			data: `{"outer": {"key1": "value1"}, "key2": "value2"}`,
			key:  "key2",
			want: []int{31, 36},
		},
		{
			name: "Nested key 3",
			data: `{\n"outer": {"key1": "value1"},\n"key2": "value2",\n"key3": "value3"\n}`,
			key:  "key3",
			want: []int{53, 58},
		},
		{
			name:    "Key not found",
			data:    `{"key1": "value1"}`,
			key:     "key2",
			want:    []int{-1, -1},
			wantErr: true,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lexer{data: []byte(tt.data)}
			key, err := l.findKeyPositionRange([]byte(tt.data), tt.key)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("findKeyPositionRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if key[0] != tt.want[0] || key[1] != tt.want[1] {
				t.Errorf("findKeyPositionRange() = %v, want %v", key, tt.want)
			}
		})
	}
}

func TestFindMultipleKeys(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		keys     []string
		expected []int
	}{
		{
			name:     "Empty data and keys",
			data:     []byte{},
			keys:     []string{},
			expected: []int{},
		},
		{
			name:     "Empty data with keys",
			data:     []byte{},
			keys:     []string{"key1", "key2"},
			expected: []int{},
		},
		{
			name:     "Data with no keys",
			data:     []byte(`{"foo": "bar"}`),
			keys:     []string{"key1", "key2"},
			expected: []int{},
		},
		{
			name:     "Data with single key",
			data:     []byte(`{"key1": "value1"}`),
			keys:     []string{"key1"},
			expected: []int{2, 7},
		},
		{
			name:     "Data with multiple keys",
			data:     []byte(`{"key1": "value1", "key2": "value2"}`),
			keys:     []string{"key1", "key2"},
			expected: []int{2, 7, 20, 25},
		},
		{
			name:     "Data with nested keys",
			data:     []byte(`[{"outer": {"key1": "value1"}, "key2": "value2"}]`),
			keys:     []string{"key1", "key2"},
			expected: []int{13, 18, 32, 37},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lexer{}
			got := l.findMultipleKeys(tt.data, tt.keys...)

			if len(got) != len(tt.expected) {
				t.Errorf("findMultipleKeys() = %v, want %v", got, tt.expected)
			}

			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("findMultipleKeys() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}