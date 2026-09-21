package abci

import (
	"bytes"
	"sort"
	"strconv"
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

	pub := ed25519.GenPrivKey().PubKey().String()

	tests := []struct {
		name      string
		entry     string
		wantPower int64
		wantErr   string // substring match; empty = no error
	}{
		{name: "valid update", entry: pub + ":7", wantPower: 7},
		{name: "valid removal", entry: pub + ":0", wantPower: 0},
		{name: "missing separator", entry: pub, wantErr: `"<pubkey>:<power>"`},
		{name: "bad pubkey", entry: "notapubkey:1", wantErr: "invalid validator pubkey"},
		{name: "negative power", entry: pub + ":-1", wantErr: "invalid voting power"},
		{name: "non-numeric power", entry: pub + ":abc", wantErr: "invalid voting power"},
		// math.MaxInt64 + 1; would overflow int64 if not capped.
		{name: "power overflowing int64", entry: pub + ":9223372036854775808", wantErr: "invalid voting power"},
		// C2 regression: bech32 payloads that amino-decode to no concrete
		// PubKey (empty/zero-byte payload) made PubKeyFromBech32 return
		// (nil, nil), which then nil-deref'd at pk.Address(). Must error,
		// not panic — a panic inside EndBlocker would halt the chain.
		{name: "C2 nil pubkey lowercase", entry: "gpub1mdgqmw:5", wantErr: "nil pubkey"},
		{name: "C2 nil pubkey uppercase", entry: "GPUB1MDGQMW:5", wantErr: "nil pubkey"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u, err := ParseValidatorUpdate(tc.entry)
			if tc.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantPower, u.Power)
				assert.Equal(t, u.PubKey.Address(), u.Address, "address must derive from pubkey")
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestParseValidatorUpdates(t *testing.T) {
	t.Parallel()

	pub1 := ed25519.GenPrivKey().PubKey().String()
	pub2 := ed25519.GenPrivKey().PubKey().String()

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		u, err := ParseValidatorUpdates(nil)
		require.NoError(t, err)
		assert.Empty(t, u)
	})

	t.Run("two valid entries", func(t *testing.T) {
		t.Parallel()
		u, err := ParseValidatorUpdates([]string{pub1 + ":1", pub2 + ":2"})
		require.NoError(t, err)
		require.Len(t, u, 2)
	})

	t.Run("error reports entry index", func(t *testing.T) {
		t.Parallel()
		_, err := ParseValidatorUpdates([]string{pub1 + ":1", "garbage"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entry 1:", "error must surface offending entry index")
	})
}

func TestEncodeValidatorUpdates(t *testing.T) {
	t.Parallel()

	pk1 := ed25519.GenPrivKey().PubKey()
	pk2 := ed25519.GenPrivKey().PubKey()
	pk3 := ed25519.GenPrivKey().PubKey()

	t.Run("empty input -> empty output", func(t *testing.T) {
		t.Parallel()
		got := EncodeValidatorUpdates(ValidatorUpdates{})
		require.NotNil(t, got)
		require.Empty(t, got)
	})

	t.Run("round-trip via ParseValidatorUpdates", func(t *testing.T) {
		t.Parallel()
		input := ValidatorUpdates{
			{Address: pk1.Address(), PubKey: pk1, Power: 5},
			{Address: pk2.Address(), PubKey: pk2, Power: 10},
			{Address: pk3.Address(), PubKey: pk3, Power: 3},
		}
		entries := EncodeValidatorUpdates(input)
		parsed, err := ParseValidatorUpdates(entries)
		require.NoError(t, err)
		require.Equal(t, len(input), len(parsed))
		// Equality modulo the in-place sort EncodeValidatorUpdates did.
		for i := range parsed {
			assert.Equal(t, input[i].PubKey.Address(), parsed[i].Address)
			assert.Equal(t, input[i].Power, parsed[i].Power)
		}
	})

	t.Run("output reflects pubkey-bytes Less ordering", func(t *testing.T) {
		t.Parallel()
		input := ValidatorUpdates{
			{Address: pk1.Address(), PubKey: pk1, Power: 1},
			{Address: pk2.Address(), PubKey: pk2, Power: 2},
			{Address: pk3.Address(), PubKey: pk3, Power: 3},
		}
		// Compute expected by sorting first, then encoding.
		sortedExpect := append(ValidatorUpdates(nil), input...)
		sort.Sort(sortedExpect)
		expectStrs := make([]string, len(sortedExpect))
		for i, u := range sortedExpect {
			expectStrs[i] = crypto.PubKeyToBech32(u.PubKey) + ":" + strconv.FormatInt(u.Power, 10)
		}
		got := EncodeValidatorUpdates(input)
		assert.Equal(t, expectStrs, got)
		require.True(t, sort.IsSorted(input), "input must be sorted in place post-call")
	})

	t.Run("deterministic across calls", func(t *testing.T) {
		t.Parallel()
		input := ValidatorUpdates{
			{Address: pk1.Address(), PubKey: pk1, Power: 5},
			{Address: pk2.Address(), PubKey: pk2, Power: 10},
		}
		a := EncodeValidatorUpdates(append(ValidatorUpdates(nil), input...))
		b := EncodeValidatorUpdates(append(ValidatorUpdates(nil), input...))
		require.Equal(t, a, b)
	})

	t.Run("nil PubKey panics", func(t *testing.T) {
		t.Parallel()
		// Use TWO entries with the second having nil PubKey. Sort by
		// Less is called on input first; if the non-nil one sorts
		// after the nil one, encoding would have already deref'd nil
		// during sort. To make the test deterministic, both have
		// nil PubKey == nil interface, and sort.Sort needs to handle
		// the comparison... actually nil.Bytes() panics in Less.
		// Skip this complication: test just one entry with nil PubKey.
		input := ValidatorUpdates{
			{Power: 1}, // PubKey nil; len=1 so sort is a no-op
		}
		assert.PanicsWithValue(t,
			"EncodeValidatorUpdates: nil PubKey at index 0 (genesis misconfig)",
			func() { _ = EncodeValidatorUpdates(input) },
		)
	})
}

// TestLessIsPubkeyBytes pins ValidatorUpdates.Less semantics with a
// hand-crafted golden input. Any silent change to Less (e.g. switching
// to address-bytes) would break this test, forcing a coordinated
// on-disk format version bump rather than a quiet wire-format drift.
func TestLessIsPubkeyBytes(t *testing.T) {
	t.Parallel()

	// Deterministic seeds so pubkey bytes are stable across runs.
	pkA := ed25519.GenPrivKeyFromSecret([]byte("seed-A")).PubKey()
	pkB := ed25519.GenPrivKeyFromSecret([]byte("seed-B")).PubKey()
	pkC := ed25519.GenPrivKeyFromSecret([]byte("seed-C")).PubKey()

	// Manually compute expected order by raw pubkey-bytes.
	all := []crypto.PubKey{pkA, pkB, pkC}
	sort.Slice(all, func(i, j int) bool {
		return bytes.Compare(all[i].Bytes(), all[j].Bytes()) < 0
	})
	expected := []ValidatorUpdate{
		{Address: all[0].Address(), PubKey: all[0], Power: 1},
		{Address: all[1].Address(), PubKey: all[1], Power: 2},
		{Address: all[2].Address(), PubKey: all[2], Power: 3},
	}

	// Feed in deliberately-shuffled order; assert sort.Sort produces
	// the pubkey-bytes-ordered result.
	input := ValidatorUpdates{
		{Address: pkC.Address(), PubKey: pkC, Power: 3},
		{Address: pkA.Address(), PubKey: pkA, Power: 1},
		{Address: pkB.Address(), PubKey: pkB, Power: 2},
	}
	sort.Sort(input)
	for i, want := range expected {
		assert.Equal(t, want.PubKey.Address(), input[i].Address,
			"index %d: Less must order by pubkey-bytes (was the sort key changed?)", i)
	}
}
