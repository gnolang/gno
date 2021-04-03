package types

import (
	sdk "github.com/tendermint/classic/sdk/types"
)

// QuerySigningInfoParams defines the params for the following queries:
// - 'custom/slashing/signingInfo'
type QuerySigningInfoParams struct {
	ConsAddress sdk.ConsAddress
}

func NewQuerySigningInfoParams(consAddr sdk.ConsAddress) QuerySigningInfoParams {
	return QuerySigningInfoParams{consAddr}
}

// QuerySigningInfosParams defines the params for the following queries:
// - 'custom/slashing/signingInfos'
type QuerySigningInfosParams struct {
	Page, Limit int
}

func NewQuerySigningInfosParams(page, limit int) QuerySigningInfosParams {
	return QuerySigningInfosParams{page, limit}
}
