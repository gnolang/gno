package core

import (
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Health returns node health. Returns empty result (200 OK) on success, no
// response in case of an error.
func (env *Environment) Health(ctx *rpctypes.Context) (*ctypes.ResultHealth, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "Health")
	defer span.End()
	return &ctypes.ResultHealth{}, nil
}
