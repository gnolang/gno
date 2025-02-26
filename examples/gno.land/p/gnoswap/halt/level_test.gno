package halt

import (
	"testing"
)

func TestBasicHaltLevel_IsOperationAllowed(t *testing.T) {
	tests := []struct {
		name        string
		level       BasicHaltLevel
		operation   Operation
		wantAllowed bool
	}{
		{
			name: "should allow swap operation when configured",
			level: BasicHaltLevel{
				baseInfo:   baseInfo{name: "test", desc: "test level"},
				level:      1,
				allowedOps: allowedOps{OpTypeSwap: true},
			},
			operation:   NewOperation(OpTypeSwap, "test swap", "swap operation"),
			wantAllowed: true,
		},
		{
			name: "should not allow swap operation when not configured",
			level: BasicHaltLevel{
				baseInfo:   baseInfo{name: "test", desc: "test level"},
				level:      1,
				allowedOps: allowedOps{OpTypeLiquidity: true},
			},
			operation:   NewOperation(OpTypeSwap, "test swap", "swap operation"),
			wantAllowed: false,
		},
		{
			name: "should not allow operation when allowed is false",
			level: BasicHaltLevel{
				baseInfo:   baseInfo{name: "test", desc: "test level"},
				level:      1,
				allowedOps: allowedOps{OpTypeSwap: false},
			},
			operation:   NewOperation(OpTypeSwap, "test swap", "swap operation"),
			wantAllowed: false,
		},
		{
			name: "should handle multiple operations correctly",
			level: BasicHaltLevel{
				baseInfo: baseInfo{name: "test", desc: "test level"},
				level:    1,
				allowedOps: allowedOps{
					OpTypeSwap:      true,
					OpTypeLiquidity: true,
					OpTypeWithdraw:  false,
				},
			},
			operation:   NewOperation(OpTypeLiquidity, "test liquidity", "liquidity operation"),
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.IsOperationAllowed(tt.operation); got != tt.wantAllowed {
				t.Errorf("BasicHaltLevel.IsOperationAllowed() = %v, want %v", got, tt.wantAllowed)
			}
		})
	}
}

func TestCompositeHaltLevel_IsOperationAllowed(t *testing.T) {
	// basic levels
	level1 := BasicHaltLevel{
		baseInfo:   baseInfo{name: "level1", desc: "test level 1"},
		level:      1,
		allowedOps: allowedOps{OpTypeSwap: true, OpTypeLiquidity: false},
	}
	level2 := BasicHaltLevel{
		baseInfo:   baseInfo{name: "level2", desc: "test level 2"},
		level:      2,
		allowedOps: allowedOps{OpTypeSwap: false, OpTypeLiquidity: true},
	}

	tests := []struct {
		name        string
		composite   CompositeHaltLevel
		operation   Operation
		wantAllowed bool
	}{
		{
			name: "AND operator - should return true when all levels allow",
			composite: CompositeHaltLevel{
				baseInfo: baseInfo{name: "composite and", desc: "test composite and"},
				levels: HaltLevels{
					level1,
					BasicHaltLevel{
						baseInfo:   baseInfo{name: "level3", desc: "test level 3"},
						level:      3,
						allowedOps: allowedOps{OpTypeSwap: true},
					},
				},
				operator: CompositeOpAnd,
			},
			operation:   NewOperation(OpTypeSwap, "test swap", "swap operation"),
			wantAllowed: true,
		},
		{
			name: "AND operator - should return false when any level denies",
			composite: CompositeHaltLevel{
				baseInfo: baseInfo{name: "composite and", desc: "test composite and"},
				levels:   HaltLevels{level1, level2},
				operator: CompositeOpAnd,
			},
			operation:   NewOperation(OpTypeSwap, "test swap", "swap operation"),
			wantAllowed: false,
		},
		{
			name: "OR operator - should return true when any level allows",
			composite: CompositeHaltLevel{
				baseInfo: baseInfo{name: "composite or", desc: "test composite or"},
				levels:   HaltLevels{level1, level2},
				operator: CompositeOpOr,
			},
			operation:   NewOperation(OpTypeSwap, "test swap", "swap operation"),
			wantAllowed: true,
		},
		{
			name: "OR operator - should return false when all levels deny",
			composite: CompositeHaltLevel{
				baseInfo: baseInfo{name: "composite or", desc: "test composite or"},
				levels:   HaltLevels{level1, level2},
				operator: CompositeOpOr,
			},
			operation:   NewOperation(OpTypeGovernance, "test governance", "governance operation"),
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.composite.IsOperationAllowed(tt.operation); got != tt.wantAllowed {
				t.Errorf("CompositeHaltLevel.IsOperationAllowed() = %v, want %v", got, tt.wantAllowed)
			}
		})
	}
}

func TestCompositeOpFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CompositeOp
		wantErr bool
	}{
		{
			name:    "should parse 'and' operator correctly",
			input:   "and",
			want:    CompositeOpAnd,
			wantErr: false,
		},
		{
			name:    "should parse 'or' operator correctly",
			input:   "or",
			want:    CompositeOpOr,
			wantErr: false,
		},
		{
			name:    "should return error for invalid operator",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "should return error for empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompositeOpFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompositeOpFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompositeOpFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
