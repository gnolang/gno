package core

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// ABCIQuery queries the application for some information.
func (env *Environment) ABCIQuery(ctx *rpctypes.Context, path string, data []byte, height int64, prove bool) (*ctypes.ResultABCIQuery, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "ABCIQuery")
	defer span.End()
	resQuery, err := env.ProxyAppQuery.QuerySync(abci.RequestQuery{
		Path:   path,
		Data:   data,
		Height: height,
		Prove:  prove,
	})
	if err != nil {
		return nil, err
	}
	env.Logger.Debug("ABCIQuery", "path", path, "data", data, "result", resQuery)
	return &ctypes.ResultABCIQuery{Response: resQuery}, nil
}

// ABCIInfo gets some info about the application.
func (env *Environment) ABCIInfo(ctx *rpctypes.Context) (*ctypes.ResultABCIInfo, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "ABCIInfo")
	defer span.End()
	resInfo, err := env.ProxyAppQuery.InfoSync(abci.RequestInfo{})
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultABCIInfo{Response: resInfo}, nil
}
