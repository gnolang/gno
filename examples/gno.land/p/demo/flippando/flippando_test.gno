package flippando

import (
	"math/rand"
	"sort"
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
)

func TestFlippandoV1(t *testing.T) {
	// Set the current implementation to V1
	SetImplementation(&FlippandoV1{})

	// Test StartGame for V1
	if result := StartGame(); result != "V1: Game Started" {
		t.Errorf("Expected 'V1: Game Started', got '%s'", result)
	}

	// Test EndGame for V1
	if result := EndGame(); result != "V1: Game Ended" {
		t.Errorf("Expected 'V1: Game Ended', got '%s'", result)
	}
}

func TestFlippandoV2(t *testing.T) {
	// Set the current implementation to V2
	SetImplementation(&FlippandoV2{})

	// Test StartGame for V2
	if result := StartGame(); result != "V2: Game Started" {
		t.Errorf("Expected 'V2: Game Started', got '%s'", result)
	}

	// Test EndGame for V2
	if result := EndGame(); result != "V2: Game Ended" {
		t.Errorf("Expected 'V2: Game Ended', got '%s'", result)
	}

	// Test GetUserGames for V2
	v2 := &FlippandoV2{}
	if result := v2.GetUserGames(); result != "V2: Fetching user games..." {
		t.Errorf("Expected 'V2: Fetching user games...', got '%s'", result)
	}
}
