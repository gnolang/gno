package cache

import (
	"container/list"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	start, end []byte
	items      []*std.KVPair
	ascending  bool
	pos        int
}

func newMemIterator(start, end []byte, items *list.List, ascending bool) *memIterator {
	itemsInDomain := make([]*std.KVPair, 0)
	var entered bool
	for e := items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*std.KVPair)
		if !dbm.IsKeyInDomain(item.Key, start, end) {
			if entered {
				break
			}
			continue
		}
		itemsInDomain = append(itemsInDomain, item)
		entered = true
	}

	pos := 0
	if !ascending {
		pos = len(itemsInDomain) - 1
	}

	return &memIterator{
		start:     start,
		end:       end,
		items:     itemsInDomain,
		ascending: ascending,
		pos:       pos,
	}
}

func (mi *memIterator) Domain() ([]byte, []byte) {
	return mi.start, mi.end
}

func (mi *memIterator) Valid() bool {
	if mi.pos < 0 || len(mi.items) <= mi.pos {
		return false
	}
	return true
}

func (mi *memIterator) assertValid() {
	if !mi.Valid() {
		panic("memIterator is invalid")
	}
}

func (mi *memIterator) Next() {
	mi.assertValid()
	if mi.ascending {
		mi.pos++
	} else {
		mi.pos--
	}
}

func (mi *memIterator) Key() []byte {
	mi.assertValid()
	return mi.items[mi.pos].Key
}

func (mi *memIterator) Value() []byte {
	mi.assertValid()
	return mi.items[mi.pos].Value
}

func (mi *memIterator) Close() {
	mi.start = nil
	mi.end = nil
	mi.items = nil
}
