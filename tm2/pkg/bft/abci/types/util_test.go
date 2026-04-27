package abci

import (
	"fmt"
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

func TestParseValidatorUpdate(t *testing.T) {
	t.Parallel()

	// Build a valid serialized entry for one identity.
	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()
	addr := pub.Address()
	good := fmt.Sprintf("%s:%s:%d", addr.String(), pub, 7)

	// Same identity but with Power=0 (removal).
	removal := fmt.Sprintf("%s:%s:%d", addr.String(), pub, 0)

	tests := []struct {
		name      string
		input     string
		wantPower int64
		wantErr   string // substring match; empty = no error
	}{
		{name: "valid update", input: good, wantPower: 7},
		{name: "valid removal", input: removal, wantPower: 0},
		{name: "wrong field count", input: "a:b", wantErr: "format <address>:<pub-key>:<voting-power>"},
		{name: "bad address", input: "notbech32:" + pub.String() + ":1", wantErr: "invalid validator address"},
		{name: "bad pubkey", input: addr.String() + ":notapubkey:1", wantErr: "invalid validator pubkey"},
		{name: "address/pubkey mismatch", input: ed25519.GenPrivKey().PubKey().Address().String() + ":" + pub.String() + ":1", wantErr: "does not match pubkey-derived"},
		{name: "negative power rejected", input: addr.String() + ":" + pub.String() + ":-1", wantErr: "invalid voting power"},
		{name: "non-numeric power", input: addr.String() + ":" + pub.String() + ":abc", wantErr: "invalid voting power"},
		// 9223372036854775808 == math.MaxInt64 + 1; would overflow int64 if not capped.
		{name: "power overflowing int64 rejected", input: addr.String() + ":" + pub.String() + ":9223372036854775808", wantErr: "invalid voting power"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u, err := ParseValidatorUpdate(tc.input)
			if tc.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantPower, u.Power)
				assert.Equal(t, addr, u.Address)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestParseValidatorUpdates(t *testing.T) {
	t.Parallel()

	priv1, priv2 := ed25519.GenPrivKey(), ed25519.GenPrivKey()
	mk := func(p crypto.PrivKey, power int64) string {
		pub := p.PubKey()
		return fmt.Sprintf("%s:%s:%d", pub.Address().String(), pub, power)
	}

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		u, err := ParseValidatorUpdates(nil)
		require.NoError(t, err)
		assert.Empty(t, u)
	})

	t.Run("two valid entries", func(t *testing.T) {
		t.Parallel()
		u, err := ParseValidatorUpdates([]string{mk(priv1, 1), mk(priv2, 2)})
		require.NoError(t, err)
		require.Len(t, u, 2)
	})

	t.Run("error reports entry index", func(t *testing.T) {
		t.Parallel()
		_, err := ParseValidatorUpdates([]string{mk(priv1, 1), "garbage"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entry 1:", "error must surface offending entry index")
	})
}
