package json

import (
	"bytes"
	"errors"
	"strings"

	"gno.land/p/demo/ufmt"
)

type buffer struct {
	data   []byte
	length uint
	index  uint

	last  States
	state States
	class Classes
}

type tokens []string

// newBuffer creates a new buffer with the given data
func NewBuffer(data []byte) *buffer {
	return &buffer{
		data:   data,
		length: uint(len(data)),
		last:   GO,
		state:  GO,
	}
}

// current returns the byte of the current index.
func (b *buffer) current() (byte, error) {
	if b.index >= b.length {
		return 0, errors.New("EOF")
	}

	return b.data[b.index], nil
}

// next moves to the next byte and returns it.
func (b *buffer) next() (byte, error) {
	b.index++
	return b.current()
}

// step just moves to the next position.
func (b *buffer) step() error {
	_, err := b.next()
	return err
}

// move moves the index by the given position.
func (b *buffer) move(pos uint) error {
	newIndex := b.index + pos

	if newIndex > b.length {
		return errors.New("EOF")
	}

	b.index = newIndex

	return nil
}

// slice returns the slice from the current index to the given position.
func (b *buffer) slice(pos uint) ([]byte, error) {
	end := b.index + pos

	if end > b.length {
		return nil, errors.New("EOF")
	}

	return b.data[b.index:end], nil
}

// sliceFromIndices returns a slice of the buffer's data starting from 'start' up to (but not including) 'stop'.
func (b *buffer) sliceFromIndices(start, stop uint) []byte {
	if start > b.length {
		start = b.length
	}

	if stop > b.length {
		stop = b.length
	}

	return b.data[start:stop]
}

// skip moves the index to skip the given byte.
func (b *buffer) skip(bs byte) error {
	for b.index < b.length {
		if b.data[b.index] == bs && !b.backslash() {
			return nil
		}

		b.index++
	}

	return errors.New("EOF")
}

// skipAny moves the index until it encounters one of the given set of bytes.
func (b *buffer) skipAny(endTokens map[byte]bool) error {
	for b.index < b.length {
		if _, exists := endTokens[b.data[b.index]]; exists {
			return nil
		}

		b.index++
	}

	// build error message
	var tokens []string
	for token := range endTokens {
		tokens = append(tokens, string(token))
	}

	return errors.New(
		ufmt.Sprintf(
			"EOF reached before encountering one of the expected tokens: %s",
			strings.Join(tokens, ", "),
		))
}

// skipAndReturnIndex moves the buffer index forward by one and returns the new index.
func (b *buffer) skipAndReturnIndex() (uint, error) {
	err := b.step()
	if err != nil {
		return 0, err
	}

	return b.index, nil
}

// skipUntil moves the buffer index forward until it encounters a byte contained in the endTokens set.
func (b *buffer) skipUntil(endTokens map[byte]bool) (uint, error) {
	for b.index < b.length {
		currentByte, err := b.current()
		if err != nil {
			return b.index, err
		}

		// Check if the current byte is in the set of end tokens.
		if _, exists := endTokens[currentByte]; exists {
			return b.index, nil
		}

		b.index++
	}

	return b.index, errors.New("EOF")
}

// significantTokens is a map where the keys are the significant characters in a JSON path.
// The values in the map are all true, which allows us to use the map as a set for quick lookups.
var significantTokens = map[byte]bool{
	DotToken:         true, // access properties of an object
	DollarToken:      true, // root object
	AtToken:          true, // current object in a filter expression
	SquareOpenToken:  true, // start of an array index or filter expression
	SquareCloseToken: true, // end of an array index or filter expression
}

// skipToNextSignificantToken advances the buffer index to the next significant character.
// Significant characters are defined based on the JSON path syntax.
func (b *buffer) skipToNextSignificantToken() {
	for b.index < b.length {
		current := b.data[b.index]

		if _, ok := significantTokens[current]; ok {
			break
		}

		b.index++
	}
}

// backslash checks to see if the number of backslashes before the current index is odd.
//
// This is used to check if the current character is escaped. However, unlike the "unescape" function,
// "backslash" only serves to check the number of backslashes.
func (b *buffer) backslash() bool {
	if b.index == 0 {
		return false
	}

	count := 0
	for i := b.index - 1; ; i-- {
		if i >= b.length || b.data[i] != BackSlashToken {
			break
		}

		count++

		if i == 0 {
			break
		}
	}

	return count%2 != 0
}

func (b *buffer) reset() {
	b.last = GO
}

func (b *buffer) pathToken() error {
	var stack []byte

	inToken := false
	inNumber := false
	first := b.index

	for b.index < b.length {
		c := b.data[b.index]

		switch {
		case c == DoublyQuoteToken || c == QuoteToken:
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

		case c == SquareOpenToken || c == RoundOpenToken:
			inToken = true
			stack = append(stack, c)

		case c == SquareCloseToken || c == RoundCloseToken:
			inToken = true
			if len(stack) == 0 || (c == SquareCloseToken && stack[len(stack)-1] != SquareOpenToken) || (c == RoundCloseToken && stack[len(stack)-1] != RoundOpenToken) {
				return errors.New("mismatched bracket or parenthesis")
			}

			stack = stack[:len(stack)-1]

		case bytes.ContainsAny([]byte{DotToken, CommaToken, DollarToken, AtToken, AesteriskToken, AndToken, OrToken}, string(c)) || ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || ('0' <= c && c <= '9'):
			inToken = true

		case c == PlusToken || c == MinusToken:
			if inNumber || (b.index > 0 && bytes.ContainsAny([]byte("eE"), string(b.data[b.index-1]))) {
				inToken = true
			} else if !inToken && (b.index+1 < b.length && bytes.IndexByte([]byte("0123456789"), b.data[b.index+1]) != -1) {
				inToken = true
				inNumber = true
			} else if !inToken {
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

	if inNumber && !bytes.ContainsAny([]byte("0123456789.eE"), string(b.data[b.index-1])) {
		inNumber = false
	}

	return nil
}

func (b *buffer) numeric(token bool) error {
	if token {
		b.last = GO
	}

	for _; b.index < b.length; b.index++ {
		b.class = b.getClasses(DoublyQuoteToken)
		if b.class == __ {
			return errors.New("invalid token found while parsing path")
		}

		b.state = StateTransitionTable[b.last][b.class]
		if b.state == __ {
			if token {
				break
			}

			return errors.New("invalid token found while parsing path")
		}

		if b.state < __ {
			return nil
		}

		if b.state < MI || b.state > E3 {
			return nil
		}

		b.last = b.state
	}

	if b.last != ZE && b.last != IN && b.last != FR && b.last != E3 {
		return errors.New("invalid token found while parsing path")
	}

	return nil
}

func (b *buffer) getClasses(c byte) Classes {
	if b.data[b.index] >= 128 {
		return C_ETC
	}

	if c == QuoteToken {
		return QuoteAsciiClasses[b.data[b.index]]
	}

	return AsciiClasses[b.data[b.index]]
}

func (b *buffer) getState() States {
	b.last = b.state

	b.class = b.getClasses(DoublyQuoteToken)
	if b.class == __ {
		return __
	}

	b.state = StateTransitionTable[b.last][b.class]

	return b.state
}
