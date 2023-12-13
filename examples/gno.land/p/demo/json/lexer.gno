package json

import "errors"

const unescapeStackBufSize = 64

type Lexer struct {
	data     []byte
	token    byte
	level    int
	keyLevel int
}

func New(data []byte) *Lexer {
	return &Lexer{
		data:     data,
		level:    0,
		keyLevel: 0,
	}
}

func (l *Lexer) findTokenStart(token byte) int {
	for i := len(l.data) - 1; i >= 0; i-- {
		switch l.data[i] {
		case token:
			return i
		case '[', '{':
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
			continue
		case '}', ']':
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

// TODO: refactor this function
func (l *Lexer) stringEnd(data []byte) (lastCharIdx int, escaped bool, err error) {
	escaped = false
	for lastCharIdx, c := range data {
		if c == '"' {
			if !escaped {
				return lastCharIdx + 1, false, nil
			}

			prevCharIdx := lastCharIdx - 1
			for prevCharIdx >= 0 && data[prevCharIdx] == '\\' {
				prevCharIdx--
			}

			if (lastCharIdx-prevCharIdx)%2 == 0 {
				return lastCharIdx + 1, true, nil
			}

			return lastCharIdx + 1, false, nil
		}

		if c == '\\' {
			escaped = true
		}
	}

	return -1, escaped, StringNotClosed
}

func (l *Lexer) blockEnd(data []byte, open, close byte) (int, error) {
	level := l.level
	i := 0

	for i < len(data) {
		switch data[i] {
		case '"':
			end, _, err := l.stringEnd(data[i+1:])
			if err != nil {
				return -1, StringNotClosed
			}

			i += end
		case open:
			level++
		case close:
			level--

			if level == 0 {
				return i + 1, nil
			}
		}

		i++
	}

	return -1, BlockIsNotClosed
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

// TODO: refactor this function
func (l *Lexer) searchKeys(data []byte, keys ...string) (keyIndex int, err error) {
	lastMatched := true
	keyLevel, level := l.keyLevel, l.level
	dataLenghth, numKeys := len(data), len(keys)

	if numKeys == 0 {
		return 0, nil
	}

	var stackbuf [unescapeStackBufSize]byte

	for keyIndex < dataLenghth {
		switch data[keyIndex] {
		case '"':
			keyIndex++
			keyBegin := keyIndex

			strEnd, keyEscaped, err := l.stringEnd(data[keyIndex:])
			if err != nil {
				return -1, StringNotClosed
			}
			keyIndex += strEnd
			keyEnd := keyIndex - 1

			valueOffset := l.nextToken(data[keyIndex:])
			if valueOffset == -1 {
				return -1, TokenNotFound
			}

			keyIndex += valueOffset
			if data[keyIndex] != ':' {
				keyIndex--
			}

			if level < 1 {
				return -1, OutOfLevel
			}

			key := data[keyBegin:keyEnd]
			keyUnesc, err := Unescape(key, stackbuf[:])
			if err != nil {
				return -1, err
			}

			if !keyEscaped {
				keyUnesc = key
			}

			if level > numKeys {
				return -1, KeyPathNotFoundError
			}

			if string(keyUnesc) == keys[level-1] {
				lastMatched = true
				if keyLevel == level-1 {
					keyLevel++

					if keyLevel == numKeys {
						return keyIndex + 1, nil
					}
				}
			}

			lastMatched = false

		case '{':
			// in case parent key is matched then only we will increase the level otherwise can directly
			// can move to the end of this block
			if !lastMatched {
				end, err := l.blockEnd(data[keyIndex:], '{', '}')
				if err != nil {
					return -1, err
				}
				keyIndex += end - 1
			} else {
				level++
			}

		case '}':
			level--
			if level == keyLevel {
				keyLevel--
			}
		case '[':
			panic(errors.New("arrays are not supported"))
		case ':':
			return -1, MalformedJSON
		}

		keyIndex++
	}

	return -1, KeyPathNotFoundError
}
