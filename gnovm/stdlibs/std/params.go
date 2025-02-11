package std

import (
	"fmt"
	"strings"
	"unicode"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// ParamsInterface is the interface through which Gno is capable of accessing
// the blockchain's params.
//
// The name is what it is to avoid a collision with Gno's Params, when
// transpiling.
type ParamsInterface interface {
	SetString(ctx sdk.Context, key, val string)
	SetBool(ctx sdk.Context, key string, val bool)
	SetInt64(ctx sdk.Context, key string, val int64)
	SetUint64(ctx sdk.Context, key string, val uint64)
	SetBytes(ctx sdk.Context, key string, val []byte)
}

func X_setParamString(m *gno.Machine, key, val string) {
	pk := pkey(m, key, "string")
	msgCtx := GetContext(m)
	sdkCtx := msgCtx.SDKContext
	msgCtx.Params.SetString(sdkCtx, pk, val)
}

func X_setParamBool(m *gno.Machine, key string, val bool) {
	pk := pkey(m, key, "bool")
	msgCtx := GetContext(m)
	sdkCtx := msgCtx.SDKContext
	msgCtx.Params.SetBool(sdkCtx, pk, val)
}

func X_setParamInt64(m *gno.Machine, key string, val int64) {
	pk := pkey(m, key, "int64")
	msgCtx := GetContext(m)
	sdkCtx := msgCtx.SDKContext
	msgCtx.Params.SetInt64(sdkCtx, pk, val)
}

func X_setParamUint64(m *gno.Machine, key string, val uint64) {
	pk := pkey(m, key, "uint64")
	msgCtx := GetContext(m)
	sdkCtx := msgCtx.SDKContext
	msgCtx.Params.SetUint64(sdkCtx, pk, val)
}

func X_setParamBytes(m *gno.Machine, key string, val []byte) {
	pk := pkey(m, key, "bytes")
	msgCtx := GetContext(m)
	sdkCtx := msgCtx.SDKContext
	msgCtx.Params.SetBytes(sdkCtx, pk, val)
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
