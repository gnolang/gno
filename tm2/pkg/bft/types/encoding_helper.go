package types

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

// bytesOrNil returns nil if the input is nil, otherwise returns
// amino.MustMarshal(item)
func bytesOrNil(item interface{}) []byte {
	if item != nil && !amino.IsTypedNil(item) && !amino.IsEmpty(item) {
		return amino.MustMarshal(item)
	}
	return nil
}
