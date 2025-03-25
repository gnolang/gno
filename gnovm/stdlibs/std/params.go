package std

import (
	"fmt"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// std.SetParam*() can only be used to set realm-local VM parameters.  All
// parameters stored in ExecContext.Params will be prefixed by "vm:<realm>:".
// TODO rename to SetRealmParam*().

type ParamsInterface interface {
	SetString(key, val string)
	SetBool(key string, val bool)
	SetInt64(key string, val int64)
	SetUint64(key string, val uint64)
	SetBytes(key string, val []byte)
	SetStrings(key string, val []string)

	GetString(key string) (string, bool)
	GetBool(key string) (bool, bool)
	GetInt64(key string) (int64, bool)
	GetUint64(key string) (uint64, bool)
	GetBytes(key string) ([]byte, bool)
	GetStrings(key string) ([]string, bool)
}

func X_setParamString(m *gno.Machine, key, val string) {
	pk := pkey(m, key)
	GetContext(m).Params.SetString(pk, val)
}

func X_setParamBool(m *gno.Machine, key string, val bool) {
	pk := pkey(m, key)
	GetContext(m).Params.SetBool(pk, val)
}

func X_setParamInt64(m *gno.Machine, key string, val int64) {
	pk := pkey(m, key)
	GetContext(m).Params.SetInt64(pk, val)
}

func X_setParamUint64(m *gno.Machine, key string, val uint64) {
	pk := pkey(m, key)
	GetContext(m).Params.SetUint64(pk, val)
}

func X_setParamBytes(m *gno.Machine, key string, val []byte) {
	pk := pkey(m, key)
	GetContext(m).Params.SetBytes(pk, val)
}

func X_setParamStrings(m *gno.Machine, key string, val []string) {
	pk := pkey(m, key)
	GetContext(m).Params.SetStrings(pk, val)
}

func X_getParamString(m *gno.Machine, key string) (string, bool) {
	// TODO @moul, note that we allow any key to be used for fetching
	// in the params keeper. This means that realms can query the
	// params of other realms (but not modify them).
	// If we want to preserve a params "scope" to the current realm,
	// let me know, and we'll guard the queries
	return GetContext(m).Params.GetString(key)
}

func X_getParamBool(m *gno.Machine, key string) (bool, bool) {
	return GetContext(m).Params.GetBool(key)
}

func X_getParamInt64(m *gno.Machine, key string) (int64, bool) {
	return GetContext(m).Params.GetInt64(key)
}

func X_getParamUint64(m *gno.Machine, key string) (uint64, bool) {
	return GetContext(m).Params.GetUint64(key)
}

func X_getParamBytes(m *gno.Machine, key string) ([]byte, bool) {
	return GetContext(m).Params.GetBytes(key)
}

func X_getParamStrings(m *gno.Machine, key string) ([]string, bool) {
	return GetContext(m).Params.GetStrings(key)
}

// NOTE: further validation must happen by implementor of ParamsInterface.
func pkey(m *gno.Machine, key string) string {
	if len(key) == 0 {
		m.Panic(typedString("empty param key"))
	}
	if strings.Contains(key, ":") {
		m.Panic(typedString("invalid param key: " + key))
	}
	_, rlmPath := currentRealm(m)
	return fmt.Sprintf("vm:%s:%s", rlmPath, key)
}
