package ordering

type Order int

const (
	less    Order = -1
	equal   Order = 0
	greater Order = 1
)

var (
	Less    = Ordering{less}
	Equal   = Ordering{equal}
	Greater = Ordering{greater}
)

type Ordering struct {
	value Order
}

func NewOrdering(order Order) Ordering {
	return Ordering{value: order}
}

func (o Ordering) IsEqual() bool {
	return o == Equal
}

func (o Ordering) IsLess() bool {
	return o == Less
}

func (o Ordering) IsGreater() bool {
	return o == Greater
}
