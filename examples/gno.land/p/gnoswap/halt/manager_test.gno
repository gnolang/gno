package halt

import (
	"testing"
)

func TestDefaultManager(t *testing.T) {
	mgr := DefaultManager()

	if mgr.Level().Name() != "NoHalt" {
		t.Errorf("Expected initial level to be 'NoHalt', got '%s'", mgr.Level().Name())
	}

	// Check that basic operations are registered
	if _, exists := mgr.operations[OpTypeSwap]; !exists {
		t.Error("Expected OpTypeSwap to be registered")
	}
	if _, exists := mgr.operations[OpTypeLiquidity]; !exists {
		t.Error("Expected OpTypeLiquidity to be registered")
	}
	if _, exists := mgr.operations[OpTypeWithdraw]; !exists {
		t.Error("Expected OpTypeWithdraw to be registered")
	}
}

func TestNewManagerWithOptions(t *testing.T) {
	customOp := NewOperation(OpType("customOp"), "Custom Op", "Custom operation")

	customLevel := NewHaltLevel(LevelID(10), "CustomLevel", "Custom level", map[OpType]bool{
		OpType("customOp"): true,
	})

	mgr := NewManager(
		WithOperations([]Operation{customOp}),
		WithLevels([]HaltLevel{customLevel}),
		WithInitialLevel(LevelID(10)),
	)

	if _, exists := mgr.operations[OpType("customOp")]; !exists {
		t.Error("Expected custom operation to be registered")
	}

	if _, exists := mgr.levels[LevelID(10)]; !exists {
		t.Error("Expected custom level to be registered")
	}

	if mgr.currentLevel.Name() != "CustomLevel" {
		t.Errorf("Expected current level to be 'CustomLevel', got '%s'", mgr.currentLevel.Name())
	}
}

func TestWithOperationsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with invalid initial level, but got none")
		}
	}()

	NewManager(WithInitialLevel(LevelID(9)))
}

func TestSetCurrentLevel(t *testing.T) {
	mgr := DefaultManager()

	if err := mgr.SetCurrentLevel(LvSwapHalt); err != nil {
		t.Errorf("Failed to set current level to 1: %v", err)
	}
	if mgr.Level().Name() != "SwapHalt" {
		t.Errorf("Expected current level to be 'SwapHalt', got '%s'", mgr.Level().Name())
	}

	// Try setting an invalid level and expect an error
	if err := mgr.SetCurrentLevel(99); err == nil {
		t.Error("Expected error when setting an invalid halt level, but got nil")
	}
}

func TestStatus(t *testing.T) {
	mgr := DefaultManager()

	if err := mgr.SetCurrentLevel(LvEmergencyHalt); err != nil {
		t.Errorf("Failed to set current level to 2: %v", err)
	}

	// The Status method should return the name of the current level
	status := mgr.Status(OpTypeSwap) // The passed operation is ignored in Status
	if status != "EmergencyHalt" {
		t.Errorf("Expected status 'EmergencyHalt', got '%s'", status)
	}
}

func TestRegisterCustomHaltLevel(t *testing.T) {
	mgr := DefaultManager()

	// Create and register a custom halt level (level 5)
	const CustomHaltLevel = LevelID(5)

	customLevel := NewHaltLevel(CustomHaltLevel, "CustomHalt", "Custom halt level", map[OpType]bool{
		OpTypeSwap:      false,
		OpTypeLiquidity: false,
		OpTypeWithdraw:  false,
	})
	mgr.RegisterHaltLevel(customLevel)

	// Set current level to the custom level
	if err := mgr.SetCurrentLevel(CustomHaltLevel); err != nil {
		t.Errorf("Failed to set current level to custom level 3: %v", err)
	}
	if mgr.Level().Name() != "CustomHalt" {
		t.Errorf("Expected current level to be 'CustomHalt', got '%s'", mgr.Level().Name())
	}
}

func TestRegisterOperation(t *testing.T) {
	mgr := DefaultManager()

	// Create and register a new operation
	newOp := NewOperation(OpType("newOp"), "New Operation", "Test new operation")
	mgr.RegisterOperation(newOp)

	// Check that the new operation is registered correctly
	registeredOp, exists := mgr.operations[newOp.Type()]
	if !exists {
		t.Error("Expected new operation to be registered")
	}
	if registeredOp.Name() != "New Operation" {
		t.Errorf("Expected new operation name to be 'New Operation', got '%s'", registeredOp.Name())
	}
}
