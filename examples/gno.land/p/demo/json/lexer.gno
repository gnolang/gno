package json

import (
	"bytes"
	"errors"
	"strconv"
)

func tokenEnd(data []byte) int {
	for i, tok := range data {
		switch tok {
		case WhiteSpaceToken, NewLineToken, CarriageReturnToken, TabToken, ValueSeparatorToken, ObjectEndToken, ArrayEndToken:
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
	if ln > 0 && (data[i] == ObjectStartToken || data[i] == ArrayStartToken) {
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

		case ArrayStartToken:
			if end := blockEnd(data[i:], data[i], ArrayEndToken); end != -1 {
				i = i + end
			}

		case ObjectStartToken:
			if end := blockEnd(data[i:], data[i], ObjectEndToken); end != -1 {
				i = i + end
			}
		}
	}

	return -1, KeyPathNotFoundError
}

func nextToken(data []byte) (int, error) {
	for i, c := range data {
		switch c {
		case WhiteSpaceToken, NewLineToken, CarriageReturnToken, TabToken:
			continue
		default:
			return i, nil
		}
	}

	return -1, TokenNotFound
}

func stringEnd(data []byte) (lastCharIdx int, escaped bool) {
	escaped = false
	for lastCharIdx, c := range data {
		switch c {
		case DoublyQuoteToken:
			if !escaped {
				return lastCharIdx + 1, false
			}

			prevCharIdx := lastCharIdx - 1
			for prevCharIdx >= 0 && data[prevCharIdx] == BackSlashToken {
				prevCharIdx--
			}

			if (lastCharIdx-prevCharIdx)%2 == 0 {
				return lastCharIdx + 1, true
			}

			return lastCharIdx + 1, false
		case BackSlashToken:
			escaped = true
		}
	}

	return -1, escaped
}

// Find end of the data structure, array or object.
// For array openSym and closeSym will be '[' and ']', for object '{' and '}'
func blockEnd(data []byte, openSym byte, closeSym byte) int {
	level := 0

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case DoublyQuoteToken:
			se, _ := stringEnd(data[i+1:])
			if se == -1 {
				return -1
			}
			i += se
		case openSym:
			level += 1
		case closeSym:
			level -= 1
			if level == 0 {
				return i + 1
			}
		}
	}

	return -1
}

// findValueIndex is a function that takes a byte slice of JSON data and a series of keys.
// It traverses the JSON data to find the position where the value of the final key in the series begins.
// The position is returned as an integer representing the index in the byte slice.
// If the key series is empty, it returns 0.
// If the keys are not found or the JSON data is malformed, it returns -1.
func findValueIndex(data []byte, keys ...string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	var (
		i            int
		keyLevel     int
		level        int
		lastMatched  bool = true
		stackbuf	 [UnescapeStackBufSize]byte
	)

	for i < len(data) {
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
			if (lastMatched && keyLevel == level-1 && keyLevel != len(keys)) {
				keyLevel += 1
			}
		
			if keyLevel == len(keys) {
				return i + 1, nil
			}
		case ObjectStartToken:
			// in case parent key is matched then only we will increase the level otherwise can directly
			// can move to the end of this block
			if !lastMatched {
				end := blockEnd(data[i:], ObjectStartToken, ObjectEndToken)
				if end == -1 {
					return -1, MalformedJson
				}
				i += end - 1
			} else {
				level += 1
			}
		case ObjectEndToken:
			level, keyLevel = decreaseLevel(level, keyLevel)
		case ArrayStartToken:
			var (
				curIdx      int
				valueFound  []byte
				valueOffset int
			)

			isArrayKey, err := isArrayKey(level, keyLevel, keys)
			if err != nil {
				return -1, err
			}

			if !isArrayKey {
				arraySkip := blockEnd(data[i:], ArrayStartToken, ArrayEndToken)
				if arraySkip < 0 {
					return -1, MalformedJson
				}

				i += arraySkip - 1
			}

			arrIdx, err := strconv.Atoi(keys[level][1:len(keys[level])-1])
			if err != nil {
				return -1, InvalidArrayIndex
			}

			currIdx := i
			ArrayEach(data[i:], func(value []byte, dataType ValueType, offset int, err error) {
				if curIdx == arrIdx {
					valueFound = value
					valueOffset = offset
					if dataType == String {
						valueOffset = valueOffset - 2
						valueFound = data[currIdx+valueOffset : currIdx+valueOffset+len(value)+2]
					}
				}
				curIdx += 1
			})

			if valueFound == nil {
				return -1, ArrayIndexNotFound
			}

			subIdx, err := findValueIndex(valueFound, keys[level+1:]...)
			if err != nil {
				return -1, err
			}

			return i + valueOffset + subIdx, nil
		case ColonToken: // If encountered, JSON data is malformed
			return -1, MalformedJson
		}

		i++
	}

	return -1, KeyPathNotFoundError
}

func ExtractKeys(data []byte) (keys []string, err error) {
	var (
		stack []byte
		inString bool
		potentialKey string
		afterString bool
	)

    for i := 0; i < len(data); i++ {
        switch data[i] {
        case ObjectStartToken, ArrayStartToken:
            stack = append(stack, data[i])
        case ObjectEndToken, ArrayEndToken:
            if len(stack) > 0 {
                stack = stack[:len(stack)-1]
            }
        case DoublyQuoteToken:
            if inString {
                // End of string
                inString = false
                afterString = true
            } else {
                // Start of string
                inString = true
                potentialKey = ""
                if len(stack) > 0 && stack[len(stack)-1] == ObjectStartToken {
                    afterString = false
                }
            }
        case ColonToken:
            if afterString && len(stack) > 0 && stack[len(stack)-1] == ObjectStartToken {
                keys = append(keys, potentialKey)
                afterString = false
            }
        default:
            if inString {
                potentialKey += string(data[i])
            }
            if !inString && data[i] != WhiteSpaceToken && data[i] != NewLineToken && data[i] != CarriageReturnToken && data[i] != TabToken {
                afterString = false
            }
        }
    }

    return keys, nil
}

func decreaseLevel(level, keyLevel int) (int, int) {
	if level -= 1; level == keyLevel {
		keyLevel -= 1
	}

	return level, keyLevel
}

func keyMatched(key []byte, keyEscaped bool, stackbuf [UnescapeStackBufSize]byte, keys []string, level int) (keyUnesc []byte, err error) {
	if !keyEscaped {
		keyUnesc = key
	} 
	
	if ku, err := Unescape(key, stackbuf[:]); err != nil {
		return nil, err
	} else {
		keyUnesc = ku
	}

	if level > len(keys) {
		return nil, KeyLevelNotMatched
	}

	return keyUnesc, err
}

func isArrayKey(level int, keyLevel int, keys []string) (bool, error) {
    if keyLevel != level {
        return false, nil
    }

    key := keys[level]
    keyLen := len(key)

    if keyLen < 3 || key[0] != ArrayStartToken || key[keyLen-1] != ArrayEndToken {
        return false, MalformedArray
    }

    return true, nil
}

func Get(data []byte, keys ...string) (value []byte, dataType ValueType, offset int, err error) {
	a, b, _, d, e := internalGet(data, keys...)
	return a, b, d, e
}
