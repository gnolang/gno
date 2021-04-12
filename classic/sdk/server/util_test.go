package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/go-amino-x"
)

func TestInsertKeyJSON(t *testing.T) {
	foo := map[string]string{"foo": "foofoo"}
	bar := map[string]string{"barInner": "barbar"}

	// create raw messages
	bz, err := amino.MarshalJSON(foo)
	require.NoError(t, err)
	fooRaw := json.RawMessage(bz)

	bz, err = amino.MarshalJSON(bar)
	require.NoError(t, err)
	barRaw := json.RawMessage(bz)

	// make the append
	appBz, err := InsertKeyJSON(fooRaw, "barOuter", barRaw)
	require.NoError(t, err)

	// test the append
	var appended map[string]json.RawMessage
	err = amino.UnmarshalJSON(appBz, &appended)
	require.NoError(t, err)

	var resBar map[string]string
	err = amino.UnmarshalJSON(appended["barOuter"], &resBar)
	require.NoError(t, err)

	require.Equal(t, bar, resBar, "appended: %v", appended)
}
