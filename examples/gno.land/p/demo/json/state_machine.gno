package json

import (
	"bytes"
	"errors"
	"strconv"

	"gno.land/p/demo/ufmt"
)

// extractValueTypeFromToken recoginizes the type of JSON value by peeking at the first byte of the value.
func extractValueTypeFromToken(b byte, value []byte) (dataType ValueType, offset int, err error) {
	if len(b) == 0 {
		return Unknown, 0, errors.New("json: empty value, cannot determine type.")
	}

	switch b {
	case 't', 'f':
		if bytes.Equal(value, trueLiteral) || bytes.Equal(value, falseLiteral) {
			dataType = Boolean
		}
		return Unknown, offset, UnknownValueType
	case 'u', 'n':
		if bytes.Equal(value, nullLiteral) {
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
		stackbuf    [unescapeStackBufSize]byte
	)

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case DoublyQuoteToken:
			i += 1
			keyBegin := i

			strEnd, keyEscaped, err := stringEnd(data[i:])
			if err != nil {
				return -1, err
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
				end, err := blockEnd(data[i:], CurlyOpenToken, CurlyCloseToken)
				if end == -1 {
					return -1, err
				}

				i += end - 1
			} else {
				level += 1
			}
		case CurlyCloseToken:
			level, keyLevel = decreaseLevel(level, keyLevel)
		case SquareOpenToken:
			// If we want to get array element by index
			if keyLevel == level && keys[level][0] == SquareOpenToken {
				keyLen := len(keys[level])
				if keyLen < 3 || keys[level][0] != SquareOpenToken || keys[level][keyLen-1] != SquareCloseToken {
					return -1, InvalidArrayIndex
				}

				aIdx, err := strconv.Atoi(keys[level][1 : keyLen-1])
				if err != nil {
					return -1, InvalidArrayIndex
				}

				var (
					curIdx      int
					valueFound  []byte
					valueOffset int
				)

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
				}

				subIndex, err := findValueIndex(valueFound, keys[level+1:]...)
				if err != nil {
					return -1, KeyPathNotFoundError
				}

				return i + valueOffset + subIndex, nil
			} else {
				// Do not search for keys inside arrays
				if arraySkip, err := blockEnd(data[i:], SquareOpenToken, SquareCloseToken); arraySkip == -1 {
					return -1, err
				} else {
					i += arraySkip - 1
				}
			}
		case ColonToken:
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

// findKeyStart finds the start of a specific key in a given byte array.
func findKeyStart(data []byte, key string) (int, error) {
	i, _ := nextToken(data)
	if i == -1 {
		return i, KeyPathNotFoundError
	}

	ln := len(data)
	if ln > 0 && (data[i] == CurlyOpenToken || data[i] == SquareOpenToken) {
		i += 1
	}

	var stackbuf [unescapeStackBufSize]byte
	if ku, err := unescape([]byte(key), stackbuf[:]); err == nil {
		key = string(ku)
	}

	for _; i < ln; i++ {
		switch data[i] {
		case DoublyQuoteToken:
			i++
			keyBegin := i

			strEnd, keyEscaped, _ := stringEnd(data[i:])
			if strEnd == -1 {
				break
			}

			i += strEnd
			keyEnd := i - 1

			valueOffset, err := nextToken(data[i:])
			if err != nil {
				break
			}

			i += valueOffset

			// if string is a key, and key level match
			k := data[keyBegin:keyEnd]

			// for unescape: if there are no escape sequences, this is cheap; if there are, it is a
			// bit more expensive, but causes no allocations unless len(key) > unescapeStackBufSize
			if keyEscaped {
				if ku, err := unescape(k, stackbuf[:]); err != nil {
					break
				} else {
					k = ku
				}
			}

			if data[i] == ColonToken && len(key) == len(k) && string(k) == key {
				return keyBegin - 1, nil
			}

		case SquareOpenToken, CurlyOpenToken:
			if end, _ := blockEnd(data[i:], data[i], SquareCloseToken); end != -1 {
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
	if value, dataType, endOffset, err = getTypeFromByteSlice(data, offset); err != nil {
		return value, dataType, offset, endOffset, err
	}

	// strip quotes from string values
	if dataType == String {
		value = value[1:len(value)-1]
	}

	// remove unnecessary whitespace characters to make more readable
	if dataType == Object {
		value = adjustIndentation(value)
	}

	return value[:len(value):len(value)], dataType, offset, endOffset, nil
}

// adjustIndentation is a kind of hack to remove the extra tabs and spaces from the Object values.
// TODO: storing depth information while parsing step, may allows us to avoid regex.
func adjustIndentation(value []byte) []byte {
	value = tabRegex.ReplaceAll(value, []byte("\t"))
	value = lastLineSpaceRegex.ReplaceAll(value, []byte("}"))

	return value
}

// Get is a function that retrieves a value from the given data based on the provided keys.
// It returns the value, data type, offset (position), and any error encountered during the retrieval process.
//
// For example:
//
// ``` gno
// data := []byte(`{"name":{"first":"Janet","last":"Prichard"},"age":47}`)
// value, dataType, offset, err := json.Get(data, "name", "last")
// ```
//
// It returns the value `Prichard` with data type `String`, the offset `37`, and `nil` for the error.
func Get(data []byte, keys ...string) (value []byte, dataType ValueType, offset int, err error) {
	a, b, _, d, e := internalGet(data, keys...)

	return a, b, d, e
}

func ArrayEach(
	data []byte,
	cb func(value []byte, dataType ValueType, offset int, err error),
	keys ...string,
) (offset int, err error) {
	if len(data) == 0 {
		return -1, errors.New("json.ArrayEach: empty data")
	}

	if next, err := nextToken(data); err != nil {
		return -1, errors.New("json.ArrayEach: invalid token")
	} else {
		offset += next
	}

	if len(keys) > 0 {
		if offset, err = findValueIndex(data, keys...); err != nil {
			return -1, errors.New(ufmt.Sprintf("json.ArrayEach: key path not found. keys: %v", keys))
		}

		// go to closest value
		if n0, err := nextToken(data[offset:]); err != nil {
			return -1, errors.New(
				ufmt.Sprintf(
					"json.ArrayEach: invalid token found. offset: %d. got: %s",
					offset,
					data[offset:offset+1],
				))
		} else {
			offset += n0
		}

		if data[offset] != SquareOpenToken {
			return -1, errors.New("json.ArrayEach: invalid array. must start with `[`")
		} else {
			offset += 1
		}
	}

	if n0, err := nextToken(data[offset:]); err != nil {
		return -1, errors.New("json.ArrayEach: invalid token. value not found")
	} else {
		offset += n0
	}

	if data[offset] == SquareCloseToken {
		return -1, nil
	}

	for true {
		v, t, o, e := Get(data[offset:])

		if e != nil {
			return -1, e
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
		if skipToToken, err := nextToken(data[offset:]); err != nil {
			return offset, errors.New(
				ufmt.Sprintf(
					"json.ArrayEach: invalid token at offset %d. got: %s",
					offset,
					data[offset:offset+1],
				))
		} else {
			offset += skipToToken
		}

		if data[offset] == SquareCloseToken {
			break
		}

		if data[offset] != CommaToken {
			return offset, errors.New(
				ufmt.Sprintf(
					"json.ArrayEach: cannot find comma at offset %d. got: %s",
					offset,
					data[offset:offset+1],
				))
		} else {
			offset += 1
		}
	}

	return offset, nil
}

func ObjectEach(
	data []byte,
	cb func(key, value []byte, dataType ValueType, offset int) error,
	keys ...string,
) (err error) {
	offset := 0

	// descent to desired key
	if len(keys) > 0 {
		if off, err := findValueIndex(data, keys...); err != nil {
			errors.New("json.ObjectEach: key path not found")
		} else {
			offset += off
		}
	}

	// validate and skip past opening token
	if off, err := nextToken(data[offset:]); err != nil {
		return MalformedObject
	} else if offset += off; data[offset] != CurlyOpenToken {
		return errors.New(ufmt.Sprintf("json.ObjectEach: invalid token at offset %d", offset))
	} else {
		offset += 1
	}

	// skip to the 1st token insize the object
	if off, err := nextToken(data[offset:]); err != nil {
		return errors.New(ufmt.Sprintf("json.ObjectEach: invalid token at offset %d", offset))
	} else if offset += off; data[offset] == CurlyCloseToken {
		// find the end of the object. should return here
		return nil
	}

	for offset < len(data) {
		// find the next key
		var key []byte

		// check the next token kind
		switch data[offset] {
		case DoublyQuoteToken:
			offset += 1 // consider as a string type and skip the first quote
		case CurlyCloseToken:
			return nil // end of the object
		default:
			return MalformedObject // invalid token found
		}

		// find the end of the key
		var escaped bool
		if off, ok, err := stringEnd(data[offset:]); err != nil {
			return errors.New(ufmt.Sprintf("json.ObjectEach: invalid string at offset %d", offset))
		} else {
			key, escaped = data[offset:offset+off-1], ok
			offset += off
		}

		// unescape the key if necessary
		if escaped {
			var stackbuf [unescapeStackBufSize]byte

			if ku, err := unescape(key, stackbuf[:]); err != nil {
				return errors.New("json.ObjectEach: string is not valid UTF-8")
			} else {
				key = ku
			}
		}

		// skip colon
		if off, err := nextToken(data[offset:]); err != nil {
			return errors.New(ufmt.Sprintf("json.ObjectEach: invalid token at offset %d", offset))
		} else if offset += off; data[offset] != ColonToken {
			return errors.New(ufmt.Sprintf("json.ObjectEach: cannot find colon at expected offset %d", offset))
		} else {
			offset += 1
		}

		// find nearby value, call callback
		if value, typ, off, err := Get(data[offset:]); err != nil {
			return errors.New(ufmt.Sprintf("json.ObjectEach: %s", err))
		} else if err := cb(key, value, typ, offset+off); err != nil {
			return errors.New(ufmt.Sprintf("json.ObjectEach: %s", err)) 
		} else {
			offset += off
		}

		// skip the next comma to the following token
		if off, err := nextToken(data[offset:]); err != nil {
			return errors.New(ufmt.Sprintf("json.ObjectEach: can't find next token at offset %d", offset))
		} else {
			offset += off

			switch data[offset] {
			case CurlyCloseToken:
				return nil // end of the object. should stop here.
			case CommaToken:
				offset += 1 // skip the comma and continue
			default:
				return errors.New(
					ufmt.Sprintf("json.ObjectEach: object is not properly terminated at offset %d", offset),
				)
			}
		}

		// skip the next token if comma is found
		if off, err := nextToken(data[offset:]); err != nil {
			return errors.New(ufmt.Sprintf("json.ObjectEach: can't find next token at offset %d", offset))
		} else {
			offset += off // skip the next token
		}

		// reset the key
		key = nil
	}

	return errors.New(ufmt.Sprintf("json.ObjectEach: object is not properly terminated at offset %d", offset))
}
