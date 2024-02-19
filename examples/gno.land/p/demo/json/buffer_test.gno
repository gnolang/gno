package json

import (
	"testing"
)

func TestBufferCurrent(t *testing.T) {
	tests := []struct {
		name string
		buffer *buffer
		expected byte
		wantErr bool
	}{
		{
			name: "Valid current byte",
			buffer: &buffer{
				data: []byte("test"),
				length: 4,
				index: 1,
			},
			expected: 'e',
			wantErr: false,
		},
		{
			name: "EOF",
			buffer: &buffer{
				data: []byte("test"),
				length: 4,
				index: 4,
			},
			expected: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func (t *testing.T) {
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
		pos     uint
		want    []byte
		wantErr bool
	}{
		{
			name: "Valid slice -- 0 characters",
			buffer: &buffer{data: []byte("test"), length: 4, index: 0},
			pos: 0,
			want: nil,
			wantErr: false,
		},
		{
			name:   "Valid slice -- 1 character",
			buffer: &buffer{data: []byte("test"), length: 4, index: 0},
			pos:    1,
			want:   []byte("t"),
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
		pos     uint
		wantErr bool
		wantIdx uint
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
        expected uint
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
            b := NewBuffer(tt.input)
            b.skipToNextSignificantToken()
            if b.index != tt.expected {
                t.Errorf("after skipToNextSignificantToken(), got index = %v, want %v", b.index, tt.expected)
            }
        })
    }
}

const testString = "abcdefg"

func mockBuffer(s string) *buffer {
    return NewBuffer([]byte(s))
}

func TestSkipAndReturnIndex(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected uint
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
        expected uint
    }{
        {"SkipToToken", "abcdefg", map[byte]bool{'c': true}, 2},
        {"SkipToEnd", "abcdefg", map[byte]bool{'h': true}, 7},
        {"SkipNone", "abcdefg", map[byte]bool{'a': true}, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            buf := mockBuffer(tt.input)
            got, err := buf.skipUntil(tt.tokens)
            if err != nil && got != uint(len(tt.input)) { // Expect error only if reached end without finding token
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
        start    uint
        end      uint
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
