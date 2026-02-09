package health

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// HealthHandler fetches the node health.
// Returns empty result (200 OK) on success, no response - in case of an error
//
//	No params
func HealthHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	if len(p) > 0 {
		return nil, spec.GenerateInvalidParamError(1)
	}

	_, span := traces.Tracer().Start(context.Background(), "Health")
	defer span.End()

	return &ResultHealth{}, nil
}
