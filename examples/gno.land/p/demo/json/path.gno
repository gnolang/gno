package json

// ref: https://support.smartbear.com/alertsite/docs/monitors/api/endpoint/jsonpath.html

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

// ParsePath takes a JSONPath string and returns a slice of strings representing the path segments.
func ParsePath(path string) ([]string, error) {
	buf := newBuffer([]byte(path))
	result := make([]string, 0)

	for {
		b, err := buf.current()
		if err != nil {
			break
		}

		switch {
		case b == dollarSign || b == atSign:
			result = append(result, string(b))
			buf.step()

		case b == dot:
			buf.step()

			if next, _ := buf.current(); next == dot {
				buf.step()
				result = append(result, "..")

				extractNextSegment(buf, &result)
			} else {
				extractNextSegment(buf, &result)
			}

		case b == bracketOpen:
			start := buf.index
			buf.step()

			for {
				if buf.index >= buf.length || buf.data[buf.index] == bracketClose {
					break
				}

				buf.step()
			}

			if buf.index >= buf.length {
				return nil, errors.New("unexpected end of path")
			}

			segment := string(buf.sliceFromIndices(start+1, buf.index))
			result = append(result, segment)

			buf.step()

		default:
			buf.step()
		}
	}

	return result, nil
}

// extractNextSegment extracts the segment from the current index
// to the next significant character and adds it to the resulting slice.
func extractNextSegment(buf *buffer, result *[]string) {
	start := buf.index
	buf.skipToNextSignificantToken()

	if buf.index <= start {
		return
	}

	segment := string(buf.sliceFromIndices(start, buf.index))
	if segment != "" {
		*result = append(*result, segment)
	}
}

// numIndex holds a map of valid numeric characters
var numIndex = map[byte]bool{
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
	'.': true,
	'e': true,
	'E': true,
}

// pathToken checks if the current token is a valid JSON path token.
func (b *buffer) pathToken() error {
	var (
		stack             []byte
		inToken, inNumber bool
	)

	first := b.index

	for b.index < b.length {
		c := b.data[b.index]

		switch {
		case c == doubleQuote, c == singleQuote:
			inToken = true
			if err := b.step(); err != nil {
				return errors.New("error stepping through buffer")
			}

			if err := b.skip(c); err != nil {
				return errors.New("unmatched quote in path")
			}

			if b.index >= b.length {
				return errors.New("unmatched quote in path")
			}

		case c == bracketOpen, c == parenOpen:
			inToken = true
			stack = append(stack, c)

		case c == bracketClose, c == parenClose:
			inToken = true
			ls := len(stack)
			if ls == 0 {
				return errors.New("unexpected end of path")
			}

			if (c == bracketClose && stack[ls-1] != bracketOpen) || (c == parenClose && stack[ls-1] != parenOpen) {
				return errors.New("mismatched bracket or parenthesis")
			}

			stack = stack[:ls-1]

		case pathStateContainsValidPathToken(c):
			inToken = true

		case c == plus, c == minus:
			if inNumber || (b.index > 0 && numIndex[b.data[b.index-1]]) {
				inToken = true
			}

			if !inToken && (b.index+1 < b.length && numIndex[b.data[b.index+1]]) {
				inToken = true
				inNumber = true
			}

			if !inToken {
				return errors.New("unexpected operator at start of token")
			}

		default:
			if len(stack) != 0 || inToken {
				inToken = true
			} else {
				goto end
			}
		}

		b.index++
	}

end:
	if len(stack) != 0 {
		return errors.New("unclosed bracket or parenthesis at end of path")
	}

	if first == b.index {
		return errors.New("no token found")
	}

	if inNumber && !numIndex[b.data[b.index-1]] {
		inNumber = false
	}

	return nil
}

func pathStateContainsValidPathToken(c byte) bool {
	if _, ok := significantTokens[c]; ok {
		return true
	}

	if _, ok := filterTokens[c]; ok {
		return true
	}

	if _, ok := numIndex[c]; ok {
		return true
	}

	if unicode.IsLetter(rune(c)) {
		return true
	}

	return false
}

func (b *buffer) pathTokenizer() ([]string, error) {
	var result []string

	for {
		b.reset()
		c, err := b.first()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var token string
		switch {
		case unicode.IsDigit(rune(c)), c == dot:
			start := b.index
			err = b.numeric(true)
			if err == io.EOF {
				token = string(b.sliceFromIndices(start, b.index))
			} else if err != nil {
				if c == dot {
					token = "."
					b.index = start
				} else {
					return nil, err
				}
			} else {
				token = string(b.sliceFromIndices(start, b.index))
				b.index--
			}

		case c == singleQuote, c == doubleQuote:
			start := b.index
			err = b.string(c, true)
			if err != nil {
				return nil, errors.New("error stepping through buffer")
			}
			token = string(b.sliceFromIndices(start, b.index+1))

		case c == dollarSign, c == atSign:
			start := b.index
			err = b.findNextToken()
			if err != nil && err != io.EOF {
				return nil, err
			}
			token = string(b.sliceFromIndices(start, b.index))
			if err != io.EOF {
				b.index--
			}

		case c == parenOpen, c == parenClose:
			token = string(c)

		default:
			start := b.index
			for ; b.index < b.length; b.index++ {
				c = b.data[b.index]
				if c == parenOpen {
					break
				}

				rc := rune(c)
				if unicode.IsLetter(rc) && !unicode.IsDigit(rc) && c != '_' {
					break
				}
			}
			err = b.step()
			if err != nil && err != io.EOF {
				return nil, err
			}

			slice := string(b.sliceFromIndices(start, b.index))
			token = strings.ToLower(slice)

			if err != io.EOF {
				b.index--
			}
		}

		result = append(result, token)
		err = b.step()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (b *buffer) findNextToken() error {
	var (
		c     byte
		stack []byte
		find  bool
		start int
	)

	for b.index < b.length {
		c = b.data[b.index]

		switch {
		case c == singleQuote, c == doubleQuote:
			find = true
			if err := b.step(); err != nil {
				return errors.New("error stepping through buffer")
			}
			if err := b.skip(c); err == io.EOF {
				return errors.New("unmatched quote in path")
			}
		case c == bracketOpen, c == parenOpen:
			find = true
			stack = append(stack, c)
		case c == bracketClose, c == parenClose:
			find = true
			ls := len(stack)
			if ls == 0 {
				if start == b.index {
					return errors.New("unexpected end of path")
				}
				break
			}
			expected := bracketOpen
			if c == parenClose {
				expected = parenOpen
			}
			if stack[ls-1] != expected {
				return errors.New("mismatched bracket or parenthesis")
			}
			stack = stack[:ls-1]
		case c == dot, c == atSign, c == dollarSign, c == question, c == asterisk, isAlphaNumeric(c):
			find = true
			b.index++
			continue
		case len(stack) != 0:
			find = true
			b.index++
			continue
		case c == minus, c == plus:
			if !find {
				find = true
				start = b.index
				if err := b.numeric(true); err == nil || err == io.EOF {
					b.index--
					b.index++
					continue
				}
				b.index = start
			}
			fallthrough
		default:
			break
		}

		if len(stack) == 0 {
			break
		}
		b.index++
	}

	if len(stack) != 0 {
		return errors.New("unclosed bracket or parenthesis at end of path")
	}
	if start == b.index {
		return b.step()
	}
	if b.index >= b.length {
		return io.EOF
	}
	return nil
}
