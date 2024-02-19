package json

import (
	"errors"
)

type buffer struct {
	data []byte
	length uint
	index  uint

	last States
	state States
	class Classes
}

type (
	rpn []string
	tokens []string
)

// newBuffer creates a new buffer with the given data
func NewBuffer(data []byte) *buffer {
	return &buffer{
		data: data,
		length: uint(len(data)),
		last: GO,
		state: GO,
	}
}

func (b *buffer) current() (byte, error) {
	if b.index >= b.length {
		return 0, newJSONError(EofErrorCode, "EOF")
	}

	return b.data[b.index], nil
}

func (b *buffer) step() error {
	if b.index+1 >= b.length {
		return newJSONError(EofErrorCode, "EOF")
	}

	b.index++

	return nil
}