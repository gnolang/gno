package keeper

import (
	"fmt"

	abci "github.com/tendermint/classic/abci/types"
	"github.com/tendermint/go-amino-x"

	sdk "github.com/tendermint/classic/sdk/types"
	"github.com/tendermint/classic/sdk/x/mint/internal/types"
)

// NewQuerier returns a minting Querier handler.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, _ abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryParameters:
			return queryParams(ctx, k)

		case types.QueryInflation:
			return queryInflation(ctx, k)

		case types.QueryAnnualProvisions:
			return queryAnnualProvisions(ctx, k)

		default:
			return nil, sdk.ErrUnknownRequest(fmt.Sprintf("unknown minting query endpoint: %s", path[0]))
		}
	}
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)

	res, err := amino.MarshalJSONIndent(params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}

func queryInflation(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	minter := k.GetMinter(ctx)

	res, err := amino.MarshalJSONIndent(minter.Inflation)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}

func queryAnnualProvisions(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	minter := k.GetMinter(ctx)

	res, err := amino.MarshalJSONIndent(minter.AnnualProvisions)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}
