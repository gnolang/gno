package json

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
)

// Field stores each field's information.
type Field struct {
	// name holds the field name.
	name string
	// typ holds the field type.
	typ ValueType
	// value holds the field value.
	value interface{}
	// arrayValues store the each value of the array.
	arrayValues []interface{}
	// nestedStruct represents the JSON object.
	nestedStruct *Struct
}

// Struct holds the whole structure of target struct.
type Struct struct {
	// fields holds the fields of the struct.
	fields []Field
	// TODO: may use map to speed up the search.
	// fmap   map[string]*Field
}

// NewStruct create new empty struct placeholder.
func NewStruct() *Struct {
	return &Struct{}
}

// addField adds a field to the struct. The order of the fields is preserved.
func (s *Struct) addField(name string, fieldType ValueType, value interface{}) *Struct {
	field := Field{name: name, typ: fieldType, value: value}
	s.fields = append(s.fields, field)
	return s
}

func (s *Struct) AddStringField(name string, value string) *Struct {
	return s.addField(name, String, value)
}

func (s *Struct) AddIntField(name string, value int) *Struct {
	return s.addField(name, Number, value)
}

func (s *Struct) AddFloatField(name string, value float64) *Struct {
	return s.addField(name, Float, value)
}

func (s *Struct) AddBoolField(name string, value bool) *Struct {
	return s.addField(name, Boolean, value)
}

func (s *Struct) AddNullField(name string) *Struct {
	return s.addField(name, Null, nil)
}

func (s *Struct) AddObjectField(name string, inner *Struct) *Struct {
	f := Field{
		name:         name,
		typ:          Object,
		nestedStruct: inner,
	}
	s.fields = append(s.fields, f)

	return s
}

func (s *Struct) AddArrayField(name string, array []interface{}) *Struct {
	f := Field{
		name:        name,
		typ:         Array,
		arrayValues: array,
	}
	s.fields = append(s.fields, f)

	return s
}

/*           Encoder              */

// Marshaler marshals the struct into a JSON byte slice.
func (s *Struct) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, field := range s.fields {
		if i > 0 {
			buf.WriteString(", ")
		}

		// `"name": `
		buf.WriteString(strconv.Quote(field.name))
		buf.WriteString(": ")

		if field.nestedStruct != nil {
			// nested struct are considered as object.
			// need to apply marshaler to each field.
			if nestedJSON, err := field.nestedStruct.Marshal(); err != nil {
				return nil, err
			} else {
				buf.Write(nestedJSON)
			}
		} else if len(field.arrayValues) > 0 {
			buf.WriteByte('[')
			for i, value := range field.arrayValues {
				if i > 0 {
					buf.WriteString(", ")
				}
				// apply marshaler to each value
				if valueStr, err := marshalValue(value); err != nil {
					return nil, err
				} else {
					buf.Write(valueStr)
				}
			}
			buf.WriteByte(']')
		} else {
			if valueStr, err := marshalValue(field.value); err != nil {
				return nil, err
			} else {
				buf.Write(valueStr)
			}
		}
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// CustomMarshaller interface that the custom type should implement.
type CustomMarshaller interface {
	MarshalCustom() ([]byte, error)
}

// marshalValue marshals the value by its type.
func marshalValue(value interface{}) ([]byte, error) {
	if customVal, ok := value.(CustomMarshaller); ok {
		if marshaled, err := customVal.MarshalCustom(); err != nil {
			return nil, err
		} else {
			return marshaled, nil
		}
	}

	var buf bytes.Buffer

	switch v := value.(type) {
	// Object
	case map[string]interface{}:
		buf.WriteByte('{')
		isFirst := true
		for key, value := range v {
			if !isFirst {
				buf.WriteString(", ")
			}

			// `"<key>": `
			buf.WriteString(strconv.Quote(key))
			buf.WriteString(": ")

			if marshaledValue, err := marshalValue(value); err != nil {
				return nil, err
			} else {
				buf.Write(marshaledValue)
			}

			isFirst = false
		}
		buf.WriteByte('}')
	// Array
	case []interface{}:
		buf.WriteByte('[')
		for i, value := range v {
			if i > 0 {
				buf.WriteString(", ")
			}
			if marshaledValue, err := marshalValue(value); err != nil {
				return nil, err
			} else {
				buf.Write(marshaledValue)
			}
		}
		buf.WriteByte(']')
	case string:
		// must handle escape characters
		quoted := strconv.Quote(v)
		buf.WriteString(quoted)
	case int, int8, int16, int32, int64:
		buf.WriteString(fmt.Sprintf("%d", v))
	case uint, uint8, uint16, uint32, uint64:
		buf.WriteString(fmt.Sprintf("%d", v))
	case float32, float64:
		buf.WriteString(fmt.Sprintf("%g", v))
	case bool:
		buf.WriteString(fmt.Sprintf("%t", v))
	case nil:
		buf.WriteString("null")
	default:
		return nil, errors.New(fmt.Sprintf("json.Marshal: unsupported type %s", value))
	}

	return buf.Bytes(), nil
}

/*              Decoder              */

// Unmarshal unmarshals the JSON data into the struct.
// The struct should be declared as a Struct instance.
func Unmarshal(data []byte, s *Struct) error {
	keys, err := extractKeysFromJSON(data)
	if err != nil {
		return err
	}

	for _, key := range keys {
		// regex-parsed key contains double quotes, so need to remove them.
		// this is unnecessary if we don't use regex when retrieving keys.
		key = strings.ReplaceAll(key, `"`, "")
		val, typ, _, err := Get(data, key)
		if err != nil {
			return err
		}

		switch typ {
		case String:
			s.AddStringField(key, string(val))
		case Number, Float:
			if got, err := ParseFloatLiteral(val); err != nil {
				return err
			} else {
				s.AddFloatField(key, got)
			}
		case Boolean:
			if got, err := ParseBoolLiteral(val); err != nil {
				return err
			} else {
				s.AddBoolField(key, got)
			}
		case Null:
			s.AddNullField(key)
		// TODO: Support Array and Object types
		default:
			return errors.New("json.Unmarshal: unknown data type")
		}
	}

	return nil
}

func extractKeysFromJSON(data []byte) ([]string, error) {
	// parse the JSON key names.
	// TODO: reduce overhead. regex is easy to use but slow.
	re := regexp.MustCompile(`"([^"\\]|\\.)*"\s*:`)

	// check if the JSON is valid.
	if !re.Match(data) {
		return nil, errors.New("Invalid JSON format")
	}

	var keys []string
	matches := re.FindAllSubmatch(data, -1)
	for _, match := range matches {
		key := string(match[0])

		// reject empty keys
		if key == "" {
			return nil, errors.New("json: key cannot be an empty string")
		}

		key = strings.TrimSuffix(key, ":")
		key = strings.TrimSpace(key)
		keys = append(keys, key)
	}

	return keys, nil
}

// String apply the marshaler to the struct and return the JSON string.
func (s *Struct) String() string {
	if json, err := s.Marshal(); err != nil {
		return err.Error()
	} else {
		return string(json)
	}
}

/*           CRUD          */
// CRUD operations for the struct.
// by using this methods, we can modify the struct also the JSON data easily.

// Search finds a field by name and returns it.
func (s *Struct) Search(name string) (*Field, error) {
	for i, field := range s.fields {
		if field.name == name {
			return &s.fields[i], nil
		}
	}

	return nil, errors.New(fmt.Sprintf("json.Search: Field name %s not found", name))
}

// Remove removes a field from the struct instance.
func (s *Struct) Remove(name string) error {
	for i, field := range s.fields {
		if field.name == name {
			// arranges the slice to remove the field.
			s.fields = append(s.fields[:i], s.fields[i+1:]...)
			break
		}
	}

	return errors.New(fmt.Sprintf("json.Search: Field name %s not found", name))
}

// Update updates the value of the field.
//
// The updated value must be the same type as the original one.
// Otherwise, it will return an error.
func (s *Struct) Update(name string, newVal interface{}) error {
	// search field by name
	f, err := s.Search(name)
	if err != nil {
		return errors.New(fmt.Sprintf("json.Update: Field name %s not found", name))
	}

	// type assertion
	newValType := getTypeFromValue(newVal)
	if f.typ != newValType {
		return errors.New(
			fmt.Sprintf("json.Update: Field type mismatch. Expected %s, got %s",
				f.typ.String(),
				newValType.String(),
			))
	}

	// update the value
	f.value = newVal
	return nil
}
