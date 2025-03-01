package std

import (
	"fmt"

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

// NOTE: further validation must happen by implementor of ParamsInterface.
func pkey(m *gno.Machine, key string) string {
	if len(key) == 0 {
		m.Panic(typedString("empty param key"))
	}
	_, rlmPath := currentRealm(m)
	return fmt.Sprintf("vm:%s:%s", rlmPath, key)
}
