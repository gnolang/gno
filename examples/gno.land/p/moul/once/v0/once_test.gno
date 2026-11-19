package once

import (
	"errors"
	"testing"
)

func TestOnce_Do(t *testing.T) {
	counter := 0
	once := New()

	increment := func() {
		counter++
	}

	// First call should execute
	once.Do(increment)
	if counter != 1 {
		t.Errorf("expected counter to be 1, got %d", counter)
	}

	// Second call should not execute
	once.Do(increment)
	if counter != 1 {
		t.Errorf("expected counter to still be 1, got %d", counter)
	}
}

func TestOnce_DoErr(t *testing.T) {
	once := New()
	expectedErr := errors.New("test error")

	fn := func() error {
		return expectedErr
	}

	// First call should return error
	if err := once.DoErr(fn); err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	// Second call should return same error
	if err := once.DoErr(fn); err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestOnce_DoOrPanic(t *testing.T) {
	once := New()
	executed := false

	fn := func() {
		executed = true
	}

	// First call should execute
	once.DoOrPanic(fn)
	if !executed {
		t.Error("function should have executed")
	}

	// Second call should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on second execution")
		}
	}()
	once.DoOrPanic(fn)
}

func TestOnce_DoValue(t *testing.T) {
	once := New()
	expected := "test value"
	counter := 0

	fn := func() any {
		counter++
		return expected
	}

	// First call should return value
	if result := once.DoValue(fn); result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Second call should return cached value
	if result := once.DoValue(fn); result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	if counter != 1 {
		t.Errorf("function should have executed only once, got %d executions", counter)
	}
}

func TestOnce_DoValueErr(t *testing.T) {
	once := New()
	expectedVal := "test value"
	expectedErr := errors.New("test error")
	counter := 0

	fn := func() (any, error) {
		counter++
		return expectedVal, expectedErr
	}

	// First call should return value and error
	val, err := once.DoValueErr(fn)
	if val != expectedVal || err != expectedErr {
		t.Errorf("expected (%v, %v), got (%v, %v)", expectedVal, expectedErr, val, err)
	}

	// Second call should return cached value and error
	val, err = once.DoValueErr(fn)
	if val != expectedVal || err != expectedErr {
		t.Errorf("expected (%v, %v), got (%v, %v)", expectedVal, expectedErr, val, err)
	}

	if counter != 1 {
		t.Errorf("function should have executed only once, got %d executions", counter)
	}
}

func TestOnce_Reset(t *testing.T) {
	once := New()
	counter := 0

	fn := func() {
		counter++
	}

	once.Do(fn)
	if counter != 1 {
		t.Errorf("expected counter to be 1, got %d", counter)
	}

	once.Reset()
	once.Do(fn)
	if counter != 2 {
		t.Errorf("expected counter to be 2 after reset, got %d", counter)
	}
}

func TestOnce_IsDone(t *testing.T) {
	once := New()

	if once.IsDone() {
		t.Error("new Once instance should not be done")
	}

	once.Do(func() {})

	if !once.IsDone() {
		t.Error("Once instance should be done after execution")
	}
}

func TestOnce_Error(t *testing.T) {
	once := New()
	expectedErr := errors.New("test error")

	if err := once.Error(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	once.DoErr(func() error {
		return expectedErr
	})

	if err := once.Error(); err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestOnce_Value(t *testing.T) {
	once := New()

	// Test unexecuted state
	val, err := once.Value()
	if err != ErrNotExecuted {
		t.Errorf("expected ErrNotExecuted, got %v", err)
	}
	if val != nil {
		t.Errorf("expected nil value, got %v", val)
	}

	// Test after execution
	expected := "test value"
	once.DoValue(func() any {
		return expected
	})

	val, err = once.Value()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if val != expected {
		t.Errorf("expected value %v, got %v", expected, val)
	}
}

func TestOnce_DoValueErr_Panic_MarkedDone(t *testing.T) {
	once := New()
	count := 0
	fn := func() (any, error) {
		count++
		panic("panic")
	}
	var r any
	func() {
		defer func() { r = recover() }()
		once.DoValueErr(fn)
	}()
	if r == nil {
		t.Error("expected panic on first call")
	}
	if !once.IsDone() {
		t.Error("expected once to be marked as done after panic")
	}
	r = nil
	func() {
		defer func() { r = recover() }()
		once.DoValueErr(fn)
	}()
	if r != nil {
		t.Error("expected no panic on subsequent call")
	}
	if count != 1 {
		t.Errorf("expected count to be 1, got %d", count)
	}
}
