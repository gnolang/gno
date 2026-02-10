package json

import (
	"bytes"
	"strings"
)

// indentGrowthFactor specifies the growth factor of indenting JSON input.
// A factor no higher than 2 ensures that wasted space never exceeds 50%.
const indentGrowthFactor = 2

// IndentJSON formats the JSON data with the specified indentation.
func Indent(data []byte, indent string) ([]byte, error) {
	var (
		out        bytes.Buffer
		level      int
		inArray    bool
		arrayDepth int
	)

	for i := 0; i < len(data); i++ {
		c := data[i] // current character

		switch c {
		case bracketOpen:
			arrayDepth++
			if arrayDepth > 1 {
				level++ // increase the level if it's nested array
				inArray = true

				if err := out.WriteByte(c); err != nil {
					return nil, err
				}

				if err := writeNewlineAndIndent(&out, level, indent); err != nil {
					return nil, err
				}
			} else {
				// case of the top-level array
				inArray = true
				if err := out.WriteByte(c); err != nil {
					return nil, err
				}
			}

		case bracketClose:
			if inArray && arrayDepth > 1 { // nested array
				level--
				if err := writeNewlineAndIndent(&out, level, indent); err != nil {
					return nil, err
				}
			}

			arrayDepth--
			if arrayDepth == 0 {
				inArray = false
			}

			if err := out.WriteByte(c); err != nil {
				return nil, err
			}

		case curlyOpen:
			// check if the empty object or array
			// we don't need to apply the indent when it's empty containers.
			if i+1 < len(data) && data[i+1] == curlyClose {
				if err := out.WriteByte(c); err != nil {
					return nil, err
				}

				i++ // skip next character
				if err := out.WriteByte(data[i]); err != nil {
					return nil, err
				}
			} else {
				if err := out.WriteByte(c); err != nil {
					return nil, err
				}

				level++
				if err := writeNewlineAndIndent(&out, level, indent); err != nil {
					return nil, err
				}
			}

		case curlyClose:
			level--
			if err := writeNewlineAndIndent(&out, level, indent); err != nil {
				return nil, err
			}
			if err := out.WriteByte(c); err != nil {
				return nil, err
			}

		case comma, colon:
			if err := out.WriteByte(c); err != nil {
				return nil, err
			}
			if inArray && arrayDepth > 1 { // nested array
				if err := writeNewlineAndIndent(&out, level, indent); err != nil {
					return nil, err
				}
			} else if c == colon {
				if err := out.WriteByte(' '); err != nil {
					return nil, err
				}
			}

		default:
			if err := out.WriteByte(c); err != nil {
				return nil, err
			}
		}
	}

	return out.Bytes(), nil
}

func writeNewlineAndIndent(out *bytes.Buffer, level int, indent string) error {
	if err := out.WriteByte('\n'); err != nil {
		return err
	}

	idt := strings.Repeat(indent, level*indentGrowthFactor)
	if _, err := out.WriteString(idt); err != nil {
		return err
	}

	return nil
}
