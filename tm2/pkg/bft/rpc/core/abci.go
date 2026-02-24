package core

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Query the application for some information.
//
// ```shell
// curl 'localhost:26657/abci_query?path=""&data="abcd"&prove=false'
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
// result, err := client.ABCIQuery("", "abcd", true)
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
//
//	{
//		"error": "",
//		"result": {
//			"response": {
//				"log": "exists",
//				"height": "0",
//				"proof": "010114FED0DAD959F36091AD761C922ABA3CBF1D8349990101020103011406AA2262E2F448242DF2C2607C3CDC705313EE3B0001149D16177BC71E445476174622EA559715C293740C",
//				"value": "61626364",
//				"key": "61626364",
//				"index": "-1",
//				"code": "0"
//			}
//		},
//		"id": "",
//		"jsonrpc": "2.0"
//	}
//
// ```
//
// ### Query Parameters
//
// | Parameter | Type   | Default | Required | Description                                    |
// |-----------+--------+---------+----------+------------------------------------------------|
// | path      | string | false   | false    | Path to the data ("/a/b/c")                    |
// | data      | []byte | false   | true     | Data                                           |
// | height    | int64  | 0       | false    | Height (0 means latest)                        |
// | prove     | bool   | false   | false    | Includes proof if true                         |
func ABCIQuery(ctx *rpctypes.Context, path string, data []byte, height int64, prove bool) (*ctypes.ResultABCIQuery, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "ABCIQuery")
	defer span.End()
	resQuery, err := proxyAppQuery.QuerySync(abci.RequestQuery{
		Path:   path,
		Data:   data,
		Height: height,
		Prove:  prove,
	})
	if err != nil {
		return nil, err
	}
	logger.Info("ABCIQuery", "path", path, "data", data, "result", resQuery)
	return &ctypes.ResultABCIQuery{Response: resQuery}, nil
}

// Get some info about the application.
//
// ```shell
// curl 'localhost:26657/abci_info'
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
// info, err := client.ABCIInfo()
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
//
//	{
//		"error": "",
//		"result": {
//			"response": {
//				"data": "{\"size\":3}"
//			}
//		},
//		"id": "",
//		"jsonrpc": "2.0"
//	}
//
// ```
func ABCIInfo(ctx *rpctypes.Context) (*ctypes.ResultABCIInfo, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "ABCIInfo")
	defer span.End()
	resInfo, err := proxyAppQuery.InfoSync(abci.RequestInfo{})
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultABCIInfo{Response: resInfo}, nil
}
