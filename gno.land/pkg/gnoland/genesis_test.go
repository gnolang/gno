package gnoland

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBalances(t *testing.T) {
	b, err := LoadGenesisBalancesFile("")
	require.NoError(t, err)
	require.Len(t, b, 56)

	_, err = LoadGenesisBalancesFile("DOES_NOT_EXIST")
	require.Error(t, err)
}

func TestTransactions(t *testing.T) {
	b, err := LoadGenesisTxsFile("", "id", "remote")
	require.NoError(t, err)
	require.Len(t, b, 17)

	_, err = LoadGenesisTxsFile("DOES_NOT_EXIST", "id", "remote")
	require.Error(t, err)
}
