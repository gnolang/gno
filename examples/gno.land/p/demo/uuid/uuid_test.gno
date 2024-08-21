package uuid

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestID(t *testing.T) {
	u := NewUUID()
	id1 := u.ID()
	id2 := u.ID()
	uuid1 := IDToUUID(id1)
	uuid2 := IDToUUID(id2)

	uassert.NotEqual(t, uuid1, uuid2)

	idMap := make(map[uint64]bool)

	for i := 0; i < 101; i++ {
		id := u.ID()
		if _, exists := idMap[id]; exists {
			t.Errorf("Duplicate ID found: %d", id)
		}
		idMap[id] = true
		uuidStr := IDToUUID(id)
		t.Logf("Generated UUID %d: %s", i, uuidStr)
	}
}

func TestGetEntropy(t *testing.T) {
	u := NewUUID()
	initialEntropyValue := u.entropy.Value()

	for i := 0; i < 10; i++ {
		u.ID()
	}

	uassert.NotEqual(t, initialEntropyValue, u.entropy.Value())
}
