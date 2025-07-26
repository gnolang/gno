package coinflip

import (
	"testing"
)

func TestGetRealmBalance(t *testing.T) {
	value := GetRealmBalance(cross)

	if value != 0 {
		t.Fatalf("Expected 0, got %v", value)
	}
}

func TestFlip(t *testing.T) {
	value := flip(1)

	if value != true {
		t.Fatalf("Expected true, got %v", value)
	}

	value = flip(0)

	if value != false {
		t.Fatalf("Expected false, got %v", value)
	}
}
