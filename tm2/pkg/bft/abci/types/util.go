package abci

import (
	"bytes"
	"sort"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

// ------------------------------------------------------------------------------

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

// ----------------------------------------
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
