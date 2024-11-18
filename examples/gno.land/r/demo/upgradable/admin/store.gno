package admin

type Store interface {
	GetCounter() uint64
	SetCounter(value uint64)
}

var store func() Store = nil

func RegisterStore(newStore func() Store) {
	if store != nil {
		panic("Store already registered")
	}
	store = newStore
}