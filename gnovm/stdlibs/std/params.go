package std

import (
	"fmt"
	"strings"
	"unicode"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// ParamsSetterInterface is the interface through which Gno is capable of accessing
// the blockchain's params.
//
// The name is what it is to avoid a collision with Gno's Params, when
// transpiling.

type ParamsInterface interface {
	SetString(key ParamKey, val string)
	SetBool(key ParamKey, val bool)
	SetInt64(key ParamKey, val int64)
	SetUint64(key ParamKey, val uint64)
	SetBytes(key ParamKey, val []byte)
}

type ParamKey struct {
	Realm  string
	Prefix string
	Key    string
	Type   string
}

func NewParamKey(m *gno.Machine, prefix, key string, kind string) (ParamKey, error) {
	// validate key.
	if err := validate(prefix, key, kind); err != nil {
		es := err.Error()
		m.Panic(typedString(es))
		return ParamKey{}, err
	}
	_, realm := currentRealm(m)

	return ParamKey{
		Realm:  realm,
		Prefix: prefix,
		Key:    key,
		Type:   kind,
	}, nil
}

// String representation of ParamKey in the format:
// <module>:(<realm>".")?<key>
func (pk ParamKey) String() string {
	pks := ""
	if pk.Prefix == "" {
		pks = fmt.Sprintf("vm:%s.%s", pk.Realm, pk.Key)
	} else {
		pks = fmt.Sprintf("%s:%s", pk.Prefix, pk.Key)
	}
	return pks
}

func X_setParamString(m *gno.Machine, key, val string) {
	pk, err := NewParamKey(m, "", key, "string")
	if err != nil {
		return
	}
	GetContext(m).Params.SetString(pk, val)
}

func X_setParamBool(m *gno.Machine, key string, val bool) {
	pk, err := NewParamKey(m, "", key, "bool")
	if err != nil {
		return
	}
	GetContext(m).Params.SetBool(pk, val)
}

func X_setParamInt64(m *gno.Machine, key string, val int64) {
	pk, err := NewParamKey(m, "", key, "int64")
	if err != nil {
		return
	}
	GetContext(m).Params.SetInt64(pk, val)
}

func X_setParamUint64(m *gno.Machine, key string, val uint64) {
	pk, err := NewParamKey(m, "", key, "uint64")
	if err != nil {
		return
	}
	GetContext(m).Params.SetUint64(pk, val)
}

func X_setParamBytes(m *gno.Machine, key string, val []byte) {
	pk, err := NewParamKey(m, "", key, "bytes")
	if err != nil {
		return
	}
	GetContext(m).Params.SetBytes(pk, val)
}

func validate(prefix, key, kind string) error {
	// validate key.
	untypedKey := strings.TrimSuffix(key, "."+kind)
	if key == untypedKey {
		return fmt.Errorf("invalid parameter key: %s", key)
	}

	if len(key) == 0 {
		return fmt.Errorf("empty param key")
	}
	first := rune(key[0])
	if !unicode.IsLetter(first) && first != '_' {
		return fmt.Errorf("invalid parameter key: %s", key)
	}
	for _, char := range untypedKey[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return fmt.Errorf("invalid parameter key: %s", key)
		}
	}

	if prefix != "" {
		// validate keeperPrefix
		first = rune(prefix[0])
		if !unicode.IsLetter(first) {
			return fmt.Errorf("invalid prefix: %s", prefix)
		}
		for _, char := range prefix[1:] {
			if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
				return fmt.Errorf("invalid prefix: %s", prefix)
			}
		}
	}
	return nil
}
