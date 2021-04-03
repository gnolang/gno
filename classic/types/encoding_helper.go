package types

import (
	cmn "github.com/tendermint/classic/libs/common"
	"github.com/tendermint/go-amino-x"
)

// bytesOrNil returns nil if the input is nil, otherwise returns
// amino.MustMarshal(item)
func bytesOrNil(item interface{}) []byte {
	if item != nil && !cmn.IsTypedNil(item) && !cmn.IsEmpty(item) {
		return amino.MustMarshal(item)
	}
	return nil
}
