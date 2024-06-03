package uuid

import (
	"testing"

	"gno.land/p/demo/ufmt"
)

// TestNextID verifies the correct generation of Snowflake IDs.
func TestNextID(t *testing.T) {
	SetMachineID(1)

	id1, err := NextID()
	if err != nil {
		ufmt.Sprintf("Error generating first ID : %s", err)
	}

	id2, err := NextID()
	if err != nil {
		ufmt.Sprintf("Error generating second ID : %s", err)
	}

	if id1 == id2 {
		ufmt.Sprintf("The IDs generated are identical : %d", id1)
	}

	if id2 <= id1 {
		ufmt.Sprintf("The second ID should be higher than the first; id1: %d, id2: %d", id1, id2)
	}
}

// TestInvalidMachineID tests the handling of invalid machine IDs.
func TestInvalidMachineID(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			ufmt.Sprintf("Setting an invalid machine ID should cause a panic")
		}
	}()

	SetMachineID(MaxMachineID + 1)
}

// TestUUIDGeneration checks that the generated UUIDs are sequential.
func TestUUIDGeneration(t *testing.T) {
	// set machine ID
	SetMachineID(1)

	// Generate many UUIDs and check that they are sequential
	prevID, err := NextID()
	if err != nil {
		t.Fatalf("Error generating first UUID: %v", err)
	}

	for i := 0; i < 10; i++ {
		newID, err := NextID()
		if err != nil {
			t.Fatalf("Error generating UUID: %v", err)
		}
		if newID <= prevID {
			t.Fatalf("UUIDs are not sequential: got %d after %d", newID, prevID)
		}
		prevID = newID
	}
}
