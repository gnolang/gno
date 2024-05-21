package simple_test

import (
	"testing"

	"gno.land/p/demo/gnorkle/storage"
	"gno.land/p/demo/gnorkle/storage/simple"
)

func TestStorage(t *testing.T) {
	var undefinedStorage *simple.Storage
	if err := undefinedStorage.Put(""); err != storage.ErrUndefined {
		t.Error("expected storage.ErrUndefined on undefined storage")
	}

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
				if err := simpleStorage.Put(value); err != nil {
					t.Fatal("unexpected error putting value in storage")
				}
			}

			latestValue := simpleStorage.GetLatest()
			if latestValue.String != tt.expLatestValueString {
				t.Errorf("expected latest value of %s, got %s", tt.expLatestValueString, latestValue.String)
			}

			if latestValue.Time.IsZero() != tt.expLatestValueTimeIsZero {
				t.Errorf(
					"expected latest value time zero value of %t, got %t",
					tt.expLatestValueTimeIsZero,
					latestValue.Time.IsZero(),
				)
			}

			historicalValues := simpleStorage.GetHistory()
			if len(historicalValues) != len(tt.expHistoricalValueStrings) {
				t.Fatal("historical values length does not match")
			}

			for i, expValue := range tt.expHistoricalValueStrings {
				if expValue != historicalValues[i].String {
					t.Errorf(
						"expected historical value at idx %d to be %s, got %s",
						i,
						expValue,
						historicalValues[i].String,
					)
				}

				if historicalValues[i].Time.IsZero() {
					t.Errorf("unexpeced zero time for historical value at index %d", i)
				}
			}
		})
	}
}
