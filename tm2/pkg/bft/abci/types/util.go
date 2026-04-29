package abci

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

//------------------------------------------------------------------------------

// ValidatorUpdates is a list of validators that implements the Sort interface
type ValidatorUpdates []ValidatorUpdate

var _ sort.Interface = (ValidatorUpdates)(nil)

func (v ValidatorUpdates) Len() int {
	return len(v)
}

func (v ValidatorUpdates) Less(i, j int) bool {
	cmpAddr := bytes.Compare(v[i].PubKey.Bytes(), v[j].PubKey.Bytes())
	if cmpAddr == 0 {
		return v[i].Power < v[j].Power
	} else {
		return cmpAddr < 0
	}
}

func (v ValidatorUpdates) Swap(i, j int) {
	v1 := v[i]
	v[i] = v[j]
	v[j] = v1
}

// UpdatesFrom compares this ValidatorUpdates set with another (v2) and returns
// a new ValidatorUpdates containing only the changes needed to go from the
// receiver to v2. It includes:
//  1. Removals: validators present in the receiver but missing in v2 (Power = 0).
//  2. Power changes: validators present in both but whose Power differs.
//  3. Additions: validators present in v2 but missing in the receiver.
func (v ValidatorUpdates) UpdatesFrom(v2 ValidatorUpdates) ValidatorUpdates {
	prevMap := make(map[string]ValidatorUpdate, len(v))
	for _, val := range v {
		prevMap[val.Address.String()] = val
	}

	propMap := make(map[string]ValidatorUpdate, len(v2))
	for _, val := range v2 {
		propMap[val.Address.String()] = val
	}

	// Worst-case: all in v removed + all in v2 added
	diffs := make(ValidatorUpdates, 0, len(v)+len(v2))

	// Find all removals and updates
	for addr, prev := range prevMap {
		if prop, ok := propMap[addr]; ok {
			// If it exists in both -> check for power change
			if prop.Power != prev.Power {
				diffs = append(diffs, ValidatorUpdate{
					Address: prop.Address,
					PubKey:  prop.PubKey,
					Power:   prop.Power,
				})
			}

			continue
		}

		// If it's in prev but not in proposed -> removal
		diffs = append(diffs, ValidatorUpdate{
			Address: prev.Address,
			PubKey:  prev.PubKey,
			Power:   0,
		})
	}

	// Find additions (new validators)
	for addr, prop := range propMap {
		if _, seen := prevMap[addr]; !seen {
			diffs = append(diffs, prop)
		}
	}

	// Sort to give a deterministic order independent of map iteration.
	// ResultsHash already excludes ABCI updates, so this is not consensus-critical,
	// but it stabilizes ABCIResponses, EventValidatorSetUpdates, and RPC output.
	sort.Sort(diffs)

	return diffs
}

// ----------------------------------------
// ValidatorUpdate

func (vu ValidatorUpdate) Equals(vu2 ValidatorUpdate) bool {
	if vu.Address == vu2.Address &&
		vu.PubKey.Equals(vu2.PubKey) &&
		vu.Power == vu2.Power {
		return true
	} else {
		return false
	}
}

//----------------------------------------
// Validator update parsing

// ParseValidatorUpdate parses a single "<pubkey>:<power>" entry.
// Address is derived from the pubkey; powers are non-negative and capped
// at math.MaxInt64 so the int64 cast can't overflow.
func ParseValidatorUpdate(entry string) (ValidatorUpdate, error) {
	pubKey, power, ok := strings.Cut(entry, ":")
	if !ok {
		return ValidatorUpdate{}, fmt.Errorf(
			`valset entry %q is not in the form "<pubkey>:<power>"`,
			entry,
		)
	}
	pk, err := crypto.PubKeyFromBech32(pubKey)
	if err != nil {
		return ValidatorUpdate{}, fmt.Errorf("invalid validator pubkey %q: %w", pubKey, err)
	}
	// PubKeyFromBech32 returns (nil, nil) when the bech32 payload
	// amino-decodes to no concrete crypto.PubKey (e.g., empty payload
	// like "gpub1mdgqmw"). Without this guard, pk.Address() below
	// nil-derefs and panics — propagating to a chain halt if reached
	// from EndBlocker.
	if pk == nil {
		return ValidatorUpdate{}, fmt.Errorf("nil pubkey from bech32 %q (empty/non-decodable amino payload)", pubKey)
	}
	// bitSize=63 caps the value at math.MaxInt64, so int64(p) below
	// can never overflow into a negative.
	p, err := strconv.ParseUint(power, 10, 63)
	if err != nil {
		return ValidatorUpdate{}, fmt.Errorf("invalid voting power %q: %w", power, err)
	}
	return ValidatorUpdate{
		Address: pk.Address(),
		PubKey:  pk,
		Power:   int64(p),
	}, nil
}

// ParseValidatorUpdates parses a list of "<pubkey>:<power>" entries.
// Errors carry the offending entry index.
func ParseValidatorUpdates(entries []string) (ValidatorUpdates, error) {
	updates := make(ValidatorUpdates, 0, len(entries))
	for i, entry := range entries {
		u, err := ParseValidatorUpdate(entry)
		if err != nil {
			return nil, fmt.Errorf("entry %d: %w", i, err)
		}
		updates = append(updates, u)
	}
	return updates, nil
}

// EncodeValidatorUpdates is the inverse of ParseValidatorUpdates: it
// formats updates as "<bech32-pubkey>:<decimal-power>" strings, sorted
// canonically via ValidatorUpdates.Less (pubkey-bytes — see comment
// at Less above).
//
// Used by the chain's InitChainer (to seed valset:current from genesis
// validators) and EndBlocker (to write valset:current after applying a
// proposal). One canonical encoding eliminates drift between the two
// writers and lets round-trip tests live in one package.
//
// The input is sorted in place. Panics on a nil PubKey at any index —
// genesis is operator-supplied and a nil pubkey would otherwise nil-deref
// inside crypto.PubKeyToBech32; panic at InitChainer surfaces as an
// unrecoverable genesis-misconfig error, which is the correct semantic.
func EncodeValidatorUpdates(v ValidatorUpdates) []string {
	sort.Sort(v)
	out := make([]string, 0, len(v))
	for i, u := range v {
		if u.PubKey == nil {
			panic(fmt.Sprintf("EncodeValidatorUpdates: nil PubKey at index %d (genesis misconfig)", i))
		}
		out = append(out, crypto.PubKeyToBech32(u.PubKey)+":"+strconv.FormatInt(u.Power, 10))
	}
	return out
}

//----------------------------------------
// ABCIError helpers

func ABCIErrorOrStringError(err error) Error {
	if err == nil {
		return nil
	}
	err = errors.Cause(err) // unwrap
	abcierr, ok := err.(Error)
	if !ok {
		return StringError(err.Error())
	} else {
		return abcierr
	}
}
