package uuid

import (
	"testing"
)

func TestID(t *testing.T) {
	u := NewUUID()

	id1 := u.ID()
	id2 := u.ID()
	uuid1 := IDToUUID(id1)
	uuid2 := IDToUUID(id2)

	if uuid1 == uuid2 {
		t.Errorf("The IDs generated are identical: %d", uuid1)
	}

	t.Logf("uuid1: %s , uuid2: %s", uuid1, uuid2)

	const numIDs = 11
	idMap := make(map[uint64]bool)

	for i := 0; i < numIDs; i++ {
		id := u.ID()
		if _, exists := idMap[id]; exists {
			t.Errorf("Duplicate ID found: %d", id)
		}
		idMap[id] = true
		uuidStr := IDToUUID(id)
		t.Logf("Generated UUID %d: %s", i, uuidStr)
	}

	t.Logf("MachineID: %d", u.MachineID)
}

func TestEntropyIncrement(t *testing.T) {
	u := NewUUID()
	initialEntropyValue := u.entropy.Value()

	for i := 0; i < 10; i++ {
		u.ID()
	}

	if u.entropy.Value() == initialEntropyValue {
		t.Errorf("Entropy did not change correctly: expected a different value, got the same")
	} else {
		t.Logf("Entropy incremented correctly: initial %d, current %d", initialEntropyValue, u.entropy.Value())
	}
}
