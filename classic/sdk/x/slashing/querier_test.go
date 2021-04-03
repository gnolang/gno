package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/classic/abci/types"
	"github.com/tendermint/go-amino-x"
)

func TestNewQuerier(t *testing.T) {
	ctx, _, _, _, keeper := createTestInput(t, keeperTestParams())
	querier := NewQuerier(keeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{"parameters"}, query)
	require.NoError(t, err)
}

func TestQueryParams(t *testing.T) {
	ctx, _, _, _, keeper := createTestInput(t, keeperTestParams())

	var params Params

	res, errRes := queryParams(ctx, keeper)
	require.NoError(t, errRes)

	err := amino.UnmarshalJSON(res, &params)
	require.NoError(t, err)
	require.Equal(t, keeper.GetParams(ctx), params)
}
