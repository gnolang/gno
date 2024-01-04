package json

import (
	"strings"
	"unicode"
)

type FieldMetaData struct {
	Name		string
	Tag			string
	Type		string
	IsNested  	bool
	IsOptional 	bool
	Depth		int
}

func ParseStruct(b []byte) []FieldMetaData {
	return scanFields(b, 0)
}

func scanFields(b []byte, depth int) []FieldMetaData {
	l := len(b)
	fields := []FieldMetaData

	if l < 2 {
		return nil
	}

	for i := 0; i < l; i++ {
		// end of current struct
		if b[i] == CurlyCloseToken {
			break
		}

		if unicode.IsLetter(rune(b[i])) {
			fd, newIndex := parseField(b, i, depth)
			fields = append(fields, fd)
			i = newIndex
		}

		if isStructIdent(b, i) {
			// Skip to the opening brace
			for i < l && b[i] != CurlyOpenToken {
				i++
			}

			if i < l {
				nested := parseFields(b[i+1:], depth+1)
				fields = append(fields, nested...)
			}
		}
	}

	return fields
}

func parseField(b []byte, start, depth int) (FieldMetaData, int) {
	field := FieldMetaData{Depth: depth}
	i := start
	l := len(b)

	// parse field name
	for i < l && b[i] != WhiteSpaceToken {
		i++
	}
	field.Name = string(b[start:i])

	// Skip to JSON tag
	for i < l && b[i] != BackTickToken {
		i++
	}

	if i < l {
		field.Tag, i	 = parseTag(b, i+1)
		field.IsOptional = isOptional(field.Tag)
	}

	return field, i
}

func parseTag(b []byte, start int) (string, int) {
	i := start
	if i+6 >= len(b) {
		return "", i
	}

	i += 6 // skip `json:"`
	start = i

	for i < len(b) && b[i] != DoublyQuoteToken {
		i++
	}

	return string(b[start:i]), i+1
}

func isOptional(tag string) bool {
	return strings.Contains(tag, ",omitempty")
}

func isStructIdent(b []byte, start int) bool {
	return strings.Contains(string(b[start:]), "struct")
}