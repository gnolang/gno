package std

import (
	"fmt"
	"strings"
	"unicode"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// ParamsInterface is the interface through which Gno is capable of accessing
// the blockchain's params.
//
// The name is what it is to avoid a collision with Gno's Params, when
// transpiling.
type ParamsInterface interface {
	SetString(key, val string)
	SetBool(key string, val bool)
	SetInt64(key string, val int64)
	SetUint64(key string, val uint64)
	SetBytes(key string, val []byte)
}

func X_setParamString(m *gno.Machine, key, val string) {
	pk := pkey(m, key, "string")
	GetContext(m).Params.SetString(pk, val)
}

func X_setParamBool(m *gno.Machine, key string, val bool) {
	pk := pkey(m, key, "bool")
	GetContext(m).Params.SetBool(pk, val)
}

func X_setParamInt64(m *gno.Machine, key string, val int64) {
	pk := pkey(m, key, "int64")
	GetContext(m).Params.SetInt64(pk, val)
}

func X_setParamUint64(m *gno.Machine, key string, val uint64) {
	pk := pkey(m, key, "uint64")
	GetContext(m).Params.SetUint64(pk, val)
}

func X_setParamBytes(m *gno.Machine, key string, val []byte) {
	pk := pkey(m, key, "bytes")
	GetContext(m).Params.SetBytes(pk, val)
}

func pkey(m *gno.Machine, key string, kind string) string {
	// validate key.
	untypedKey := strings.TrimSuffix(key, "."+kind)
	if key == untypedKey {
		m.Panic(typedString("invalid param key: " + key))
	}

	if len(key) == 0 {
		m.Panic(typedString("empty param key"))
	}
	first := rune(key[0])
	if !unicode.IsLetter(first) && first != '_' {
		m.Panic(typedString("invalid param key: " + key))
	}
	for _, char := range untypedKey[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			m.Panic(typedString("invalid param key: " + key))
		}
	}

	// decorate key with realm and type.
	_, rlmPath := currentRealm(m)
	return fmt.Sprintf("%s.%s", rlmPath, key)
}
