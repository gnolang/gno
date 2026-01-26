package core_types

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// TODO move to respective packages

// Info abci msg
type ResultABCIInfo struct {
	Response abci.ResponseInfo `json:"response"`
}

// Query abci msg
type ResultABCIQuery struct {
	Response abci.ResponseQuery `json:"response"`
}

// empty results
type (
	ResultUnsafeFlushMempool struct{}
	ResultUnsafeProfile      struct{}
	ResultHealth             struct{}
)

// Event data from a subscription
type ResultEvent struct {
	Event types.TMEvent `json:"event"`
}
