package simple_test

import (
	"testing"

	"gno.land/p/demo/gnorkle/storage"
	"gno.land/p/demo/gnorkle/storage/simple"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestStorage(t *testing.T) {
	var undefinedStorage *simple.Storage
	err := undefinedStorage.Put("")
	uassert.ErrorIs(t, err, storage.ErrUndefined, "expected storage.ErrUndefined on undefined storage")

	tests := []struct {
		name                      string
		valuesToPut               []string
		expLatestValueString      string
		expLatestValueTimeIsZero  bool
		expHistoricalValueStrings []string
	}{
		{
			name:                     "empty",
			expLatestValueTimeIsZero: true,
		},
		{
			name:                      "one value",
			valuesToPut:               []string{"one"},
			expLatestValueString:      "one",
			expHistoricalValueStrings: []string{"one"},
		},
		{
			name:                      "two values",
			valuesToPut:               []string{"one", "two"},
			expLatestValueString:      "two",
			expHistoricalValueStrings: []string{"one", "two"},
		},
		{
			name:                      "three values",
			valuesToPut:               []string{"one", "two", "three"},
			expLatestValueString:      "three",
			expHistoricalValueStrings: []string{"two", "three"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simpleStorage := simple.NewStorage(2)
			for _, value := range tt.valuesToPut {
				err := simpleStorage.Put(value)
				urequire.NoError(t, err, "unexpected error putting value in storage")
			}

			latestValue := simpleStorage.GetLatest()
			uassert.Equal(t, tt.expLatestValueString, latestValue.String)
			uassert.Equal(t, tt.expLatestValueTimeIsZero, latestValue.Time.IsZero())

			historicalValues := simpleStorage.GetHistory()
			urequire.Equal(t, len(tt.expHistoricalValueStrings), len(historicalValues), "historical values length does not match")

			for i, expValue := range tt.expHistoricalValueStrings {
				uassert.Equal(t, historicalValues[i].String, expValue)
				urequire.False(t, historicalValues[i].Time.IsZero(), ufmt.Sprintf("unexpeced zero time for historical value at index %d", i))
			}
		})
	}
}
