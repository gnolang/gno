package params

import (
	"fmt"
	"strings"
	"unicode"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

func X_setPrefixedString(m *gno.Machine, keeperPrefix, key, val string) {
	pk := pkey(m, keeperPrefix, key, "string")
	std.GetContext(m).Params.SetString(pk, val)
}

func X_setPrefixedBool(m *gno.Machine, keeperPrefix, key string, val bool) {
	pk := pkey(m, keeperPrefix, key, "bool")
	std.GetContext(m).Params.SetBool(pk, val)
}

func X_setPrefixedInt64(m *gno.Machine, keeperPrefix, key string, val int64) {
	pk := pkey(m, keeperPrefix, key, "int64")
	std.GetContext(m).Params.SetInt64(pk, val)
}

func X_setPrefixedUint64(m *gno.Machine, keeperPrefix, key string, val uint64) {
	pk := pkey(m, keeperPrefix, key, "uint64")
	std.GetContext(m).Params.SetUint64(pk, val)
}

func X_setPrefixedBytes(m *gno.Machine, keeperPrefix, key string, val []byte) {
	pk := pkey(m, keeperPrefix, key, "bytes")
	std.GetContext(m).Params.SetBytes(pk, val)
}

func pkey(m *gno.Machine, keeperPrefix, key string, kind string) string {
	// validate key.
	untypedKey := strings.TrimSuffix(key, "."+kind)
	if key == untypedKey {
		m.Panic(std.TypedString("invalid param key: " + key))
	}

	if len(key) == 0 {
		m.Panic(std.TypedString("empty param key"))
	}
	first := rune(key[0])
	if !unicode.IsLetter(first) && first != '_' {
		m.Panic(std.TypedString("invalid param key: " + key))
	}
	for _, char := range untypedKey[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			m.Panic(std.TypedString("invalid param key: " + key))
		}
	}

	first = rune(keeperPrefix[0])
	if !unicode.IsLetter(first) && first != '_' {
		m.Panic(std.TypedString("invalid module prefix: " + keeperPrefix))
	}
	for _, char := range keeperPrefix[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			m.Panic(std.TypedString("invalid module prefix: " + keeperPrefix))
		}
	}

	_, rlmPath := std.CurrentRealm(m)
	// decorate key with realm and type.
	return fmt.Sprintf("%s.%s:%s", rlmPath, keeperPrefix, key)
}
