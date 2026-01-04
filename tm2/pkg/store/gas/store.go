package gas

import (
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/gnolang/gno/tm2/pkg/store/utils"
)

var _ types.Store = &Store{}

// Store applies gas tracking to an underlying Store. It implements the
// Store interface.
type Store struct {
	gasMeter  types.GasMeter
	gasConfig types.GasConfig
	parent    types.Store
}

// New returns a reference to a new GasStore.
func New(parent types.Store, gasMeter types.GasMeter, gasConfig types.GasConfig) *Store {
	kvs := &Store{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
	}
	return kvs
}

// Implements Store.
func (gs *Store) Get(key []byte) (value []byte) {
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostFlat, types.GasReadCostFlatDesc)
	value = gs.parent.Get(key)

	gas := overflow.Mulp(gs.gasConfig.ReadCostPerByte, types.Gas(len(value)))
	gs.gasMeter.ConsumeGas(gas, types.GasReadPerByteDesc)

	return value
}

// Implements Store.
func (gs *Store) Set(key []byte, value []byte) {
	types.AssertValidValue(value)
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostFlat, types.GasWriteCostFlatDesc)

	gas := overflow.Mulp(gs.gasConfig.WriteCostPerByte, types.Gas(len(value)))
	gs.gasMeter.ConsumeGas(gas, types.GasWritePerByteDesc)
	gs.parent.Set(key, value)
}

// Implements Store.
func (gs *Store) Has(key []byte) bool {
	gs.gasMeter.ConsumeGas(gs.gasConfig.HasCost, types.GasHasDesc)
	return gs.parent.Has(key)
}

// Implements Store.
func (gs *Store) Delete(key []byte) {
	// charge gas to prevent certain attack vectors even though space is being freed
	gs.gasMeter.ConsumeGas(gs.gasConfig.DeleteCost, types.GasDeleteDesc)
	gs.parent.Delete(key)
}

// Iterator implements the Store interface. It returns an iterator which
// incurs a flat gas cost for seeking to the first key/value pair and a variable
// gas cost based on the current value's length if the iterator is valid.
func (gs *Store) Iterator(start, end []byte) types.Iterator {
	return gs.iterator(start, end, true)
}

// ReverseIterator implements the Store interface. It returns a reverse
// iterator which incurs a flat gas cost for seeking to the first key/value pair
// and a variable gas cost based on the current value's length if the iterator
// is valid.
func (gs *Store) ReverseIterator(start, end []byte) types.Iterator {
	return gs.iterator(start, end, false)
}

// Implements Store.
func (gs *Store) CacheWrap() types.Store {
	panic("cannot CacheWrap a gas.Store")
}

// Implements Store.
func (gs *Store) Write() {
	gs.parent.Write()
}

func (gs *Store) iterator(start, end []byte, ascending bool) types.Iterator {
	var parent types.Iterator
	if ascending {
		parent = gs.parent.Iterator(start, end)
	} else {
		parent = gs.parent.ReverseIterator(start, end)
	}

	gi := newGasIterator(gs.gasMeter, gs.gasConfig, parent)
	if gi.Valid() {
		gi.(*gasIterator).consumeSeekGas()
	}

	return gi
}

func (gs *Store) Print() {
	if ps, ok := gs.parent.(types.Printer); ok {
		ps.Print()
	} else {
		utils.Print(gs.parent)
	}
}

func (gs *Store) Flush() {
	if cts, ok := gs.parent.(types.Flusher); ok {
		cts.Flush()
	} else {
		panic("underlying store does not implement Flush()")
	}
}

type gasIterator struct {
	gasMeter  types.GasMeter
	gasConfig types.GasConfig
	parent    types.Iterator
}

func newGasIterator(gasMeter types.GasMeter, gasConfig types.GasConfig, parent types.Iterator) types.Iterator {
	return &gasIterator{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
	}
}

// Implements Iterator.
func (gi *gasIterator) Domain() (start []byte, end []byte) {
	return gi.parent.Domain()
}

// Implements Iterator.
func (gi *gasIterator) Valid() bool {
	return gi.parent.Valid()
}

// Next implements the Iterator interface. It seeks to the next key/value pair
// in the iterator. It incurs a flat gas cost for seeking and a variable gas
// cost based on the current value's length if the iterator is valid.
func (gi *gasIterator) Next() {
	if gi.Valid() {
		gi.consumeSeekGas()
	}

	gi.parent.Next()
}

// Key implements the Iterator interface. It returns the current key and it does
// not incur any gas cost.
func (gi *gasIterator) Key() (key []byte) {
	key = gi.parent.Key()
	return key
}

// Value implements the Iterator interface. It returns the current value and it
// does not incur any gas cost.
func (gi *gasIterator) Value() (value []byte) {
	value = gi.parent.Value()
	return value
}

func (gi *gasIterator) Error() error {
	return gi.parent.Error()
}

// Implements Iterator.
func (gi *gasIterator) Close() error {
	return gi.parent.Close()
}

// consumeSeekGas consumes a flat gas cost for seeking and a variable gas cost
// based on the current value's length.
func (gi *gasIterator) consumeSeekGas() {
	value := gi.Value()
	gas := overflow.Mulp(gi.gasConfig.ReadCostPerByte, types.Gas(len(value)))
	gi.gasMeter.ConsumeGas(gi.gasConfig.IterNextCostFlat, types.GasIterNextCostFlatDesc)
	gi.gasMeter.ConsumeGas(gas, types.GasValuePerByteDesc)
}
