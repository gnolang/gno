package command

import (
	"reflect"
	"testing"
	"time"
)

func TestApplyFlagToFieldReflectString(t *testing.T) {
	// Test 1: time.Duration
	duration := time.Duration(0)
	field := reflect.ValueOf(&duration)
	flagValue := "10s"
	err := applyFlagToFieldReflectString(field, flagValue)
	if err != nil {
		t.Fatal(err)
	}
	if duration.String() != flagValue {
		t.Errorf("Expected %s, got %s", flagValue, duration)
	}
}
