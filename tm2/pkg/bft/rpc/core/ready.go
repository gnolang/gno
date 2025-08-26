package core

import (
	"fmt"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

// Get node readyness. Returns 200 OK on success, 500 Internal Server Error - in
// case of an error.
//
// ```shell
// curl 'localhost:26657/ready'
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
// result, err := client.Ready()
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
func Ready(ctx *rpctypes.Context) (*ctypes.ResultReady, error) {
	var latestHeight int64
	if getFastSync() {
		latestHeight = blockStore.Height()
	} else {
		latestHeight = consensusState.GetLastHeight()
	}

	if latestHeight < 1 {
		return &ctypes.ResultReady{}, rpctypes.NewHTTPStatusError(503, fmt.Sprintf("not ready: latest height is %d", latestHeight))
	}

	return &ctypes.ResultReady{}, nil
}
