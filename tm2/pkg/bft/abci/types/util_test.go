package abci

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdatesFrom(t *testing.T) {
	t.Parallel()

	newVU := func(addr string, power int64) ValidatorUpdate {
		return ValidatorUpdate{
			Address: testutils.TestAddress(t, addr),
			PubKey:  nil,
			Power:   power,
		}
	}

	tests := []struct {
		name            string
		prev, proposed  ValidatorUpdates
		expectedUpdates ValidatorUpdates
	}{
		{
			name:            "no changes",
			prev:            ValidatorUpdates{newVU("D", 8)},
			proposed:        ValidatorUpdates{newVU("D", 8)},
			expectedUpdates: nil,
		},
		{
			name:            "removal",
			prev:            ValidatorUpdates{newVU("A", 10)},
			proposed:        nil,
			expectedUpdates: ValidatorUpdates{newVU("A", 0)},
		},
		{
			name:            "addition",
			prev:            nil,
			proposed:        ValidatorUpdates{newVU("B", 20)},
			expectedUpdates: ValidatorUpdates{newVU("B", 20)},
		},
		{
			name:            "power change",
			prev:            ValidatorUpdates{newVU("C", 5)},
			proposed:        ValidatorUpdates{newVU("C", 7)},
			expectedUpdates: ValidatorUpdates{newVU("C", 7)},
		},
		{
			name: "mixed",
			prev: ValidatorUpdates{
				newVU("A", 1),
				newVU("B", 2),
				newVU("C", 3),
			},
			proposed: ValidatorUpdates{
				newVU("B", 20), // modified
				newVU("D", 4),  // new
			},
			expectedUpdates: ValidatorUpdates{
				newVU("A", 0),  // removed
				newVU("B", 20), // changed
				newVU("C", 0),  // removed
				newVU("D", 4),  // added
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			updates := testCase.prev.UpdatesFrom(testCase.proposed)

			// Make sure the contents match
			require.ElementsMatch(t, testCase.expectedUpdates, updates)

			// Make sure the lengths match
			assert.Len(t, updates, len(testCase.expectedUpdates))
		})
	}
}
