package std

import (
	"fmt"
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
	// XXX: GetString(key string) (string, error)?
}

func X_setConfigString(m *gno.Machine, key, val string) {
	pk := pkey(m, key, "string")
	GetContext(m).Params.SetString(pk, val)
}

func X_setConfigBool(m *gno.Machine, key string, val bool) {
	pk := pkey(m, key, "bool")
	GetContext(m).Params.SetBool(pk, val)
}

func X_setConfigInt64(m *gno.Machine, key string, val int64) {
	pk := pkey(m, key, "int64")
	GetContext(m).Params.SetInt64(pk, val)
}

func X_setConfigUint64(m *gno.Machine, key string, val uint64) {
	pk := pkey(m, key, "uint64")
	GetContext(m).Params.SetUint64(pk, val)
}

func pkey(m *gno.Machine, key string, kind string) string {
	// validate key.
	if len(key) == 0 {
		panic("empty param key")
	}
	first := rune(key[0])
	if !unicode.IsLetter(first) && first != '_' {
		panic("invalid param key: " + key)
	}
	for _, char := range key[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			panic("invalid param key: " + key)
		}
	}

	// decorate key with realm and type.
	_, rlmPath := currentRealm(m)
	return fmt.Sprintf("%s.%s.%s", rlmPath, key, kind)
}
