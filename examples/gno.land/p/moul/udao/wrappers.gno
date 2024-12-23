package udao

import (
	"gno.land/p/demo/avl/list"
)

// PropListWrapper wraps an IReadOnlyList to implement PropList
type PropListWrapper struct {
	list list.IList
}

// WrapAsPropList converts an IReadOnlyList to a PropList
func WrapAsPropList(list list.IList) PropList {
	return &PropListWrapper{list: list}
}

func (pl *PropListWrapper) Len() int {
	return pl.list.Len()
}

func (pl *PropListWrapper) Get(index int) Proposal {
	return pl.list.Get(index).(Proposal)
}

func (pl *PropListWrapper) Slice(startIndex, endIndex int) []Proposal {
	slice := pl.list.Slice(startIndex, endIndex)
	result := make([]Proposal, len(slice))
	for i, v := range slice {
		result[i] = v.(Proposal)
	}
	return result
}
