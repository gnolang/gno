package json

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"strconv"

	"gno.land/p/demo/ufmt"
)

type FieldType string

// Field stores each field's information.
type Field struct {
	Name  		 	string
	Type  		 	ValueType
	Value 		 	interface{}
	ArrayValues  	[]interface{}
	NestedStruct	*Struct
}

// Struct holds the whole structure of target struct.
type Struct struct {
	Fields []Field
}

// NewStruct create new empty struct placeholder.
func NewStruct() *Struct {
	return &Struct{}
}

// addField adds a field to the struct. The order of the fields is preserved.
func (s *Struct) addField(name string, fieldType ValueType, value interface{}) *Struct {
	field := Field{Name: name, Type: fieldType, Value: value}
	s.Fields = append(s.Fields, field)
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
	f := Field {
		Name: name,
		Type: Object,
		NestedStruct: inner,
	}
	s.Fields = append(s.Fields, f)

	return s
}

func (s *Struct) AddArrayField(name string, array []interface{}) *Struct {
	f := Field {
		Name: name,
		Type: Array,
		ArrayValues: array,
	}
	s.Fields = append(s.Fields, f)

	return s
}

// Marshaler marshals the struct into a JSON byte slice
func (s *Struct) Marshaler() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, field := range s.Fields {
		// separate fields with comma
		if i > 0 {
			buf.WriteByte(',')
			buf.WriteByte(' ')
		}

		// `"name": `
		buf.WriteString(strconv.Quote(field.Name))
		buf.WriteByte(':')
		buf.WriteByte(' ')

		if field.NestedStruct != nil {
			if nestedJSON, err := field.NestedStruct.Marshaler(); err != nil {
				return nil, err
			} else {
				buf.Write(nestedJSON)
			}
		} else if len(field.ArrayValues) > 0 {
			buf.WriteByte('[')
			for i, value := range field.ArrayValues {
				if i > 0 {
					buf.WriteByte(',')
					buf.WriteByte(' ')
				}
				if valueStr, err := marshalValue(value); err != nil {
					return nil, err
				} else {
					buf.Write(valueStr)
				}
			}
			buf.WriteByte(']')
		} else {
			if valueStr, err := marshalValue(field.Value); err != nil {
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
				buf.WriteByte(',')
				buf.WriteByte(' ')
			}

			buf.WriteString(strconv.Quote(key))
			buf.WriteByte(':')
			buf.WriteByte(' ')

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
				buf.WriteByte(',')
				buf.WriteByte(' ')
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
		buf.WriteByte('n')
		buf.WriteByte('u')
		buf.WriteByte('l')
		buf.WriteByte('l')
	default:
		return nil, errors.New("Marshal: unknown data type")
	}

	return buf.Bytes(), nil
}

// Unmarshaler unmarshals the JSON string into the struct.
func Unmarshaler(data []byte, s *Struct) error {
    for _, field := range s.Fields {
        value, dataType, _, err := Get(data, field.Name)
        if err != nil {
            return err
        }

        switch dataType {
        case String:
            field.Value = string(value)
        case Number, Float:
			if n, err := ParseFloatLiteral(value); err != nil {
				return err
			} else {
				field.Value = n
			}
        case Boolean:
			if b, err := ParseBoolLiteral(value); err != nil {
				return err
			} else {
				field.Value = b
			}
		case Object:
			if field.NestedStruct != nil {
				if val, _, _, err := Get(data, field.Name); err != nil {
					return err
				} else {
					if err = Unmarshaler(val, field.NestedStruct); err != nil {
						return err
					}
				}
			}
        default:
            return fmt.Errorf("json.Unmarshaler: unsupported type: %v", dataType)
        }
    }
    return nil
}
