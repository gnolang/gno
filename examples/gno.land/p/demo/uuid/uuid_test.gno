package uuid

import (
	"testing"

	"gno.land/p/demo/ufmt"
)

// TestNextID vérifie la génération correcte des IDs Snowflake.
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

// TestInvalidMachineID teste la gestion des IDs de machine non valides.
func TestInvalidMachineID(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			ufmt.Sprintf("Setting an invalid machine ID should cause a panic")
		}
	}()

	// Définit un ID de machine supérieur à la limite maximale.
	SetMachineID(MaxMachineID + 1)
}
