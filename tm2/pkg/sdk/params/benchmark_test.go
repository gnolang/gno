package params

import (
	"fmt"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

func BenchmarkParamsHandler_Query(b *testing.B) {
	var (
		prefix = "benchmark"
		key    = "random/key.string"
		value  = "keeper value"

		// Prepare the keeper
		env = setupTestEnv(b, prefix)

		// Prepare the handler
		handler = NewHandler(env.keeper)

		query = abci.RequestQuery{
			Path: fmt.Sprintf("params/%s/%s", prefix, key),
		}
	)

	// Set an initial value in the keeper
	env.keeper.SetString(env.ctx, key, value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Query(env.ctx, query)
	}
}
