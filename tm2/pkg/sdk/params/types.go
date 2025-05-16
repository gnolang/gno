package params

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Param struct {
	Key   string
	Type  string
	Value any
}

const (
	ParamTypeString  = "string"
	ParamTypeInt64   = "int64"
	ParamTypeUint64  = "uint64"
	ParamTypeBool    = "bool"
	ParamTypeBytes   = "bytes"
	ParamTypeStrings = "strings"
)

// NOTE: do not support structs here.
func NewParam(key string, value any) Param {
	switch value.(type) {
	case string:
		return Param{Key: key, Type: ParamTypeString, Value: value}
	case int64:
		return Param{Key: key, Type: ParamTypeInt64, Value: value}
	case uint64:
		return Param{Key: key, Type: ParamTypeUint64, Value: value}
	case bool:
		return Param{Key: key, Type: ParamTypeBool, Value: value}
	case []byte:
		return Param{Key: key, Type: ParamTypeBytes, Value: value}
	case []string:
		return Param{Key: key, Type: ParamTypeStrings, Value: value}
	default:
		panic(fmt.Sprintf("unexpected param value type %v", reflect.TypeOf(value)))
	}
}

func (p Param) ValidateBasic() error {
	// XXX: validate type and value
	return nil
}

// As would appear in genesis.json.
func (p *Param) Parse(entry string) error {
	parts := strings.SplitN(strings.TrimSpace(entry), "=", 2) // <key>.<type>=<value>
	if len(parts) != 2 {
		return fmt.Errorf("malformed entry: %q", entry)
	}

	keyWithType := parts[0]
	rawValue := parts[1]
	p.Type = keyWithType[strings.LastIndex(keyWithType, ".")+1:]
	p.Key = strings.TrimSuffix(keyWithType, "."+p.Type)
	switch p.Type {
	case ParamTypeString:
		p.Value = rawValue
	case ParamTypeInt64:
		v, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return err
		}
		p.Value = v
	case ParamTypeBool:
		v, err := strconv.ParseBool(rawValue)
		if err != nil {
			return err
		}
		p.Value = v
	case ParamTypeUint64:
		v, err := strconv.ParseUint(rawValue, 10, 64)
		if err != nil {
			return err
		}
		p.Value = v
	case ParamTypeBytes:
		v, err := hex.DecodeString(rawValue)
		if err != nil {
			return err
		}
		p.Value = v
	case ParamTypeStrings:
		parts := strings.Split(rawValue, ",")
		p.Value = parts
	default:
		return errors.New("unsupported param type: " + p.Type + " (" + entry + ")")
	}

	return p.ValidateBasic()
}

func (p Param) String() string {
	typedKey := p.Key + "." + p.Type
	switch p.Type {
	case ParamTypeString:
		return fmt.Sprintf("%s=%s", typedKey, p.Value)
	case ParamTypeInt64:
		return fmt.Sprintf("%s=%d", typedKey, p.Value)
	case ParamTypeUint64:
		return fmt.Sprintf("%s=%d", typedKey, p.Value)
	case ParamTypeBool:
		if p.Value.(bool) {
			return fmt.Sprintf("%s=true", typedKey)
		}
		return fmt.Sprintf("%s=false", typedKey)
	case ParamTypeBytes:
		return fmt.Sprintf("%s=%x", typedKey, p.Value)
	case ParamTypeStrings:
		valstr := strings.Join(p.Value.([]string), ",")
		return fmt.Sprintf("%s=%s", typedKey, valstr)
	}
	panic("invalid param type:" + p.Type)
}

func (p *Param) UnmarshalAmino(rep string) error {
	return p.Parse(rep)
}

func (p Param) MarshalAmino() (string, error) {
	return p.String(), nil
}
