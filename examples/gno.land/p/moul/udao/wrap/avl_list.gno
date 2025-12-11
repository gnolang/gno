package wrap

import (
	"gno.land/p/demo/avl/list"
	"gno.land/p/moul/udao"
)

// PropListWrapper wraps an IReadOnlyList to implement PropList
type PropListWrapper struct {
	list list.IList
}

// WrapAsPropList converts an IReadOnlyList to a PropList
func WrapAsPropList(list list.IList) udao.PropList {
	return &PropListWrapper{list: list}
}

func (pl *PropListWrapper) Len() int {
	return pl.list.Len()
}

func (pl *PropListWrapper) Get(index int) udao.Proposal {
	return pl.list.Get(index).(udao.Proposal)
}

func (pl *PropListWrapper) Slice(startIndex, endIndex int) []udao.Proposal {
	slice := pl.list.Slice(startIndex, endIndex)
	result := make([]udao.Proposal, len(slice))
	for i, v := range slice {
		result[i] = v.(udao.Proposal)
	}
	return result
}
