package core

import (
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Get node health. Returns empty result (200 OK) on success, no response - in
// case of an error.
//
// ```shell
// curl 'localhost:26657/health'
// ```
//
// ```go
// client := client.NewHTTP("tcp://0.0.0.0:26657", "/websocket")
// err := client.Start()
//
//	if err != nil {
//	  // handle error
//	}
//
// defer client.Stop()
// result, err := client.Health()
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
//
//	{
//		"error": "",
//		"result": {},
//		"id": "",
//		"jsonrpc": "2.0"
//	}
//
// ```
func Health(ctx *rpctypes.Context) (*ctypes.ResultHealth, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "Health")
	defer span.End()
	return &ctypes.ResultHealth{}, nil
}
