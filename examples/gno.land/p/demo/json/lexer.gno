package json

import (
	"bytes"
	"errors"
	"strconv"
)

func findTokenStart(data []byte, token byte) int {
	for i := len(data) - 1; i >= 0; i-- {
		switch data[i] {
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

func nextToken(data []byte) (int, error) {
	for i, b := range data {
		switch b {
		case '{', '[':
			continue
		case '}', ']':
			continue
		default:
			return i, nil
		}
	}

	return -1, TokenNotFound
}

func findLastTokenPosition(data []byte) int {
	for i := len(data) - 1; i >= 0; i-- {
		switch data[i] {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}

	return -1
}

// TODO: refactor this function
func stringEnd(data []byte) (lastCharIdx int, escaped bool, err error) {
	escaped = false
	for lastCharIdx, c := range data {
		switch c {
		case '"':
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
		case '\\':
			escaped = true
		}
	}

	return -1, escaped, StringNotClosed
}

func blockEnd(data []byte, open, close byte) (int, error) {
	level := 0
	i := 0

	for i < len(data) {
		switch data[i] {
		case '"':
			end, _, err := stringEnd(data[i+1:])
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
func searchKeys(data []byte, keys ...string) (keyIndex int, err error) {
	lastMatched := true
	keyLevel, level := 0, 0
	dataLenghth, numKeys := len(data), len(keys)

	if numKeys == 0 {
		return 0, nil
	}

	var stackbuf [UnescapeStackBufSize]byte

	for keyIndex < dataLenghth {
		switch data[keyIndex] {
		case '"':
			keyIndex++
			keyBegin := keyIndex

			strEnd, keyEscaped, err := stringEnd(data[keyIndex:])
			if err != nil {
				return -1, StringNotClosed
			}
			keyIndex += strEnd
			keyEnd := keyIndex - 1

			valueOffset, err := nextToken(data[keyIndex:])
			if err != nil {
				return -1, err
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

			if !bytes.Equal(keyUnesc, []byte(keys[level-1])) || keyLevel+1 != numKeys {
				lastMatched = false
				break
			}

			if keyLevel+1 == numKeys {
				return keyIndex + 1, nil
			}

			lastMatched = true

		case '{':
			if !lastMatched {
				end, err := blockEnd(data[keyIndex:], '{', '}')
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

				var valueFound []byte
				var valueOffset int

				currIdx := keyIndex
				ArrayEach(data[keyIndex:], func(value []byte, dataType ValueType, offset int, err error) {
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

				subIndex, err := searchKeys(valueFound, keys[level+1:]...)
				if err != nil {
					return -1, err
				}

				return keyIndex + valueOffset + subIndex, nil
			}

			skip, err := blockEnd(data[keyIndex:], '[', ']')
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
func ArrayEach(
	data []byte,
	f func(value []byte, dataType ValueType, offset int, err error),
	keys ...string,
) (offset int, err error) {
	if len(data) == 0 {
		return -1, MalformedObjectError
	}

	next, err := nextToken(data)
	if err != nil {
		return -1, MalformedJsonError
	}

	offset = next + 1

	if len(keys) > 0 {
		offset, err = searchKeys(data, keys...)
		if err != nil {
			return offset, err
		}

		nextOffset, err := nextToken(data[offset:])
		if err != nil {
			return offset, MalformedJsonError
		}

		offset += nextOffset

		if data[offset] != '[' {
			return offset, MalformedArrayError
		}

		offset++
	}

	nextOffset, err := nextToken(data[offset:])
	if err != nil {
		return offset, MalformedJsonError
	}

	offset += nextOffset

	if data[offset] == ']' {
		return offset, nil
	}

	for true {
		offset, err := handleGetResult(data, offset, f)
		if err != nil {
			return offset, err
		}

		offset, err = handleNextTokenResult(data, offset)
		if err != nil {
			return offset, err
		}

		offset, err = handleArrayError(data, offset)
		if err != nil {
			return offset, err
		}
	}

	return offset, nil
}

func Get(data []byte, keys ...string) (value []byte, dataType ValueType, offset int, err error) {
	a, b, _, d, e := internalGet(data, keys...)
	return a, b, d, e
}

func internalGet(data []byte, keys ...string) (value []byte, dataType ValueType, offset, endOffset int, err error) {
	if len(keys) <= 0 {
		return nil, NotExist, -1, -1, errors.New("no keys")
	}

	offset, err = searchKeys(data, keys...)
	if err != nil {
		return nil, NotExist, -1, -1, err
	}

	// Go to closest value
	nextOffset, err := nextToken(data[offset:])
	if err != nil {
		return nil, NotExist, offset, -1, MalformedJsonError
	}

	offset += nextOffset
	value, dataType, endOffset, err = getType(data, offset)
	if err != nil {
		return value, dataType, offset, endOffset, err
	}

	// Strip quotes from string values
	if dataType == String {
		value = value[1 : len(value)-1]
	}

	return value[:len(value):len(value)], dataType, offset, endOffset, nil
}

func getType(data []byte, offset int) ([]byte, ValueType, int, error) {
	var dataType ValueType
	endOffset := offset

	switch data[offset] {
	case '"':
		dataType = String
		idx, _, err := stringEnd(data[offset+1:])
		if err != nil {
			return nil, dataType, offset, MalformedStringError
		}
		endOffset += idx + 1
	case '[':
		return processNextCollectionElement(data, offset, Array)
	case '{':
		return processNextCollectionElement(data, offset, Object)
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

func handleGetResult(
	data []byte,
	offset int,
	f func(value []byte, dataType ValueType, offset int, err error),
) (int, error) {
	v, t, o, e := Get(data[offset:])
	if e != nil {
		return offset, e
	}

	if o == 0 {
		return offset, nil
	}

	if t != NotExist {
		f(v, t, offset+o-len(v), e)
	}

	if e != nil {
		return offset, e
	}

	return offset + o, nil
}

func handleNextTokenResult(data []byte, offset int) (int, error) {
	nextTokenOffset, err := nextToken(data[offset:])
	if err != nil {
		return offset, MalformedArrayError
	}

	return offset + nextTokenOffset, nil
}

func handleArrayError(data []byte, offset int) (int, error) {
	if data[offset] == ']' {
		return offset, nil
	}

	if data[offset] != ',' {
		return offset, MalformedArrayError
	}

	return offset + 1, nil
}

func processNextCollectionElement(data []byte, offset int, dataType ValueType) ([]byte, ValueType, int, error) {
	switch dataType {
	case Array:
		endOffset, err := blockEnd(data[offset:], '[', ']')
		if err != nil {
			return nil, NotExist, offset, MalformedArrayError
		}

		return data[offset : offset+endOffset], Array, offset + endOffset, nil
	case Object:
		endOffset, err := blockEnd(data[offset:], '{', '}')
		if err != nil {
			return nil, NotExist, offset, MalformedObjectError
		}

		return data[offset : offset+endOffset], Object, offset + endOffset, nil
	default:
		return nil, NotExist, offset, errors.New("not a collection")
	}
}
