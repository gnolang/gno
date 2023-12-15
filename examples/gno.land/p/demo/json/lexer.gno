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

func findKeyStart(data []byte, key string) (int, error) {
	i, _ := nextToken(data)
	if i == -1 {
		return i, KeyPathNotFoundError
	}
	ln := len(data)
	if ln > 0 && (data[i] == '{' || data[i] == '[') {
		i += 1
	}
	var stackbuf [UnescapeStackBufSize]byte // stack-allocated array for allocation-free unescaping of small strings

	if ku, err := Unescape([]byte(key), stackbuf[:]); err == nil {
		key = string(ku)
	}

	for i < ln {
		switch data[i] {
		case '"':
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

			if data[i] == ':' && len(key) == len(k) && string(k) == key {
				return keyBegin - 1, nil
			}

		case '[':
			end := blockEnd(data[i:], data[i], ']')
			if end != -1 {
				i = i + end
			}
		case '{':
			end := blockEnd(data[i:], data[i], '}')
			if end != -1 {
				i = i + end
			}
		}
		i++
	}

	return -1, KeyPathNotFoundError
}

func tokenStart(data []byte) int {
	for i := len(data) - 1; i >= 0; i-- {
		switch data[i] {
		case '\n', '\r', '\t', ',', '{', '[':
			return i
		}
	}

	return 0
}

func nextToken(data []byte) (int, error) {
	for i, c := range data {
		switch c {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i, nil
		}
	}

	return -1, TokenNotFound
}

func lastToken(data []byte) int {
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

func stringEnd(data []byte) (lastCharIdx int, escaped bool) {
	escaped = false
	for lastCharIdx, c := range data {
		switch c {
		case '"':
			if !escaped {
				return lastCharIdx + 1, false
			}

			prevCharIdx := lastCharIdx - 1
			for prevCharIdx >= 0 && data[prevCharIdx] == '\\' {
				prevCharIdx--
			}

			if (lastCharIdx-prevCharIdx)%2 == 0 {
				return lastCharIdx + 1, true
			}

			return lastCharIdx + 1, false
		case '\\':
			escaped = true
		}
	}

	return -1, escaped
}

// Find end of the data structure, array or object.
// For array openSym and closeSym will be '[' and ']', for object '{' and '}'
func blockEnd(data []byte, openSym byte, closeSym byte) int {
	level := 0
	i := 0
	ln := len(data)

	for i < ln {
		switch data[i] {
		case '"': // If inside string, skip it
			se, _ := stringEnd(data[i+1:])
			if se == -1 {
				return -1
			}
			i += se
		case openSym: // If open symbol, increase level
			level++
		case closeSym: // If close symbol, increase level
			level--

			// If we have returned to the original level, we're done
			if level == 0 {
				return i + 1
			}
		}
		i++
	}

	return -1
}

// TODO: change function name to something more appropriate
//
// searchKeys is a function that takes a byte slice of JSON data and a series of keys.
// It traverses the JSON data to find the position where the value of the final key in the series begins.
// The position is returned as an integer representing the index in the byte slice.
// If the key series is empty, it returns 0.
// If the keys are not found or the JSON data is malformed, it returns -1.
func searchKeys(data []byte, keys ...string) (int, error) {
	keyLevel := 0
	level := 0
	i := 0
	ln := len(data)
	lk := len(keys)
	lastMatched := true

	if lk == 0 {
		return 0, nil
	}

	var stackbuf [UnescapeStackBufSize]byte // stack-allocated array for allocation-free unescaping of small strings

	for i < ln {
		switch data[i] {
		case '"':
			i++
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

			// if string is a key

			if data[i] == ':' {
				if level < 1 {
					return -1, MalformedJson
				}

				key := data[keyBegin:keyEnd]

				// for unescape: if there are no escape sequences, this is cheap; if there are, it is a
				// bit more expensive, but causes no allocations unless len(key) > unescapeStackBufSize
				var keyUnesc []byte
				if !keyEscaped {
					keyUnesc = key
				} else if ku, err := Unescape(key, stackbuf[:]); err != nil {
					return -1, err
				} else {
					keyUnesc = ku
				}

				if level > len(keys) {
					return -1, KeyLevelNotMatched
				}

				if bytes.Equal(keyUnesc, []byte(keys[level-1])) {
					lastMatched = true

					// if key level match
					if keyLevel == level-1 {
						keyLevel++
						// If we found all keys in path
						if keyLevel == lk {
							return i + 1, nil
						}
					}
				} else {
					lastMatched = false
				}
			} else {
				i--
			}
		case '{':
			// in case parent key is matched then only we will increase the level otherwise can directly
			// can move to the end of this block
			if !lastMatched {
				end := blockEnd(data[i:], '{', '}')
				if end == -1 {
					return -1, MalformedJson
				}
				i += end - 1
			} else {
				level += 1
			}
		case '}':
			level -= 1
			if level == keyLevel {
				keyLevel -= 1
			}
		case '[':
			// If we want to get array element by index
			if keyLevel == level && keys[level][0] == '[' {
				keyLen := len(keys[level])
				if keyLen < 3 || keys[level][0] != '[' || keys[level][keyLen-1] != ']' {
					return -1, MalformedArray
				}

				arrIdx, err := strconv.Atoi(keys[level][1 : keyLen-1])
				if err != nil {
					return -1, InvalidArrayIndex
				}
				var curIdx int
				var valueFound []byte
				var valueOffset int
				curI := i
				ArrayEach(data[i:], func(value []byte, dataType ValueType, offset int, err error) {
					if curIdx == arrIdx {
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
				}

				subIndex, err := searchKeys(valueFound, keys[level+1:]...)
				if err != nil {
					return -1, err
				}

				return i + valueOffset + subIndex, nil
			} else {
				arraySkip := blockEnd(data[i:], '[', ']')
				if arraySkip < 0 {
					return -1, MalformedJson
				}

				i += arraySkip - 1
			}
		case ':': // If encountered, JSON data is malformed
			return -1, MalformedJson
		}

		i++
	}

	return -1, KeyPathNotFoundError
}

func Get(data []byte, keys ...string) (value []byte, dataType ValueType, offset int, err error) {
	a, b, _, d, e := internalGet(data, keys...)
	return a, b, d, e
}
