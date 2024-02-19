package json

import (
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
		return 0, newJSONError(EofError, "EOF")
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
		return newJSONError(EofError, "EOF")
	}

	b.index = newIndex

	return nil
}

// slice returns the slice from the current index to the given position.
func (b *buffer) slice(pos uint) ([]byte, error) {
	end := b.index + pos

	if end > b.length {
		return nil, newJSONError(EofError, "EOF")
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

	return newJSONError(EofError, "EOF")
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

	return newJSONError(
		EofError,
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
            return b.index, newJSONError(EofError, "Reached end of buffer without finding token")
        }

        // Check if the current byte is in the set of end tokens.
        if _, exists := endTokens[currentByte]; exists {
            return b.index, nil
        }

        b.index++
    }

    return b.index, newJSONError(EofError, "End of file reached before token found")
}


// significantTokens is a map where the keys are the significant characters in a JSON path.
// The values in the map are all true, which allows us to use the map as a set for quick lookups.
var significantTokens = map[byte]bool{
	DotToken:         true,    // access properties of an object
	DollarToken:      true,    // root object
	AtToken:          true,    // current object in a filter expression
	SquareOpenToken:  true,    // start of an array index or filter expression
	SquareCloseToken: true,    // end of an array index or filter expression
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
        if i >= b.length || b.data[i] != '\\' {
            break
        }

        count++

        if i == 0 {
            break
        }
    }

	return count%2 != 0
}

// reset initializes the state of the buffer.
func (b *buffer) reset() {
	b.last = GO
}
