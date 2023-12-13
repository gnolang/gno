package json

import (
	"bytes"
	"errors"
	"strconv"
)

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

func tokenEnd(data []byte) int {
	for i, tok := range data {
		switch tok {
		case ' ', '\n', '\r', '\t', ',', '}', ']':
			return i
		}
	}

	return len(data)
}

// func (l *Lexer) tokenStart() int {
// 	for i := len(l.data) - 1; i >= 0; i-- {
// 		switch l.data[i] {
// 		case '\n', '\r', '\t', ',', '{', '[':
// 			return i
// 		}
// 	}

// 	return 0
// }

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
			decreaseLevel(level, keyLevel)
		case '[':
			if keyLevel == level && keys[level][0] == '[' {
				keyLen := len(keys[level])
				if keyLen < 3 || keys[level][0] != '[' || keys[level][keyLen-1] != ']' {
					return -1, MalformedArrayError
				}

				arrIdx, err := strconv.Atoi(keys[level][1 : keyLen-1])
				if err != nil {
					return -1, MalformedArrayError
				}

				var valueFound  []byte
				var valueOffset int

				currIdx := keyIndex
				l.ArrayEach(data[keyIndex:], func(value []byte, dataType ValueType, offset int, err error) {
					if arrIdx == currIdx {
						valueFound = value
						valueOffset = offset

						if dataType == String {
							valueOffset = valueOffset - 2
							start := currIdx + valueOffset
							end := start + len(value) + 2
							
							valueFound = valueFound[start:end]
						}
					}
					currIdx++
				})

				if valueFound == nil {
					return -1, KeyPathNotFoundError
				}

				subIndex, err := l.searchKeys(valueFound, keys[level+1:]...)
				if err != nil {
					return -1, err
				}

				return keyIndex + valueOffset + subIndex, nil
			}

			skip, err := l.blockEnd(data[keyIndex:], '[', ']')
			if err != nil {
				return -1, err
			}

			keyIndex += skip - 1
		case ':':
			return -1, MalformedJsonError
		}

		keyIndex++
	}

	return -1, KeyPathNotFoundError
}

func decreaseLevel(level, keyLevel int) {
	level--
	if level == keyLevel {
		keyLevel--
	}
}

// ArrayEach is used when iterating arrays, accepts a callback function with the same return arguments as `Get`.
func (l *Lexer) ArrayEach(data []byte, cb func(value []byte, dataType ValueType, offset int, err error), keys ...string) (offset int, err error) {
	if len(data) == 0 {
		return -1, MalformedObjectError
	}

	nT := l.nextToken(data)
	if nT == -1 {
		return -1, MalformedJsonError
	}

	offset = nT + 1

	if len(keys) > 0 {
		offset, err = l.searchKeys(data, keys...)
		if err != nil {
			return offset, err
		}

		// Go to closest value
		nO := l.nextToken(data[offset:])
		if nO == -1 {
			return offset, MalformedJsonError
		}

		offset += nO

		if data[offset] != '[' {
			return offset, MalformedArrayError
		}

		offset++
	}

	nO := l.nextToken(data[offset:])
	if nO == -1 {
		return offset, MalformedJsonError
	}

	offset += nO

	if data[offset] == ']' {
		return offset, nil
	}

	for true {
		v, t, o, e := l.Get(data[offset:])

		if e != nil {
			return offset, e
		}

		if o == 0 {
			break
		}

		if t != NotExist {
			cb(v, t, offset+o-len(v), e)
		}

		if e != nil {
			break
		}

		offset += o

		skipToToken := l.nextToken(data[offset:])
		if skipToToken == -1 {
			return offset, MalformedArrayError
		}
		offset += skipToToken

		if data[offset] == ']' {
			break
		}

		if data[offset] != ',' {
			return offset, MalformedArrayError
		}

		offset++
	}

	return offset, nil
}

func (l *Lexer) Get(data []byte, keys ...string) (value []byte, dataType ValueType, offset int, err error) {
	a, b, _, d, e := l.internalGet(data, keys...)
	return a, b, d, e
}

func (l *Lexer) internalGet(data []byte, keys ...string) (value []byte, dataType ValueType, offset, endOffset int, err error) {
	if len(keys) > 0 {
		offset, err = l.searchKeys(data, keys...)
		if err != nil {
			return nil, NotExist, -1, -1, err
		}
	}

	// Go to closest value
	nO := l.nextToken(data[offset:])
	if nO == -1 {
		return nil, NotExist, offset, -1, MalformedJsonError
	}

	offset += nO
	value, dataType, endOffset, err = l.getType(data, offset)
	if err != nil {
		return value, dataType, offset, endOffset, err
	}

	// Strip quotes from string values
	if dataType == String {
		value = value[1 : len(value)-1]
	}

	return value[:len(value):len(value)], dataType, offset, endOffset, nil
}

func (l *Lexer) getType(data []byte, offset int) ([]byte, ValueType, int, error) {
	var dataType ValueType
	endOffset := offset

	switch data[offset] {
	case '"':
		dataType = String
		idx, _, err := l.stringEnd(data[offset+1:])
		if err != nil {
			return nil, dataType, offset, MalformedStringError
		}
		endOffset += idx + 1
	case '[':
		dataType = Array
		// break label, for stopping nested loops
		endOffset, err := l.blockEnd(data[offset:], '[', ']')
		if err != nil {
			return nil, dataType, offset, MalformedArrayError
		}
		endOffset += offset
	case '{':
		dataType = Object

		endOffset, err := l.blockEnd(data[offset:], '{', '}')
		if err != nil {
			return nil, dataType, offset, MalformedObjectError
		}

		endOffset += offset
	default:
		end := tokenEnd(data[endOffset:])
		if end == -1 {
			return nil, dataType, offset, MalformedValueError
		}

		value := data[offset : endOffset+end]
		dataType, offset, err := extractValueTypeFromToken(data[offset], value)
		endOffset += end
	}

	return data[offset:endOffset], dataType, endOffset, nil
}
