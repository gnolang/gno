package dao_utils

import (
	"testing"
	"time"

	"gno.land/p/demo/json"
)

func TestMatch(t *testing.T) {
	ex := ExpirationNever{}
	switch Expiration(ex).(type) {
	case ExpirationNever:
		t.Log("ExpirationNever")
	default:
		t.Errorf("expected a match")
	}
}

func TestJSONNever(t *testing.T) {
	ex := ExpirationNever{}
	node := ex.ToJSON()
	b, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"never":{}}`
	if string(b) != expected {
		t.Errorf("expected %v, got %v", expected, string(b))
	}
}

func TestJSONAtTime(t *testing.T) {
	ex := ExpirationAtTime(time.Unix(1712830928, 0))
	node := ex.ToJSON()
	b, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"atTime":"2024-04-11T10:22:08Z"}`
	if string(b) != expected {
		t.Errorf("expected %v, got %v", expected, string(b))
	}
}

func TestJSONAtHeight(t *testing.T) {
	ex := ExpirationAtHeight(123456)
	node := ex.ToJSON()
	b, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"atHeight":"123456"}`
	if string(b) != expected {
		t.Errorf("expected %v, got %v", expected, string(b))
	}
}
