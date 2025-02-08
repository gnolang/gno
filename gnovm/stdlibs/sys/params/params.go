package params

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

func X_setPrefixedString(m *gno.Machine, keeperPrefix, key, val string) {
	pk, err := std.NewParamKey(m, keeperPrefix, key, "string")
	if err != nil {
		return
	}
	std.GetContext(m).Params.SetString(pk, val)
}

func X_setPrefixedBool(m *gno.Machine, keeperPrefix, key string, val bool) {
	pk, err := std.NewParamKey(m, keeperPrefix, key, "bool")
	if err != nil {
		return
	}
	std.GetContext(m).Params.SetBool(pk, val)
}

func X_setPrefixedInt64(m *gno.Machine, keeperPrefix, key string, val int64) {
	pk, err := std.NewParamKey(m, keeperPrefix, key, "int64")
	if err != nil {
		return
	}
	std.GetContext(m).Params.SetInt64(pk, val)
}

func X_setPrefixedUint64(m *gno.Machine, keeperPrefix, key string, val uint64) {
	pk, err := std.NewParamKey(m, keeperPrefix, key, "uint64")
	if err != nil {
		return
	}
	std.GetContext(m).Params.SetUint64(pk, val)
}

func X_setPrefixedBytes(m *gno.Machine, keeperPrefix, key string, val []byte) {
	pk, err := std.NewParamKey(m, keeperPrefix, key, "bytes")
	if err != nil {
		return
	}
	std.GetContext(m).Params.SetBytes(pk, val)
}
