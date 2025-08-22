package internal

import "github.com/gnolang/gno/tm2/pkg/db"

// We need a copy of all of the keys.
// Not the best, but probably not a bottleneck depending.
type MemIterator struct {
	db    db.DB
	cur   int
	keys  []string
	start []byte
	end   []byte
	err   error
}

var _ db.Iterator = (*MemIterator)(nil)

// Keys is expected to be in reverse order for reverse iterators.
func NewMemIterator(db db.DB, keys []string, start, end []byte) *MemIterator {
	return &MemIterator{
		db:    db,
		cur:   0,
		keys:  keys,
		start: start,
		end:   end,
	}
}

// Implements Iterator.
func (itr *MemIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Implements Iterator.
func (itr *MemIterator) Valid() bool {
	return 0 <= itr.cur && itr.cur < len(itr.keys)
}

// Implements Iterator.
func (itr *MemIterator) Next() {
	itr.assertIsValid()
	itr.cur++
}

// Implements Iterator.
func (itr *MemIterator) Key() []byte {
	itr.assertIsValid()
	return []byte(itr.keys[itr.cur])
}

// Implements Iterator.
func (itr *MemIterator) Value() []byte {
	itr.assertIsValid()
	key := []byte(itr.keys[itr.cur])
	v, err := itr.db.Get(key)
	if err != nil {
		itr.err = err
	}
	return v
}

func (itr *MemIterator) Error() error {
	return itr.err
}

// Implements Iterator.
func (itr *MemIterator) Close() error {
	itr.keys = nil
	itr.db = nil
	return nil
}

func (itr *MemIterator) assertIsValid() {
	if !itr.Valid() {
		panic("MemIterator is invalid")
	}
}
