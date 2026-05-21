package params

import (
	"fmt"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
)

// std.SetParam*() can only be used to set realm-local VM parameters.  All
// parameters stored in ExecContext.Params will be prefixed by "vm:<realm>:".
// TODO rename to SetRealmParam*().

func SetString(m *gno.Machine, key, val string) {
	pk := pkey(m, key)
	execctx.GetContext(m).Params.SetString(pk, val)
}

func SetBool(m *gno.Machine, key string, val bool) {
	pk := pkey(m, key)
	execctx.GetContext(m).Params.SetBool(pk, val)
}

func SetInt64(m *gno.Machine, key string, val int64) {
	pk := pkey(m, key)
	execctx.GetContext(m).Params.SetInt64(pk, val)
}

func SetUint64(m *gno.Machine, key string, val uint64) {
	pk := pkey(m, key)
	execctx.GetContext(m).Params.SetUint64(pk, val)
}

func SetBytes(m *gno.Machine, key string, val []byte) {
	pk := pkey(m, key)
	execctx.GetContext(m).Params.SetBytes(pk, val)
}

func SetBytesKey(m *gno.Machine, key []byte, val []byte) {
	pk := pkeyRaw(m, key)
	execctx.GetContext(m).Params.SetBytes(pk, val)
}

func SetStrings(m *gno.Machine, key string, val []string) {
	pk := pkey(m, key)
	execctx.GetContext(m).Params.SetStrings(pk, val)
}

func UpdateParamStrings(m *gno.Machine, key string, val []string, add bool) {
	pk := pkey(m, key)
	execctx.GetContext(m).Params.UpdateStrings(pk, val, add)
}

func GetBytesKey(m *gno.Machine, key []byte) ([]byte, bool) {
	pk := pkeyRaw(m, key)
	var out []byte
	ok := execctx.GetContext(m).Params.GetBytes(pk, &out)
	return out, ok
}

// NOTE: further validation must happen by implementor of ParamsInterface.
func pkey(m *gno.Machine, key string) string {
	if len(key) == 0 {
		m.PanicString("empty param key")
	}
	if strings.Contains(key, ":") {
		m.PanicString("invalid param key: " + key)
	}
	_, rlmPath := execctx.CurrentRealm(m)
	return fmt.Sprintf("vm:%s:%s", rlmPath, key)
}

// pkeyRaw builds the realm-prefixed storage key for a raw binary key.
// Unlike pkey it does not reject ':' in the key: the key is appended verbatim.
func pkeyRaw(m *gno.Machine, key []byte) string {
	if len(key) == 0 {
		m.PanicString("empty param key")
	}
	_, rlmPath := execctx.CurrentRealm(m)
	return "vm:" + rlmPath + ":" + string(key)
}
