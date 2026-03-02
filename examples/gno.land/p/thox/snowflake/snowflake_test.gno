package snowflake

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestID(t *testing.T) {
	u := NewSnowflakeID()
	id1 := u.GenerateID()
	id2 := u.GenerateID()
	uuid1 := SnowflakeIDToUUIDString(id1)
	uuid2 := SnowflakeIDToUUIDString(id2)

	uassert.NotEqual(t, uuid1, uuid2)

	idMap := make(map[uint64]bool)
	uuidMap := make(map[string]bool)

	for i := 0; i < 101; i++ {
		id := u.GenerateID()
		if _, exists := idMap[id]; exists {
			t.Errorf("Duplicate ID found: %d", id)
		}
		idMap[id] = true

		uuidStr := SnowflakeIDToUUIDString(id)
		if _, exists := uuidMap[uuidStr]; exists {
			t.Errorf("Duplicate UUID found: %s", uuidStr)
		}
		uuidMap[uuidStr] = true
		t.Logf("Generated UUID %d: %s", i, uuidStr)
	}
}

func TestGetEntropy(t *testing.T) {
	u := NewSnowflakeID()
	initialEntropyValue := u.entropy.Value()

	for i := 0; i < 10; i++ {
		u.GenerateID()
	}

	uassert.NotEqual(t, initialEntropyValue, u.entropy.Value())
}

// Output:
// Generated UUID 0: 0007a63f-a542-4310-807a-63fa54253102
// Generated UUID 1: 0007a63f-a542-43fe-807a-63fa54253fe1
// Generated UUID 2: 0007a63f-a542-43e3-807a-63fa54253e34
// Generated UUID 3: 0007a63f-a542-432f-807a-63fa542532fb
// Generated UUID 4: 0007a63f-a542-4353-807a-63fa54253536
// Generated UUID 5: 0007a63f-a542-43be-807a-63fa54253be5
// Generated UUID 6: 0007a63f-a542-43e0-807a-63fa54253e08
// Generated UUID 7: 0007a63f-a542-4329-807a-63fa5425329f
// Generated UUID 8: 0007a63f-a542-430a-807a-63fa542530aa
// Generated UUID 9: 0007a63f-a542-43f2-807a-63fa54253f29
// Generated UUID 10: 0007a63f-a542-4351-807a-63fa5425351c
// .......
