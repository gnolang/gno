package json

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gno.land/p/demo/ufmt"
)

// Tatget Struct 
// -> Extract Fields and Save them in a FieldMetaData Struct
// -> Search Values from JSON based on the Extracted Fields via Get method
// -> Zip the Fields with the Values
// -> Return the Zipped Data

// TODO: current regex approach is not robust and flexible. So, it needs to be improved while maintaining the recursive descent approach.
const fieldRegex = `(\w+)\s+(\*?\w+|\[\]\w+|struct\s*{[^}]*}|map\[\w+\]\w+|interface{})\s+` + "`" + `json:"([^"]+)"` + "`"

type Fields []FieldMetaData

type FieldMetaData struct {
	Name       string
	Tag        string
	Type       string
	IsNested   bool
	Depth      int
}

func ParseStruct(data string, depth int) (fields Fields) {
	// separate fields using regular expressions
	re := regexp.MustCompile(fieldRegex)
	matches := re.FindAllStringSubmatch(data, -1)

	for _, match := range matches {
		field := FieldMetaData{
			Name:  match[1],
			Tag:   match[3],
			Depth: depth,
		}

		// check if the field is a nested struct
		fieldType := match[2]
		field = field.checkFieldType(fieldType)
		fields = append(fields, field)

		// apply recursion when the field is a nested structure
		if field.IsNested {
			nestedData := match[2]
			nestedFields := ParseStruct(nestedData, depth+1)
			fields = append(fields, nestedFields...)
		}
	}

	return fields
}

func (f FieldMetaData) checkFieldType(hayStack string) FieldMetaData {
	switch {
	case strings.HasPrefix(hayStack, StructLiteral):
		f.Type = "struct"
		f.IsNested = true
	case strings.HasPrefix(hayStack, MapLiteral):
		f.Type = "map"
		f.IsNested = false
		panic("ParseStruct: map type not supported")
	case strings.HasPrefix(hayStack, InterfaceLiteral):
		f.Type = "interface"
		panic("ParseStruct: interface type not supported")
	default:
		f.Type = hayStack
		f.IsNested = false
	}

	return f
}

// Names returns the names of the fields
// which will used to search the values from JSON.
func (fs *Fields) Names() (names []string) {
	for _, f := range *fs {
		names = append(names, f.Name)
	}

	return names
}

// TODO: support for complex types
// Zip zips the fields with the values from JSON.
func Zip(json []byte, extractedKeys ...string) (zippedData map[string]interface{}, err error) {
	zippedData = make(map[string]interface{})

	for _, key := range extractedKeys {
		value, dataType, offset, err := Get(json, key)
		if err != nil {
			return nil, err
		}

		switch dataType {
		case String:
			zippedData[key] = string(value)
		case Number:
			if n, err := ParseFloatLiteral(value); err != nil {
				return nil, err
			} else {
				zippedData[key] = n
			}
		case Boolean:
			if b, err := ParseBoolLiteral(value); err != nil {
				return nil, err
			} else {
				zippedData[key] = b
			}
		case Array:
			panic("Zip: TODO array type not supported")
		case Object:
			panic("Zip: TODO object type not supported")
		case Null:
			zippedData[key] = nil
		default:
			return nil, errors.New("Zip: unknown data type")
		}
	}

	return zippedData, nil
}

func Marshal(data map[string]interface{}) (string, error) {
	var sb strings.Builder
	sb.WriteString("{")

	isFirst := true
	for key, value := range data {
		if !isFirst {
			sb.WriteString(", ")
		}
		sb.WriteString(ufmt.Sprintf("\"%s\": ", key))

		switch v := value.(type) {
		case string:
			sb.WriteString(ufmt.Sprintf("\"%s\"", v))
		case float32:
			sb.WriteString(ufmt.Sprintf("%f", v))
		case bool:
			sb.WriteString(ufmt.Sprintf("%t", v))
		default:
			return "", errors.New("Marshal: unknown data type")
		}

		isFirst = false
	}

	sb.WriteString("}")
	return sb.String(), nil
}