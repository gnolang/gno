// ref: https://github.com/spyzhov/ajson/blob/master/decode.go

package json

import (
	"errors"
	"io"

	"gno.land/p/nt/ufmt/v0"
)

// This limits the max nesting depth to prevent stack overflow.
// This is permitted by https://tools.ietf.org/html/rfc7159#section-9
const maxNestingDepth = 10000

// Unmarshal parses the JSON-encoded data and returns a Node.
// The data must be a valid JSON-encoded value.
//
// Usage:
//
//	node, err := json.Unmarshal([]byte(`{"key": "value"}`))
//	if err != nil {
//		ufmt.Println(err)
//	}
//	println(node) // {"key": "value"}
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
					current, nesting, err = createNestedNode(current, buf, String, nesting, useKey())
					if err != nil {
						return nil, err
					}

					err = buf.string(doubleQuote, false)
					if err != nil {
						return nil, err
					}

					current, nesting = updateNode(current, buf, nesting, true)
					buf.state = OK
				}

			case MI, ZE, IN: // number
				current, err = processNumericNode(current, buf, useKey())
				if err != nil {
					return nil, err
				}

			case T1, F1: // boolean
				literal := falseLiteral
				if buf.state == T1 {
					literal = trueLiteral
				}

				current, nesting, err = processLiteralNode(current, buf, Boolean, literal, useKey(), nesting)
				if err != nil {
					return nil, err
				}

			case N1: // null
				current, nesting, err = processLiteralNode(current, buf, Null, nullLiteral, useKey(), nesting)
				if err != nil {
					return nil, err
				}
			}
		} else {
			// region action
			switch state {
			case ec, cc: // <empty> }
				if key != nil {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				current, nesting, err = updateNodeAndSetBufferState(current, buf, nesting, Object)
				if err != nil {
					return nil, err
				}

			case bc: // ]
				current, nesting, err = updateNodeAndSetBufferState(current, buf, nesting, Array)
				if err != nil {
					return nil, err
				}

			case co, bo: // { [
				valTyp, bState := Object, OB
				if state == bo {
					valTyp, bState = Array, AR
				}

				current, nesting, err = createNestedNode(current, buf, valTyp, nesting, useKey())
				if err != nil {
					return nil, err
				}

				buf.state = bState

			case cm: // ,
				if current == nil {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				if !current.isContainer() {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				if current.IsObject() {
					buf.state = KE // key expected
				} else {
					buf.state = VA // value expected
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

// UnmarshalSafe parses the JSON-encoded data and returns a Node.
func UnmarshalSafe(data []byte) (*Node, error) {
	var safe []byte
	safe = append(safe, data...)
	return Unmarshal(safe)
}

// processNumericNode creates a new node, processes a numeric value,
// sets the node's borders, and moves to the previous node.
func processNumericNode(current *Node, buf *buffer, key **string) (*Node, error) {
	var err error
	current, err = createNode(current, buf, Number, key)
	if err != nil {
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

	return current, nil
}

// processLiteralNode creates a new node, processes a literal value,
// sets the node's borders, and moves to the previous node.
func processLiteralNode(
	current *Node,
	buf *buffer,
	literalType ValueType,
	literalValue []byte,
	useKey **string,
	nesting int,
) (*Node, int, error) {
	var err error
	current, nesting, err = createLiteralNode(current, buf, literalType, literalValue, useKey, nesting)
	if err != nil {
		return nil, nesting, err
	}
	return current, nesting, nil
}

// isValidContainerType checks if the current node is a valid container (object or array).
// The container must satisfy the following conditions:
//  1. The current node must not be nil.
//  2. The current node must be an object or array.
//  3. The current node must not be ready.
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

// createNode creates a new node and sets the key if it is not nil.
func createNode(
	current *Node,
	buf *buffer,
	nodeType ValueType,
	key **string,
) (*Node, error) {
	var err error
	current, err = NewNode(current, buf, nodeType, key)
	if err != nil {
		return nil, err
	}

	return current, nil
}

// createNestedNode creates a new nested node (array or object) and sets the key if it is not nil.
func createNestedNode(
	current *Node,
	buf *buffer,
	nodeType ValueType,
	nesting int,
	key **string,
) (*Node, int, error) {
	var err error
	if nesting, err = checkNestingDepth(nesting); err != nil {
		return nil, nesting, err
	}

	if current, err = createNode(current, buf, nodeType, key); err != nil {
		return nil, nesting, err
	}

	return current, nesting, nil
}

// createLiteralNode creates a new literal node and sets the key if it is not nil.
// The literal is a byte slice that represents a boolean or null value.
func createLiteralNode(
	current *Node,
	buf *buffer,
	literalType ValueType,
	literal []byte,
	useKey **string,
	nesting int,
) (*Node, int, error) {
	var err error
	if current, err = createNode(current, buf, literalType, useKey); err != nil {
		return nil, 0, err
	}

	if err = buf.word(literal); err != nil {
		return nil, 0, err
	}

	current, nesting = updateNode(current, buf, nesting, false)
	buf.state = OK

	return current, nesting, nil
}

// updateNode updates the current node and returns the previous node.
func updateNode(
	current *Node, buf *buffer, nesting int, decreaseLevel bool,
) (*Node, int) {
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

// updateNodeAndSetBufferState updates the current node and sets the buffer state to OK.
func updateNodeAndSetBufferState(
	current *Node,
	buf *buffer,
	nesting int,
	typ ValueType,
) (*Node, int, error) {
	if !isValidContainerType(current, typ) {
		return nil, nesting, unexpectedTokenError(buf.data, buf.index)
	}

	current, nesting = updateNode(current, buf, nesting, true)
	buf.state = OK

	return current, nesting, nil
}

// checkNestingDepth checks if the nesting depth is within the maximum allowed depth.
func checkNestingDepth(nesting int) (int, error) {
	if nesting >= maxNestingDepth {
		return nesting, errors.New("maximum nesting depth exceeded")
	}

	return nesting + 1, nil
}

func unexpectedTokenError(data []byte, index int) error {
	return ufmt.Errorf("unexpected token at index %d. data %b", index, data)
}
