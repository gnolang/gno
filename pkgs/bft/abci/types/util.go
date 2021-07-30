package abci

import (
	"bytes"
	"sort"
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
