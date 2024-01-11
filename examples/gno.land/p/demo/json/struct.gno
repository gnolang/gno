package json

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"strconv"

	"gno.land/p/demo/ufmt"
)

const (
	StringType  FieldType = "string"
	IntType     		  = "int"
	FloatType   		  = "float"
	BooleanType 		  = "boolean"
	NilType     		  = "nil"
)

type FieldType string

type Field struct {
	Name  string
	Type  FieldType
	Value interface{}
}

type Struct struct {
	Fields []Field
}

func NewStruct() *Struct {
	return &Struct{}
}

func (s *Struct) AddField(name string, fieldType FieldType, value interface{}) *Struct {
	field := Field{Name: name, Type: fieldType, Value: value}
	s.Fields = append(s.Fields, field)
	return s
}

func (s *Struct) AddStringField(name string, value string) *Struct {
	return s.AddField(name, StringType, value)
}

func (s *Struct) AddIntField(name string, value int) *Struct {
	return s.AddField(name, IntType, value)
}

func (s *Struct) AddFloatField(name string, value float64) *Struct {
	return s.AddField(name, FloatType, value)
}

func (s *Struct) AddBoolField(name string, value bool) *Struct {
	return s.AddField(name, BooleanType, value)
}

func (s *Struct) AddNilField(name string) *Struct {
	return s.AddField(name, NilType, nil)
}

func (s *Struct) Marshal() (string, error) {
	var sb strings.Builder
	sb.WriteString("{")

	for i, field := range s.Fields {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(fmt.Sprintf("\"%s\": ", field.Name))

		if valueStr, err := marshalValue(field.Value); err != nil {
			return "", err
		} else {
		sb.WriteString(valueStr)
    }
	}

	sb.WriteString("}")
	return sb.String(), nil
}

type CustomMarshaller interface {
	MarshalCustom() (string, error)
}

func marshalValue(value interface{}) (string, error) {
	if customVal, ok := value.(CustomMarshaller); ok {
		return customVal.MarshalCustom()
	}

	var sb strings.Builder

	switch v := value.(type) {
	case map[string]interface{}:
		sb.WriteString("{")
		isFirst := true
		for key, value := range v {
			if !isFirst {
				sb.WriteString(", ")
			}
			sb.WriteString(ufmt.Sprintf("\"%s\": ", key))

			if marshaledValue, err := marshalValue(value); err != nil {
				return "", err
			} else {
				sb.WriteString(marshaledValue)
			}

			isFirst = false
		}
		sb.WriteString("}")
	case []interface{}:
		sb.WriteString("[")
		for i, value := range v {
			if i > 0 {
				sb.WriteString(", ")
			}
			if marshaledValue, err := marshalValue(value); err != nil {
				return "", err
			} else {
				sb.WriteString(marshaledValue)
			}
		}
		sb.WriteString("]")
	case string:
		sb.WriteString(strconv.Quote(v))
	case int, int8, int16, int32, int64:
		sb.WriteString(fmt.Sprintf("%d", v))
	case uint, uint8, uint16, uint32, uint64:
		sb.WriteString(fmt.Sprintf("%d", v))
	case float32, float64:
		sb.WriteString(fmt.Sprintf("%g", v))
	case bool:
		sb.WriteString(fmt.Sprintf("%t", v))
	case nil:
		sb.WriteString("null")
	default:
		return "", errors.New("Marshal: unknown data type")
	}

	return sb.String(), nil
}

// Marshal converts the map to JSON string.
func Marshal(data interface{}) (string, error) {
	var sb strings.Builder

	switch v := data.(type) {
	case map[string]interface{}:
		sb.WriteString("{")
		isFirst := true
		for key, value := range v {
			if !isFirst {
				sb.WriteString(", ")
			}
			sb.WriteString(ufmt.Sprintf("\"%s\": ", key))

			if marshaledValue, err := Marshal(value); err != nil {
				return "", err
			} else {
				sb.WriteString(marshaledValue)
			}

			isFirst = false
		}
		sb.WriteString("}")
	case []interface{}:
		sb.WriteString("[")
		for i, value := range v {
			if i > 0 {
				sb.WriteString(", ")
			}
			if marshaledValue, err := Marshal(value); err != nil {
				return "", err
			} else {
				sb.WriteString(marshaledValue)
			}
		}
		sb.WriteString("]")
	case string:
		sb.WriteString(fmt.Sprintf("\"%s\"", v))
	case int, int8, int16, int32, int64:
		sb.WriteString(fmt.Sprintf("%d", v))
	case uint, uint8, uint16, uint32, uint64:
		sb.WriteString(fmt.Sprintf("%d", v))
	case float32, float64:
		sb.WriteString(fmt.Sprintf("%f", v))
	case bool:
		sb.WriteString(fmt.Sprintf("%t", v))
	case nil:
		sb.WriteString("null")
	default:
		return "", errors.New("Marshal: unknown data type")
	}

	return sb.String(), nil
}

func Unmarshal(data []byte, s *Struct) error {
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
        default:
            return fmt.Errorf("unsupported type: %v", dataType)
        }
    }
    return nil
}
