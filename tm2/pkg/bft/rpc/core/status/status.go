package status

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/params"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

type BuildStatusFn func() (*ResultStatus, error)

// Handler is the status RPC handler
type Handler struct {
	buildFn BuildStatusFn
}

// NewHandler creates a new instance of the status RPC handler
func NewHandler(buildFn BuildStatusFn) *Handler {
	return &Handler{
		buildFn: buildFn,
	}
}

// StatusHandler fetches the Tendermint status, including node info, pubkey, latest block
// hash, app hash, block height and time.
//
//	Params:
//	- heightGte (optional, defaults to 0)
func (h *Handler) StatusHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	_, span := traces.Tracer().Start(context.Background(), "Status")
	defer span.End()

	const idxHeightGte = 0

	heightGte, err := params.AsInt64(p, idxHeightGte)
	if err != nil {
		return nil, err
	}

	res, buildErr := h.buildFn()
	if buildErr != nil {
		return nil, spec.GenerateResponseError(buildErr)
	}

	latestHeight := res.SyncInfo.LatestBlockHeight

	if heightGte > 0 && latestHeight < heightGte {
		return nil, spec.NewJSONError(
			fmt.Sprintf(
				"latest height is %d, which is less than %d",
				latestHeight,
				heightGte,
			),
			spec.InvalidRequestErrorCode,
		)
	}

	return res, nil
}
