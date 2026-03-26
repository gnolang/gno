package gnoland

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// nodeParamsKeeper implements a minimal ParamfulKeeper for the "node" module.
// It validates node-level parameters set through governance proposals.
type nodeParamsKeeper struct{}

// WillSetParam validates node parameters before they are written to the params store.
func (nodeParamsKeeper) WillSetParam(_ sdk.Context, key string, value any) {
	switch key {
	case "p:halt_height":
		h, ok := value.(int64)
		if !ok {
			panic(fmt.Sprintf("halt_height must be an int64, got %T", value))
		}
		if h < 0 {
			panic(fmt.Sprintf("halt_height must be non-negative, got %d", h))
		}
	default:
		if strings.HasPrefix(key, "p:") {
			panic(fmt.Sprintf("unknown node param key: %q", key))
		}
	}
}
