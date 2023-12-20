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
