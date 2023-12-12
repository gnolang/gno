package json

import (
	"errors"
)

const (
	KeyPathNotFoundError error = errors.New("key path not found")
	NotSpecifiedKeyError error = errors.New("key not specified")
	ClosingBracketError  error = errors.New("closing bracket not found")
	ClosingQuoteError    error = errors.New("closing quote not found")
)

const unescapeStackBufSize = 64

type Lexer struct {
	data  []byte
	token byte
	pos   int
	depth int
}

func New(data []byte) *Lexer {
	return &Lexer{
		data: data,
	}
}

func (l *Lexer) findTokenStart(token byte) int {
	for i := len(l.data) - 1; i >= 0; i-- {
		switch l.data[i] {
		case token:
			return i
		case '[', '{':
			l.depth++
			return 0
		}
	}

	return 0
}

func (l *Lexer) tokenEnd() int {
	for i, tok := range l.data {
		switch tok {
		case ' ', '\n', '\r', '\t', ',', '}', ']':
			return i
		}
	}

	return len(l.data)
}

func (l *Lexer) tokenStart() int {
	for i := len(l.data) - 1; i >= 0; i-- {
		switch l.data[i] {
		case '\n', '\r', '\t', ',', '{', '[':
			return i
		}
	}

	return 0
}

func (l *Lexer) nextToken(data []byte) int {
	for i, b := range data {
		switch b {
		case '{', '[':
			l.depth++
			continue
		case '}', ']':
			l.depth--
			continue
		default:
			return i
		}
	}

	return -1
}

func (l *Lexer) findLastTokenPosition() int {
	for i := len(l.data) - 1; i >= 0; i-- {
		switch l.data[i] {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}

	return -1
}

func (l *Lexer) stringEnd(data []byte) (idx int, escaped bool) {
	backSlashes := 0
	escaped = isEven(backSlashes)

	for idx, b := range data {
		if b == '\\' {
			backSlashes++
			continue
		}

		if b != '"' {
			backSlashes = 0
			continue
		}

		return idx + 1, !isEven(backSlashes)
	}

	return -1, escaped
}

func (l *Lexer) blockEnd(data []byte, open, close byte) int {
	level := 0
	i := 0

	for i < len(data) {
		switch data[i] {
		case '"':
			end, _ := l.stringEnd(data[i+1:])
			if end == -1 {
				return -1
			}

			i += end
		case open:
			level++
		case close:
			level--

			if level == 0 {
				return i + 1
			}
		}

		i++
	}

	return -1
}

// findKeyPositionRange identifies the start and end positions of a key within the given data,
// including positions enclosed in double quotes (`"`).
func (l *Lexer) findKeyPositionRange(data []byte, key string) (start, end int, err error) {
	keyLen := len(key)
	if keyLen == 0 {
		return -1, -1, NotSpecifiedKeyError
	}

	keyLastPos, err := searchKey(data, key, 0)
	if err != nil {
		return -1, -1, err
	}

	start = keyLastPos - keyLen - 1
	end = keyLastPos

	return start, end, nil
}

func searchKey(data []byte, key string, offset int) (int, error) {
	for i := offset; i < len(data); i++ {
		switch data[i] {
		case '"':
			begin, end, nextIndex, err := findNextKey(data, i)
			if err != nil {
				return -1, err
			}

			if string(data[begin:end]) == key {
				return nextIndex, nil
			}

			i = nextIndex
		}
	}

	return -1, KeyPathNotFoundError
}

func findNextKey(data []byte, start int) (keyBegin, keyEnd, nextIndex int, err error) {
	keyBegin = -1
	keyEnd = -1
	nextIndex = -1

	for i := start; i < len(data); i++ {
		switch data[i] {
		case '"':
			if keyBegin != -1 {
				keyEnd = i
				nextIndex = i + 1
				return keyBegin, keyEnd, nextIndex, nil
			}

			keyBegin = i + 1
		case '\\':
			i++
		}
	}

	return keyBegin, keyEnd, nextIndex, KeyPathNotFoundError
}
