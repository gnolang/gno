package json

import (
	"bytes"
	"errors"
	"math"
	"strconv"

	"gno.land/p/demo/json/ryu"
	"gno.land/p/demo/ufmt"
)

// Marshal returns the JSON encoding of a Node.
func Marshal(node *Node) ([]byte, error) {
	var (
		buf  bytes.Buffer
		sVal string
		bVal bool
		nVal float64
		oVal []byte
		err  error
	)

	if node == nil {
		return nil, errors.New("node is nil")
	}

	if !node.modified && !node.ready() {
		return nil, errors.New("node is not ready")
	}

	if !node.modified && node.ready() {
		buf.Write(node.source())
	}

	if node.modified {
		switch node.nodeType {
		case Null:
			buf.Write(nullLiteral)

		case Number:
			nVal, err = node.GetNumeric()
			if err != nil {
				return nil, err
			}

			// ufmt does not support %g. by doing so, we need to check if the number is an integer
			// after then, apply the correct format for each float and integer numbers.
			if math.Mod(nVal, 1.0) == 0 {
				// must convert float to integer. otherwise it will be overflowed.
				num := ufmt.Sprintf("%d", int(nVal))
				buf.WriteString(num)
			} else {
				// use ryu algorithm to convert float to string
				num := ryu.FormatFloat64(nVal)
				buf.WriteString(num)
			}

		case String:
			sVal, err = node.GetString()
			if err != nil {
				return nil, err
			}

			quoted := ufmt.Sprintf("%s", strconv.Quote(sVal))
			buf.WriteString(quoted)

		case Boolean:
			bVal, err = node.GetBool()
			if err != nil {
				return nil, err
			}

			bStr := ufmt.Sprintf("%t", bVal)
			buf.WriteString(bStr)

		case Array:
			buf.WriteByte(bracketOpen)

			for i := 0; i < len(node.next); i++ {
				if i != 0 {
					buf.WriteByte(comma)
				}

				elem, ok := node.next[strconv.Itoa(i)]
				if !ok {
					return nil, ufmt.Errorf("array element %d is not found", i)
				}

				oVal, err = Marshal(elem)
				if err != nil {
					return nil, err
				}

				buf.Write(oVal)
			}

			buf.WriteByte(bracketClose)

		case Object:
			buf.WriteByte(curlyOpen)

			bVal = false
			for k, v := range node.next {
				if bVal {
					buf.WriteByte(comma)
				} else {
					bVal = true
				}

				key := ufmt.Sprintf("%s", strconv.Quote(k))
				buf.WriteString(key)
				buf.WriteByte(colon)

				oVal, err = Marshal(v)
				if err != nil {
					return nil, err
				}

				buf.Write(oVal)
			}

			buf.WriteByte(curlyClose)
		}
	}

	return buf.Bytes(), nil
}
