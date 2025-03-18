package halt

import (
	"testing"
)

func TestSetOperationStatus(t *testing.T) {
	t.Run("Basic operation toggling", func(t *testing.T) {
		mgr := DefaultManager()
		// Initially pool should be allowed at NoHalt level
		poolOp := NewOperation(OpTypePool, "Pool", "Pool management operations")
		if !mgr.Level().IsOperationAllowed(poolOp) {
			t.Errorf("Expected pool operation to be allowed initially")
		}

		// Disable swap operation
		err := mgr.SetOperationStatus(OpTypePool, false)
		if err != nil {
			t.Errorf("Failed to set operation status: %v", err)
		}

		// Verify pool is now disabled
		if mgr.Level().IsOperationAllowed(poolOp) {
			t.Errorf("Expected pool operation to be disabled after SetOperationStatus(false)")
		}

		// Re-enable swap operation
		err = mgr.SetOperationStatus(OpTypePool, true)
		if err != nil {
			t.Errorf("Failed to re-enable operation: %v", err)
		}

		// Verify pool is enabled again
		if !mgr.Level().IsOperationAllowed(poolOp) {
			t.Errorf("Expected pool operation to be re-enabled after SetOperationStatus(true)")
		}
	})

	t.Run("Custom operation handling", func(t *testing.T) {
		const OpTypeStake OpType = "stake"

		customOp := NewOperation(OpTypeStake, "Stake Tokens", "Stake tokens for rewards")

		// Create a manager with default levels
		mgr := DefaultManager()
		mgr.RegisterOperation(customOp)

		// Initially the custom operation should not be allowed
		if mgr.Level().IsOperationAllowed(customOp) {
			t.Errorf("Expected custom operation to be disabled initially")
		}

		// Enable the custom operation at current level (NoHalt)
		err := mgr.SetOperationStatus(OpTypeStake, true)
		if err != nil {
			t.Errorf("Failed to enable custom operation: %v", err)
		}

		// Verify custom operation is now enabled
		if !mgr.Level().IsOperationAllowed(customOp) {
			t.Errorf("Expected custom operation to be enabled after SetOperationStatus")
		}

		// Disable the custom operation
		err = mgr.SetOperationStatus(OpTypeStake, false)
		if err != nil {
			t.Errorf("Failed to disable custom operation: %v", err)
		}

		// Verify custom operation is now disabled
		if mgr.Level().IsOperationAllowed(customOp) {
			t.Errorf("Expected custom operation to be disabled after SetOperationStatus(false)")
		}
	})

	t.Run("Operation not previously in allowedOps map", func(t *testing.T) {
		mgr := DefaultManager()

		// Operation that wasn't previously in the allowedOps map
		newOpType := OpType("newOperation")
		newOp := NewOperation(newOpType, "New Op", "New operation type")

		// Attempt to enable the new operation type
		err := mgr.SetOperationStatus(newOpType, true)
		if err != nil {
			t.Errorf("Failed to set status for new operation type: %v", err)
		}

		// Verify operation is now allowed
		if !mgr.Level().IsOperationAllowed(newOp) {
			t.Errorf("Expected new operation to be allowed after adding it")
		}
	})

	t.Run("Preservation across level changes", func(t *testing.T) {
		mgr := DefaultManager()

		// Modify NoHalt level to disallow pools
		err := mgr.SetOperationStatus(OpTypePool, false)
		if err != nil {
			t.Errorf("Failed to disable pool: %v", err)
		}

		// Change level and then switch back
		if err := mgr.SetCurrentLevel(LvContractHalt); err != nil {
			t.Errorf("Failed to change level: %v", err)
		}
		if err := mgr.SetCurrentLevel(LvNoHalt); err != nil {
			t.Errorf("Failed to change back to original level: %v", err)
		}

		// Verify the modification persisted
		poolOp := NewOperation(OpTypePool, "Pool", "Pool management operations")
		if mgr.Level().IsOperationAllowed(poolOp) {
			t.Errorf("Expected pool to remain disabled after level change and back")
		}
	})

	t.Run("Composite halt level should return error", func(t *testing.T) {
		// Create basic levels for the composite
		level1 := NewHaltLevel(LevelID(100), "Level1", "Test level 1", map[OpType]bool{
			OpTypePool: true,
		})
		level2 := NewHaltLevel(LevelID(101), "Level2", "Test level 2", map[OpType]bool{
			OpTypePool: false,
		})

		// Create a composite level
		compositeLevel := &CompositeHaltLevel{
			baseInfo: baseInfo{name: "Composite", desc: "Composite level"},
			levels:   HaltLevels{level1, level2},
			operator: CompositeOpAnd,
		}

		// Create a manager
		mgr := &Manager{
			currentLevel: compositeLevel,
			levels:       make(map[LevelID]HaltLevel),
			operations:   make(map[OpType]Operation),
		}

		// Attempt to modify composite level should fail
		err := mgr.SetOperationStatus(OpTypePool, true)
		if err == nil {
			t.Error("Expected error when trying to modify composite halt level, got nil")
		}
	})
}

func TestSetOperationStatus_ToggleOperations(t *testing.T) {
	tests := []struct {
		name      string
		setupFn   func() *Manager
		opType    OpType
		newStatus bool
		wantErr   bool
		verifyFn  func(*testing.T, *Manager)
	}{
		{
			name: "disable withdraw in NoHalt level",
			setupFn: func() *Manager {
				mgr := DefaultManager()
				// NoHalt level should have withdraw enabled by default
				return mgr
			},
			opType:    OpTypeWithdraw,
			newStatus: false,
			wantErr:   false,
			verifyFn: func(t *testing.T, mgr *Manager) {
				withdrawOp := NewOperation(OpTypeWithdraw, "test withdraw", "withdraw operation")
				if mgr.Level().IsOperationAllowed(withdrawOp) {
					t.Error("Expected withdraw to be disabled, but it's still allowed")
				}
			},
		},
		{
			name: "enable swap in SwapHalt level",
			setupFn: func() *Manager {
				mgr := DefaultManager()
				if err := mgr.SetCurrentLevel(LvContractHalt); err != nil {
					t.Fatalf("Failed to set up test: %v", err)
				}
				// ContractHalt level should have swap disabled by default
				return mgr
			},
			opType:    OpTypePool,
			newStatus: true,
			wantErr:   false,
			verifyFn: func(t *testing.T, mgr *Manager) {
				poolOp := NewOperation(OpTypePool, "test pool", "pool management operations")
				if !mgr.Level().IsOperationAllowed(poolOp) {
					t.Error("Expected pool to be enabled, but it's still disallowed")
				}
			},
		},
		{
			name: "toggle multiple operations in NoHalt level",
			setupFn: func() *Manager {
				mgr := DefaultManager()
				return mgr
			},
			opType:    OpTypePool, // We'll toggle multiple in verifyFn
			newStatus: false,      // Initial change for pool
			wantErr:   false,
			verifyFn: func(t *testing.T, mgr *Manager) {
				// First verify pool was disabled
				poolOp := NewOperation(OpTypePool, "test pool", "pool management operations")
				if mgr.Level().IsOperationAllowed(poolOp) {
					t.Error("Expected pool to be disabled, but it's still allowed")
				}

				// Toggle withdraw and verify
				if err := mgr.SetOperationStatus(OpTypeWithdraw, false); err != nil {
					t.Errorf("Unexpected error toggling withdraw: %v", err)
				}

				withdrawOp := NewOperation(OpTypeWithdraw, "test withdraw", "withdraw operation")
				if mgr.Level().IsOperationAllowed(withdrawOp) {
					t.Error("Expected withdraw to be disabled, but it's still allowed")
				}
			},
		},
		{
			name: "test with non-existent operation type",
			setupFn: func() *Manager {
				mgr := DefaultManager()
				return mgr
			},
			opType:    OpType("nonexistent"),
			newStatus: true,
			wantErr:   false,
			verifyFn: func(t *testing.T, mgr *Manager) {
				// The operation should be added to the allowedOps map
				customOp := NewOperation(OpType("nonexistent"), "test custom", "custom operation")
				if !mgr.Level().IsOperationAllowed(customOp) {
					t.Error("Expected custom operation to be enabled, but it's not allowed")
				}
			},
		},
		{
			name: "test with composite halt level",
			setupFn: func() *Manager {
				mgr := DefaultManager()

				// Create a composite level
				level1 := BasicHaltLevel{
					baseInfo:   baseInfo{name: "level1", desc: "test level 1"},
					level:      LevelID(100),
					allowedOps: allowedOps{OpTypePool: true, OpTypePosition: false},
				}

				level2 := BasicHaltLevel{
					baseInfo:   baseInfo{name: "level2", desc: "test level 2"},
					level:      LevelID(101),
					allowedOps: allowedOps{OpTypePool: false, OpTypePosition: true},
				}

				composite := &CompositeHaltLevel{
					baseInfo: baseInfo{name: "composite", desc: "composite level"},
					levels:   HaltLevels{level1, level2},
					operator: CompositeOpAnd,
				}

				mgr.currentLevel = composite

				return mgr
			},
			opType:    OpTypePool,
			newStatus: true,
			wantErr:   true, // Should error on composite level
			verifyFn:  nil,  // No verification needed for error case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := tt.setupFn()

			err := mgr.SetOperationStatus(tt.opType, tt.newStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetOperationStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.verifyFn != nil {
				tt.verifyFn(t, mgr)
			}
		})
	}
}

func TestSetOperationStatusScenarios(t *testing.T) {
	t.Run("emergency_scenario", func(t *testing.T) {
		mgr := DefaultManager()
		if err := mgr.SetCurrentLevel(LvCompleteHalt); err != nil {
			t.Fatalf("Failed to set halt level: %v", err)
		}

		withdrawOp := createOperation(OpTypeWithdraw)
		assertOperationStatus(t, mgr, withdrawOp, true)

		// Other operations should remain disabled
		poolOp := createOperation(OpTypePool)
		positionOp := createOperation(OpTypePosition)
		assertOperationStatus(t, mgr, poolOp, false)
		assertOperationStatus(t, mgr, positionOp, false)
	})

	t.Run("maintenance_scenario", func(t *testing.T) {
		mgr := DefaultManager()

		// enable all operations
		setOperationStatusOrFail(t, mgr, OpTypePool, true)
		setOperationStatusOrFail(t, mgr, OpTypeWithdraw, true)

		poolOp := createOperation(OpTypePool)
		assertOperationStatus(t, mgr, poolOp, true)

		setOperationStatusOrFail(t, mgr, OpTypePool, false)
		assertOperationStatus(t, mgr, poolOp, false)

		// Other operations should remain enabled
		withdrawOp := createOperation(OpTypeWithdraw)
		assertOperationStatus(t, mgr, withdrawOp, true)

		setOperationStatusOrFail(t, mgr, OpTypePool, true)
		assertOperationStatus(t, mgr, poolOp, true)
	})

	t.Run("custom_operation_scenario", func(t *testing.T) {
		mgr := DefaultManager()

		// enable all operations
		setOperationStatusOrFail(t, mgr, OpTypePool, true)
		setOperationStatusOrFail(t, mgr, OpTypeWithdraw, true)

		const OpTypeStake = OpType("stake")
		stakeOp := createOperation(OpTypeStake)

		assertOperationStatus(t, mgr, stakeOp, false)

		setOperationStatusOrFail(t, mgr, OpTypeStake, true)
		assertOperationStatus(t, mgr, stakeOp, true)

		poolOp := createOperation(OpTypePool)
		assertOperationStatus(t, mgr, poolOp, true)
	})
}

func assertOperationStatus(t *testing.T, mgr *Manager, op Operation, expectedAllowed bool) {
	t.Helper()
	if mgr.Level().IsOperationAllowed(op) != expectedAllowed {
		if expectedAllowed {
			t.Errorf("Expected %s operation to be enabled", op.Type())
		} else {
			t.Errorf("Expected %s operation to be disabled", op.Type())
		}
	}
}

func setOperationStatusOrFail(t *testing.T, mgr *Manager, opType OpType, status bool) {
	t.Helper()
	if err := mgr.SetOperationStatus(opType, status); err != nil {
		t.Fatalf("Failed to set %s operation status to %v: %v", opType, status, err)
	}
}

func createOperation(opType OpType) Operation {
	return NewOperation(opType, string(opType), string(opType)+" operation")
}
