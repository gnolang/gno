package json

import (
	"regexp"
	"strings"
)

type FieldMetaData struct {
	Name       string
	Tag        string
	Type       string
	IsNested   bool
	IsOptional bool
	Depth      int
}

func ParseStruct(data []byte, depth int) []FieldMetaData {
	var fields []FieldMetaData
	dataStr := string(data)

	// separate fields using regular expressions
	re := regexp.MustCompile(`(\w+)\s+(\*?\w+|\[\]\w+|struct\s*{[^}]*})\s+` + "`" + `json:"([^"]+)"` + "`")
	matches := re.FindAllStringSubmatch(dataStr, -1)

	for _, match := range matches {
		field := FieldMetaData{
			Name:  match[1],
			Tag:   match[3],
			Depth: depth,
		}

		// check if the field is a nested struct
		// if it is, set the type to "struct" literal and set IsNested to true
		if strings.HasPrefix(match[2], "struct") {
			field.Type = "struct"
			field.IsNested = true
		} else {
			field.Type = match[2]
			field.IsNested = false
		}

		field.IsOptional = strings.Contains(field.Tag, "omitempty")

		fields = append(fields, field)

		// if the field is a nested struct, parse it recursively
		if field.IsNested {
			nestedData := []byte(match[2])
			nestedFields := ParseStruct(nestedData, depth+1)
			fields = append(fields, nestedFields...)
		}
	}

	return fields
}
