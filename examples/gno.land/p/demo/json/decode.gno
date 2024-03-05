package json

import (
	"errors"
	"io"

	"gno.land/p/demo/ufmt"
)

func unmarshal2(data []byte) (*Node, error) {
	buf := NewBuffer(data)

	var (
		state   States
		key     string
		current *Node
		hasKey  bool
	)

	_, err := buf.first()
	if err != nil {
		return nil, io.EOF
	}

	for {
		state = buf.getState()
		if state == __ {
			return nil, unexpectedTokenError(buf.data, buf.index)
		}

		if state >= GO {
			switch buf.state {
			case ST:
				if current != nil && current.isObject() && !hasKey {
					// key detected
					key, err = getString(buf)
					hasKey = true
					buf.state = CO
				} else {
					// string detected
					err = createOrUpdateNode(&current, buf, String, &hasKey, &key)
					if err != nil {
						return nil, err
					}

					if err = buf.string(DoublyQuoteToken, false); err != nil {
						return nil, err
					}

					current.borders[1] = buf.index + 1
					buf.state = OK

					updateCurrentNode(&current)
				}

			case MI, ZE, IN:
				if err = createOrUpdateNode(&current, buf, Number, &hasKey, &key); err != nil {
					return nil, err
				}

				if err = buf.numeric(false); err != nil {
					return nil, err
				}

				current.borders[1] = buf.index
				buf.index -= 1
				buf.state = OK

				updateCurrentNode(&current)

			case T1, F1:
				if err = createOrUpdateNode(&current, buf, Boolean, &hasKey, &key); err != nil {
					return nil, err
				}

				if buf.state == T1 {
					err = buf.word(trueLiteral)
				} else {
					err = buf.word(falseLiteral)
				}

				if err != nil {
					return nil, err
				}

				current.borders[1] = buf.index + 1
				buf.state = OK

				updateCurrentNode(&current)

			case N1:
				if err = createOrUpdateNode(&current, buf, Null, &hasKey, &key); err != nil {
					return nil, err
				}

				if err = buf.word(nullLiteral); err != nil {
					return nil, err
				}

				current.borders[1] = buf.index + 1
				buf.state = OK

				updateCurrentNode(&current)
			}
		} else {
			switch state {
			case ec, cc:
				if current != nil && current.isObject() && !current.ready() {
					current.borders[1] = buf.index + 1
					updateCurrentNode(&current)
				} else {
					if err = unexpectedTokenError(buf.data, buf.index); err != nil {
						return nil, err
					}
				}

				buf.state = OK

			case bc:
				if current != nil && current.isArray() && !current.ready() {
					current.borders[1] = buf.index + 1
					updateCurrentNode(&current)
				} else {
					if err = unexpectedTokenError(buf.data, buf.index); err != nil {
						return nil, err
					}
				}

				buf.state = OK

			case co:
				if err = createOrUpdateNode(&current, buf, Object, &hasKey, &key); err != nil {
					return nil, err
				}

				buf.state = OB

			case bo:
				if err = createOrUpdateNode(&current, buf, Array, &hasKey, &key); err != nil {
					return nil, err
				}

				buf.state = AR

			case cm:
				if current == nil {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				if current.isObject() {
					buf.state = KE // key expected
				} else if current.isArray() {
					buf.state = VA // value expected
				} else {
					if err = unexpectedTokenError(buf.data, buf.index); err != nil {
						return nil, err
					}
				}

			case cl:
				if current == nil || !current.isObject() || !hasKey {
					if err = unexpectedTokenError(buf.data, buf.index); err != nil {
						return nil, err
					}
				}

				buf.state = VA

			default:
				if err = unexpectedTokenError(buf.data, buf.index); err != nil {
					return nil, err
				}
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

	return checkRootNode(current, buf)
}

func UnmarshalSafe(data []byte) (*Node, error) {
	var safe []byte
	safe = append(safe, data...)
	return unmarshal2(safe)
}

func Must(root *Node, expect error) *Node {
	if expect != nil {
		panic(expect)
	}

	return root
}

// TODO: improve this function to avoid pointer dereferencing
func createOrUpdateNode(curr **Node, buf *buffer, t ValueType, hasKey *bool, key *string) error {
	var err error

	if *hasKey {
		*curr, err = NewNode(*curr, buf, t, key)
		*hasKey = false
	} else {
		*curr, err = NewNode(*curr, buf, t, nil)
	}

	return err
}

// getString extracts a string from the buffer and advances the buffer index past the string.
func getString(b *buffer) (string, error) {
	if err := b.string(DoublyQuoteToken, false); err != nil {
		return "", err
	}

	start := b.index
	value, err := unquote(b.data[start:start+1], DoublyQuoteToken)
	if err != nil {
		return "", errors.New(ufmt.Sprintf("token %s is not a valid string", b.data[start:start+1]))
	}

	return value, nil
}

func updateCurrentNode(current **Node) {
	if (*current).prev != nil {
		*current = (*current).prev
	}
}

func checkRootNode(current *Node, buf *buffer) (*Node, error) {
	if current == nil || buf.state != OK {
		return nil, io.EOF
	}

	root := current.root()
	if !root.ready() {
		return nil, io.EOF
	}

	return root, nil
}

func unexpectedTokenError(data []byte, index uint) error {
    return errors.New(ufmt.Sprintf("unexpected token %s", data[index:index+1]))
}