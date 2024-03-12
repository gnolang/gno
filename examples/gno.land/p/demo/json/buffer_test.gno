package json

import (
	"testing"
)

func TestBufferCurrent(t *testing.T) {
	tests := []struct {
		name     string
		buffer   *buffer
		expected byte
		wantErr  bool
	}{
		{
			name: "Valid current byte",
			buffer: &buffer{
				data:   []byte("test"),
				length: 4,
				index:  1,
			},
			expected: 'e',
			wantErr:  false,
		},
		{
			name: "EOF",
			buffer: &buffer{
				data:   []byte("test"),
				length: 4,
				index:  4,
			},
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.buffer.current()
			if (err != nil) != tt.wantErr {
				t.Errorf("buffer.current() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("buffer.current() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBufferStep(t *testing.T) {
	tests := []struct {
		name    string
		buffer  *buffer
		wantErr bool
	}{
		{
			name:    "Valid step",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			wantErr: false,
		},
		{
			name:    "EOF error",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 3},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.buffer.step()
			if (err != nil) != tt.wantErr {
				t.Errorf("buffer.step() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBufferNext(t *testing.T) {
	tests := []struct {
		name    string
		buffer  *buffer
		want    byte
		wantErr bool
	}{
		{
			name:    "Valid next byte",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			want:    'e',
			wantErr: false,
		},
		{
			name:    "EOF error",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 3},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.buffer.next()
			if (err != nil) != tt.wantErr {
				t.Errorf("buffer.next() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buffer.next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBufferSlice(t *testing.T) {
	tests := []struct {
		name    string
		buffer  *buffer
		pos     int
		want    []byte
		wantErr bool
	}{
		{
			name:    "Valid slice -- 0 characters",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			pos:     0,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Valid slice -- 1 character",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			pos:     1,
			want:    []byte("t"),
			wantErr: false,
		},
		{
			name:    "Valid slice -- 2 characters",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 1},
			pos:     2,
			want:    []byte("es"),
			wantErr: false,
		},
		{
			name:    "Valid slice -- 3 characters",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			pos:     3,
			want:    []byte("tes"),
			wantErr: false,
		},
		{
			name:    "Valid slice -- 4 characters",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			pos:     4,
			want:    []byte("test"),
			wantErr: false,
		},
		{
			name:    "EOF error",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 3},
			pos:     2,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.buffer.slice(tt.pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("buffer.slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != string(tt.want) {
				t.Errorf("buffer.slice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBufferMove(t *testing.T) {
	tests := []struct {
		name    string
		buffer  *buffer
		pos     int
		wantErr bool
		wantIdx int
	}{
		{
			name:    "Valid move",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 1},
			pos:     2,
			wantErr: false,
			wantIdx: 3,
		},
		{
			name:    "Move beyond length",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 1},
			pos:     4,
			wantErr: true,
			wantIdx: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.buffer.move(tt.pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("buffer.move() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.buffer.index != tt.wantIdx {
				t.Errorf("buffer.move() index = %v, want %v", tt.buffer.index, tt.wantIdx)
			}
		})
	}
}

func TestBufferSkip(t *testing.T) {
	tests := []struct {
		name    string
		buffer  *buffer
		b       byte
		wantErr bool
	}{
		{
			name:    "Skip byte",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			b:       'e',
			wantErr: false,
		},
		{
			name:    "Skip to EOF",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			b:       'x',
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.buffer.skip(tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("buffer.skip() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBufferSkipAny(t *testing.T) {
	tests := []struct {
		name    string
		buffer  *buffer
		s       map[byte]bool
		wantErr bool
	}{
		{
			name:    "Skip any valid byte",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			s:       map[byte]bool{'e': true, 'o': true},
			wantErr: false,
		},
		{
			name:    "Skip any to EOF",
			buffer:  &buffer{data: []byte("test"), length: 4, index: 0},
			s:       map[byte]bool{'x': true, 'y': true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.buffer.skipAny(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("buffer.skipAny() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSkipToNextSignificantToken(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int
	}{
		{"No significant chars", []byte("abc"), 3},
		{"One significant char at start", []byte(".abc"), 0},
		{"Significant char in middle", []byte("ab.c"), 2},
		{"Multiple significant chars", []byte("a$.c"), 1},
		{"Significant char at end", []byte("abc$"), 3},
		{"Only significant chars", []byte("$."), 0},
		{"Empty string", []byte(""), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBuffer(tt.input)
			b.skipToNextSignificantToken()
			if b.index != tt.expected {
				t.Errorf("after skipToNextSignificantToken(), got index = %v, want %v", b.index, tt.expected)
			}
		})
	}
}

func mockBuffer(s string) *buffer {
	return newBuffer([]byte(s))
}

func TestSkipAndReturnIndex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"StartOfString", "", 0},
		{"MiddleOfString", "abcdef", 1},
		{"EndOfString", "abc", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := mockBuffer(tt.input)
			got, err := buf.skipAndReturnIndex()
			if err != nil && tt.input != "" { // Expect no error unless input is empty
				t.Errorf("skipAndReturnIndex() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("skipAndReturnIndex() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSkipUntil(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tokens   map[byte]bool
		expected int
	}{
		{"SkipToToken", "abcdefg", map[byte]bool{'c': true}, 2},
		{"SkipToEnd", "abcdefg", map[byte]bool{'h': true}, 7},
		{"SkipNone", "abcdefg", map[byte]bool{'a': true}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := mockBuffer(tt.input)
			got, err := buf.skipUntil(tt.tokens)
			if err != nil && got != len(tt.input) { // Expect error only if reached end without finding token
				t.Errorf("skipUntil() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("skipUntil() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSliceFromIndices(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		end      int
		expected string
	}{
		{"FullString", "abcdefg", 0, 7, "abcdefg"},
		{"Substring", "abcdefg", 2, 5, "cde"},
		{"OutOfBounds", "abcdefg", 5, 10, "fg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := mockBuffer(tt.input)
			got := buf.sliceFromIndices(tt.start, tt.end)
			if string(got) != tt.expected {
				t.Errorf("sliceFromIndices() = %v, want %v", string(got), tt.expected)
			}
		})
	}
}

func TestBufferToken(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		index int
		isErr bool
	}{
		{
			name:  "Simple valid path",
			path:  "@.length",
			index: 8,
			isErr: false,
		},
		{
			name:  "Path with array expr",
			path:  "@['foo'].0.bar",
			index: 14,
			isErr: false,
		},
		{
			name:  "Path with array expr and simple fomula",
			path:  "@['foo'].[(@.length - 1)].*",
			index: 27,
			isErr: false,
		},
		{
			name:  "Path with filter expr",
			path:  "@['foo'].[?(@.bar == 1 & @.baz < @.length)].*",
			index: 45,
			isErr: false,
		},
		{
			name:  "addition of foo and bar",
			path:  "@.foo+@.bar",
			index: 11,
			isErr: false,
		},
		{
			name:  "logical AND of foo and bar",
			path:  "@.foo && @.bar",
			index: 14,
			isErr: false,
		},
		{
			name:  "logical OR of foo and bar",
			path:  "@.foo || @.bar",
			index: 14,
			isErr: false,
		},
		{
			name:  "accessing third element of foo",
			path:  "@.foo,3",
			index: 7,
			isErr: false,
		},
		{
			name:  "accessing last element of array",
			path:  "@.length-1",
			index: 10,
			isErr: false,
		},
		{
			name:  "number 1",
			path:  "1",
			index: 1,
			isErr: false,
		},
		{
			name:  "float",
			path:  "3.1e4",
			index: 5,
			isErr: false,
		},
		{
			name:  "float with minus",
			path:  "3.1e-4",
			index: 6,
			isErr: false,
		},
		{
			name:  "float with plus",
			path:  "3.1e+4",
			index: 6,
			isErr: false,
		},
		{
			name:  "negative number",
			path:  "-12345",
			index: 6,
			isErr: false,
		},
		{
			name:  "negative float",
			path:  "-3.1e4",
			index: 6,
			isErr: false,
		},
		{
			name:  "negative float with minus",
			path:  "-3.1e-4",
			index: 7,
			isErr: false,
		},
		{
			name:  "negative float with plus",
			path:  "-3.1e+4",
			index: 7,
			isErr: false,
		},
		{
			name:  "string number",
			path:  "'12345'",
			index: 7,
			isErr: false,
		},
		{
			name:  "string with backslash",
			path:  "'foo \\'bar '",
			index: 12,
			isErr: false,
		},
		{
			name:  "string with inner double quotes",
			path:  "'foo \"bar \"'",
			index: 12,
			isErr: false,
		},
		{
			name:  "parenthesis 1",
			path:  "(@abc)",
			index: 6,
			isErr: false,
		},
		{
			name:  "parenthesis 2",
			path:  "[()]",
			index: 4,
			isErr: false,
		},
		{
			name:  "parenthesis mismatch",
			path:  "[(])",
			index: 2,
			isErr: true,
		},
		{
			name:  "parenthesis mismatch 2",
			path:  "(",
			index: 1,
			isErr: true,
		},
		{
			name:  "parenthesis mismatch 3",
			path:  "())]",
			index: 2,
			isErr: true,
		},
		{
			name:  "bracket mismatch",
			path:  "[()",
			index: 3,
			isErr: true,
		},
		{
			name:  "bracket mismatch 2",
			path:  "()]",
			index: 2,
			isErr: true,
		},
		{
			name:  "path does not close bracket",
			path:  "@.foo[)",
			index: 6,
			isErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := newBuffer([]byte(tt.path))

			err := buf.pathToken()
			if tt.isErr {
				if err == nil {
					t.Errorf("Expected an error for path `%s`, but got none", tt.path)
				}
			}

			if err == nil && tt.isErr {
				t.Errorf("Expected an error for path `%s`, but got none", tt.path)
			}

			if buf.index != tt.index {
				t.Errorf("Expected final index %d, got %d (token: `%s`) for path `%s`", tt.index, buf.index, string(buf.data[buf.index]), tt.path)
			}
		})
	}
}

func TestBufferFirst(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected byte
	}{
		{
			name:     "Valid first byte",
			data:     []byte("test"),
			expected: 't',
		},
		{
			name:     "Empty buffer",
			data:     []byte(""),
			expected: 0,
		},
		{
			name:     "Whitespace buffer",
			data:     []byte("   "),
			expected: 0,
		},
		{
			name:     "whitespace in middle",
			data:     []byte("hello world"),
			expected: 'h',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBuffer(tt.data)

			got, err := b.first()
			if err != nil && tt.expected != 0 {
				t.Errorf("Unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("Expected first byte to be %q, got %q", tt.expected, got)
			}
		})
	}
}
