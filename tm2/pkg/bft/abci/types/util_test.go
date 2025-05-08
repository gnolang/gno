package abci

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdatesFrom(t *testing.T) {
	t.Parallel()

	newVU := func(key crypto.PubKey, power int64) ValidatorUpdate {
		return ValidatorUpdate{
			Address: key.Address(),
			PubKey:  key,
			Power:   power,
		}
	}

	generatePubKeys := func(count int) []crypto.PubKey {
		keys := make([]crypto.PubKey, 0, count)

		for range count {
			keys = append(keys, ed25519.GenPrivKey().PubKey())
		}

		return keys
	}

	validatorsKeys := generatePubKeys(4)

	tests := []struct {
		name            string
		prev, proposed  ValidatorUpdates
		expectedUpdates ValidatorUpdates
	}{
		{
			name:            "no changes",
			prev:            ValidatorUpdates{newVU(validatorsKeys[0], 8)},
			proposed:        ValidatorUpdates{newVU(validatorsKeys[0], 8)},
			expectedUpdates: nil,
		},
		{
			name:            "removal",
			prev:            ValidatorUpdates{newVU(validatorsKeys[0], 10)},
			proposed:        nil,
			expectedUpdates: ValidatorUpdates{newVU(validatorsKeys[0], 0)},
		},
		{
			name:            "addition",
			prev:            nil,
			proposed:        ValidatorUpdates{newVU(validatorsKeys[0], 20)},
			expectedUpdates: ValidatorUpdates{newVU(validatorsKeys[0], 20)},
		},
		{
			name:            "power change",
			prev:            ValidatorUpdates{newVU(validatorsKeys[0], 5)},
			proposed:        ValidatorUpdates{newVU(validatorsKeys[0], 7)},
			expectedUpdates: ValidatorUpdates{newVU(validatorsKeys[0], 7)},
		},
		{
			name: "mixed",
			prev: ValidatorUpdates{
				newVU(validatorsKeys[0], 1),
				newVU(validatorsKeys[1], 2),
				newVU(validatorsKeys[2], 3),
			},
			proposed: ValidatorUpdates{
				newVU(validatorsKeys[1], 20), // modified
				newVU(validatorsKeys[3], 4),  // new
			},
			expectedUpdates: ValidatorUpdates{
				newVU(validatorsKeys[0], 0),  // removed
				newVU(validatorsKeys[1], 20), // changed
				newVU(validatorsKeys[2], 0),  // removed
				newVU(validatorsKeys[3], 4),  // added
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
