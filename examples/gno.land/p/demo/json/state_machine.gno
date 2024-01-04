package json

import (
	"bytes"
	"errors"
	"strconv"
)

func extractValueTypeFromToken(b byte, value []byte) (dataType ValueType, offset int, err error) {
	switch b {
	case 't', 'f':
		if bytes.Equal(value, TrueLiteral) || bytes.Equal(value, FalseLiteral) {
			dataType = Boolean
		}
		return Unknown, offset, UnknownValueType
	case 'u', 'n':
		if bytes.Equal(value, NullLiteral) {
			dataType = Null
		}
		return Unknown, offset, UnknownValueType
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		dataType = Number
	default:
		return Unknown, offset, UnknownValueType
	}
}

// findValueIndex is a function that takes a byte slice of JSON data and a series of keys.
// It traverses the JSON data to find the position where the value of the final key in the series begins.
// The position is returned as an integer representing the index in the byte slice.
func findValueIndex(data []byte, keys ...string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	var (
		keyLevel    int
		level       int
		lastMatched bool = true
		stackbuf    [UnescapeStackBufSize]byte
	)

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case DoublyQuoteToken:
			i += 1
			keyBegin := i

			strEnd, keyEscaped := stringEnd(data[i:])
			if strEnd == -1 {
				return -1, MalformedJson
			}
			i += strEnd
			keyEnd := i - 1

			valueOffset, err := nextToken(data[i:])
			if err != nil {
				return -1, err
			}

			i += valueOffset
			if data[i] != ColonToken {
				i -= 1
				continue
			}

			if level < 1 {
				return -1, MalformedJson
			}

			key := data[keyBegin:keyEnd]
			keyUnesc, err := keyMatched(key, keyEscaped, stackbuf, keys, level)
			if err != nil {
				return -1, err
			}

			lastMatched = bytes.Equal(keyUnesc, []byte(keys[level-1]))
			if lastMatched && keyLevel == level-1 && keyLevel != len(keys) {
				keyLevel += 1
			}

			if keyLevel == len(keys) {
				return i + 1, nil
			}
		case CurlyOpenToken:
			// in case parent key is matched then only we will increase the level otherwise can directly
			// can move to the end of this block
			if !lastMatched {
				end := blockEnd(data[i:], CurlyOpenToken, CurlyCloseToken)
				if end == -1 {
					return -1, MalformedJson
				}
				i += end - 1
			} else {
				level += 1
			}
		case CurlyCloseToken:
			level, keyLevel = decreaseLevel(level, keyLevel)
		case SquareOpenToken:
			// if keyLevel == level && keys[level][0] == '[' {
			// 	keyLen := len(keys[level])
			// 	if keyLen < 3 || keys[level][0] != '[' || keys[level][keyLen-1] != ']' {
			// 		return -1, InvalidArrayIndex
			// 	}
			// }
			// var (
			// 	curIdx      int
			// 	valueFound  []byte
			// 	valueOffset int
			// )

			// isArrayKey, err := isArrayKey(level, keyLevel, keys)
			// if err != nil {
			// 	return -1, err
			// }

			// if !isArrayKey {
			// 	arraySkip := blockEnd(data[i:], SquareOpenToken, SquareCloseToken)
			// 	if arraySkip < 0 {
			// 		return -1, MalformedJson
			// 	}

			// 	i += arraySkip - 1
			// }

			// arrIdx, err := strconv.Atoi(keys[level][1 : len(keys[level])-1])
			// if err != nil {
			// 	return -1, InvalidArrayIndex
			// }

			// currIdx := i
			// ArrayEach(data[i:], func(value []byte, dataType ValueType, offset int, err error) {
			// 	if curIdx == arrIdx {
			// 		valueFound = value
			// 		valueOffset = offset
			// 		if dataType == String {
			// 			valueOffset = valueOffset - 2
			// 			valueFound = data[currIdx+valueOffset : currIdx+valueOffset+len(value)+2]
			// 		}
			// 	}
			// 	curIdx += 1
			// })

			// if valueFound == nil {
			// 	return -1, ArrayIndexNotFound
			// }

			// subIdx, err := findValueIndex(valueFound, keys[level+1:]...)
			// if err != nil {
			// 	return -1, err
			// }

			// return i + valueOffset + subIdx, nil
			// If we want to get array element by index
			if keyLevel == level && keys[level][0] == '[' {
				keyLen := len(keys[level])
				if keyLen < 3 || keys[level][0] != '[' || keys[level][keyLen-1] != ']' {
					return -1, InvalidArrayIndex
				}
				aIdx, err := strconv.Atoi(keys[level][1 : keyLen-1])
				if err != nil {
					return -1, InvalidArrayIndex
				}
				var curIdx int
				var valueFound []byte
				var valueOffset int
				curI := i
				ArrayEach(data[i:], func(value []byte, dataType ValueType, offset int, err error) {
					if curIdx == aIdx {
						valueFound = value
						valueOffset = offset
						if dataType == String {
							valueOffset = valueOffset - 2
							valueFound = data[curI+valueOffset : curI+valueOffset+len(value)+2]
						}
					}
					curIdx += 1
				})

				if valueFound == nil {
					return -1, ArrayIndexNotFound
				} else {
					subIndex, err := findValueIndex(valueFound, keys[level+1:]...)
					if err != nil {
						return -1, KeyPathNotFoundError
					}
					return i + valueOffset + subIndex, nil
				}
			} else {
				// Do not search for keys inside arrays
				if arraySkip := blockEnd(data[i:], '[', ']'); arraySkip == -1 {
					return -1, MalformedJson
				} else {
					i += arraySkip - 1
				}
			}
		case ColonToken: // If encountered, JSON data is malformed
			return -1, MalformedJson
		}
	}

	return -1, KeyPathNotFoundError
}

func decreaseLevel(level, keyLevel int) (int, int) {
	if level -= 1; level == keyLevel {
		keyLevel -= 1
	}

	return level, keyLevel
}

func findKeyStart(data []byte, key string) (int, error) {
	i, _ := nextToken(data)
	if i == -1 {
		return i, KeyPathNotFoundError
	}

	ln := len(data)
	if ln > 0 && (data[i] == CurlyOpenToken || data[i] == SquareOpenToken) {
		i += 1
	}

	var stackbuf [UnescapeStackBufSize]byte
	if ku, err := Unescape([]byte(key), stackbuf[:]); err == nil {
		key = string(ku)
	}

	for _; i < ln; i++ {
		switch data[i] {
		case DoublyQuoteToken:
			i++
			keyBegin := i

			strEnd, keyEscaped := stringEnd(data[i:])
			if strEnd == -1 {
				break
			}
			i += strEnd
			keyEnd := i - 1

			valueOffset, _ := nextToken(data[i:])
			if valueOffset == -1 {
				break
			}

			i += valueOffset

			// if string is a key, and key level match
			k := data[keyBegin:keyEnd]
			// for unescape: if there are no escape sequences, this is cheap; if there are, it is a
			// bit more expensive, but causes no allocations unless len(key) > unescapeStackBufSize
			if keyEscaped {
				if ku, err := Unescape(k, stackbuf[:]); err != nil {
					break
				} else {
					k = ku
				}
			}

			if data[i] == ColonToken && len(key) == len(k) && string(k) == key {
				return keyBegin - 1, nil
			}

		case SquareOpenToken:
			if end := blockEnd(data[i:], data[i], SquareCloseToken); end != -1 {
				i = i + end
			}

		case CurlyOpenToken:
			if end := blockEnd(data[i:], data[i], CurlyCloseToken); end != -1 {
				i = i + end
			}
		}
	}

	return -1, KeyPathNotFoundError
}

func internalGet(data []byte, keys ...string) (value []byte, dataType ValueType, offset, endOffset int, err error) {
	if len(keys) > 0 {
		offset, err = findValueIndex(data, keys...)
		if err != nil {
			return nil, NotExist, -1, -1, err
		}
	}

	// Go to closest value
	nO, err := nextToken(data[offset:])
	if err != nil {
		return nil, NotExist, offset, -1, MalformedJson
	}

	offset += nO
	if value, dataType, endOffset, err = getType(data, offset); err != nil {
		return value, dataType, offset, endOffset, err
	}

	// Strip quotes from string values
	if dataType == String {
		value = value[1 : len(value)-1]
	}

	return value[:len(value):len(value)], dataType, offset, endOffset, nil
}

// Get is a function that retrieves a value from the given data based on the provided keys.
// It returns the value, data type, offset, and any error encountered during the retrieval process.
func Get(data []byte, keys ...string) (value []byte, dataType ValueType, offset int, err error) {
	a, b, _, d, e := internalGet(data, keys...)
	return a, b, d, e
}

func MapKeyTypes(data []byte) (map[string]ValueType, error) {
	keyTypeMap := make(map[string]ValueType)

	keys, err := extractKeys(data)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		_, typ, _, err := Get(data, key)
		if err != nil {
			return nil, err
		}

		keyTypeMap[key] = typ
	}

	return keyTypeMap, nil
}
