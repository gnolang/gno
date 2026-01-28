package abci

import abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"

type ResultABCIInfo struct {
	Response abci.ResponseInfo `json:"response"`
}

type ResultABCIQuery struct {
	Response abci.ResponseQuery `json:"response"`
}
