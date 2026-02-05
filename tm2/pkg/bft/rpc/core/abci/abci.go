package abci

import (
	"context"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/params"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Handler is the ABCI RPC handler
type Handler struct {
	proxyAppQuery appconn.Query
}

// NewHandler creates a new instance of the ABCI RPC handler
func NewHandler(proxyAppQuery appconn.Query) *Handler {
	return &Handler{
		proxyAppQuery: proxyAppQuery,
	}
}

// QueryHandler queries the application (synchronously) for some information
//
//		Params:
//	  - path   string (optional, default "")
//	  - data   []byte (required)
//	  - height int64  (optional, default 0)
//	  - prove  bool   (optional, default false)
func (h *Handler) QueryHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	const (
		idxPath   = 0
		idxData   = 1
		idxHeight = 2
		idxProve  = 3
	)

	path, err := params.AsString(p, idxPath)
	if err != nil {
		return nil, err
	}

	data, err := params.AsBytes(p, idxData, true)
	if err != nil {
		return nil, err
	}

	height, err := params.AsInt64(p, idxHeight)
	if err != nil {
		return nil, err
	}

	prove, err := params.AsBool(p, idxProve)
	if err != nil {
		return nil, err
	}

	resQuery, queryErr := h.proxyAppQuery.QuerySync(abci.RequestQuery{
		Path:   path,
		Data:   data,
		Height: height,
		Prove:  prove,
	})
	if queryErr != nil {
		return nil, spec.GenerateResponseError(queryErr)
	}

	return &ResultABCIQuery{
		Response: resQuery,
	}, nil
}

// InfoHandler gets some info about the application.
//
//	No params
func (h *Handler) InfoHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	// Make sure there are no params
	if len(p) > 0 {
		return nil, spec.GenerateInvalidParamError(1)
	}

	_, span := traces.Tracer().Start(context.Background(), "ABCIInfo")
	defer span.End()

	resInfo, err := h.proxyAppQuery.InfoSync(abci.RequestInfo{})
	if err != nil {
		return nil, spec.GenerateResponseError(err)
	}

	return &ResultABCIInfo{
		Response: resInfo,
	}, nil
}
