package abci

import (
	"bytes"
	"sort"

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

//----------------------------------------
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
