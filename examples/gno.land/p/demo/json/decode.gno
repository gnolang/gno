// ref: https://github.com/spyzhov/ajson/blob/master/decode.go

package json

import (
	"errors"
	"io"

	"gno.land/p/demo/ufmt"
)

// This limits the max nesting depth to prevent stack overflow.
// This is permitted by https://tools.ietf.org/html/rfc7159#section-9
const maxNestingDepth = 10000

// Unmarshal parses the JSON-encoded data and returns a Node.
// The data must be a valid JSON-encoded value.
//
// Usage:
// 	node, err := json.Unmarshal([]byte(`{"key": "value"}`))
// 	if err != nil {
// 		ufmt.Println(err)
// 	}
// 	println(node) // {"key": "value"}
func Unmarshal(data []byte) (*Node, error) {
	buf := newBuffer(data)

	var (
		state   States
		key     *string
		current *Node
		nesting int
		useKey  = func() **string {
			tmp := cptrs(key)
			key = nil
			return &tmp
		}
		err error
	)

	if _, err = buf.first(); err != nil {
		return nil, io.EOF
	}

	for {
		state = buf.getState()
		if state == __ {
			return nil, unexpectedTokenError(buf.data, buf.index)
		}

		// region state machine
		if state >= GO {
			switch buf.state {
			case ST: // string
				if current != nil && current.IsObject() && key == nil {
					// key detected
					if key, err = getString(buf); err != nil {
						return nil, err
					}

					buf.state = CO
				} else {
					if nesting, err = checkNestingDepth(nesting); err != nil {
						return nil, err
					}

					current, err = createNode(current, buf, String, useKey())
					if err != nil {
						return nil, err
					}

					if err = buf.string(doubleQuote, false); err != nil {
						return nil, err
					}

					current, nesting = updateNode(current, buf, nesting, true)
					buf.state = OK
				}

			case MI, ZE, IN: // number
				if current, err = createNode(current, buf, Number, useKey()); err != nil {
					return nil, err
				}

				if err = buf.numeric(false); err != nil {
					return nil, err
				}

				current.borders[1] = buf.index
				if current.prev != nil {
					current = current.prev
				}

				buf.index -= 1
				buf.state = OK

			case T1, F1: // boolean
				if current, err = createNode(current, buf, Boolean, useKey()); err != nil {
					return nil, err
				}

				var literal []byte
				if buf.state == T1 {
					literal = trueLiteral
				} else {
					literal = falseLiteral
				}

				if err = buf.word(literal); err != nil {
					return nil, err
				}

				current, nesting = updateNode(current, buf, nesting, false)
				buf.state = OK

			case N1: // null
				if current, err = createNode(current, buf, Null, useKey()); err != nil {
					return nil, err
				}

				if err = buf.word(nullLiteral); err != nil {
					return nil, err
				}

				current, nesting = updateNode(current, buf, nesting, false)
				buf.state = OK
			}
		} else {
			// region action
			switch state {
			case ec, cc: // <empty> }
				if key != nil {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				if !isValidContainerType(current, Object) {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				current, nesting = updateNode(current, buf, nesting, true)
				buf.state = OK

			case bc: // ]
				if !isValidContainerType(current, Array) {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				current, nesting = updateNode(current, buf, nesting, true)
				buf.state = OK

			case co: // {
				if nesting, err = checkNestingDepth(nesting); err != nil {
					return nil, err
				}

				if current, err = createNode(current, buf, Object, useKey()); err != nil {
					return nil, err
				}

				buf.state = OB

			case bo: // [
				if nesting, err = checkNestingDepth(nesting); err != nil {
					return nil, err
				}

				if current, err = createNode(current, buf, Array, useKey()); err != nil {
					return nil, err
				}

				buf.state = AR

			case cm: // ,
				if current == nil {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				if current.IsObject() {
					buf.state = KE // key expected
				} else if current.IsArray() {
					buf.state = VA // value expected
				} else {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

			case cl: // :
				if current == nil || !current.IsObject() || key == nil {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				buf.state = VA

			default:
				return nil, unexpectedTokenError(buf.data, buf.index)
			}
		}

		if buf.step() != nil {
			break
		}

		if _, err = buf.first(); err != nil {
			err = nil
			break
		}
	}

	if current == nil || buf.state != OK {
		return nil, io.EOF
	}

	root := current.root()
	if !root.ready() {
		return nil, io.EOF
	}

	return root, err
}

func isValidContainerType(current *Node, nodeType ValueType) bool {
	switch nodeType {
	case Object:
		return current != nil && current.IsObject() && !current.ready()
	case Array:
		return current != nil && current.IsArray() && !current.ready()
	default:
		return false
	}
}

// getString extracts a string from the buffer and advances the buffer index past the string.
func getString(b *buffer) (*string, error) {
	start := b.index
	if err := b.string(doubleQuote, false); err != nil {
		return nil, err
	}

	value, ok := Unquote(b.data[start:b.index+1], doubleQuote)
	if !ok {
		return nil, unexpectedTokenError(b.data, start)
	}

	return &value, nil
}

func unexpectedTokenError(data []byte, index int) error {
	return errors.New(ufmt.Sprintf("unexpected token at index %d. data %b", index, data))
}

func createNode(current *Node, buf *buffer, nodeType ValueType, key **string) (*Node, error) {
	var err error
	current, err = NewNode(current, buf, nodeType, key)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func updateNode(current *Node, buf *buffer, nesting int, decreaseLevel bool) (*Node, int) {
	current.borders[1] = buf.index + 1

	prev := current.prev
	if prev == nil {
		return current, nesting
	}

	current = prev
	if decreaseLevel {
		nesting--
	}

	return current, nesting
}

func checkNestingDepth(nesting int) (int, error) {
	if nesting >= maxNestingDepth {
		return nesting, errors.New("maximum nesting depth exceeded")
	}

	return nesting + 1, nil
}

func UnmarshalSafe(data []byte) (*Node, error) {
	var safe []byte
	safe = append(safe, data...)
	return Unmarshal(safe)
}
