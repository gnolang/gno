package pool

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/gnoswap/consts"
)

func TestNewPoolParams(t *testing.T) {
	params := newPoolParams(
		"token0",
		"token1",
		500,       // 0.05% fee
		"1000000", // example sqrt price
	)

	if params.Token0Path() != "token0" {
		t.Errorf("Expected token0Path to be 'token0', got %s", params.Token0Path())
	}

	if params.TickSpacing() != 10 { // 500 fee should have 10 tick spacing
		t.Errorf("Expected tick spacing 10, got %d", params.TickSpacing())
	}

	if !params.isInOrder() {
		t.Errorf("Expected token0Path(token0) < token1Path(token1)")
	}

	params = newPoolParams(
		consts.GNOT,
		"token1",
		500,
		"1000000",
	)
	token0, token1 := params.wrap()
	if token0 != consts.WRAPPED_WUGNOT {
		t.Errorf("Expected GNOT to be wrapped to WUGNOT")
	}

	params = newPoolParams(
		"token0",
		"token0",
		500,
		"1000000",
	)
	if !params.isSameTokenPath() {
		t.Errorf("Expected token0Path(token0) == token1Path(token0)")
	}
}

func TestGetPoolPath(t *testing.T) {
	path := GetPoolPath("gno.land/r/onbloc/bar", "gno.land/r/onbloc/foo", 500)
	expected := "gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500"
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}

	path = GetPoolPath("gno.land/r/onbloc/foo", "gno.land/r/onbloc/bar", 500)
	if path != expected {
		t.Errorf("Expected tokens to be sorted, expected %s, got %s", expected, path)
	}
}

func TestTickSpacingMap(t *testing.T) {
	tests := []struct {
		fee         uint32
		tickSpacing int32
	}{
		{100, 1},     // 0.01%
		{500, 10},    // 0.05%
		{3000, 60},   // 0.3%
		{10000, 200}, // 1%
	}

	for _, tt := range tests {
		spacing := GetFeeAmountTickSpacing(tt.fee)
		if spacing != tt.tickSpacing {
			t.Errorf("For fee %d, expected tick spacing %d, got %d",
				tt.fee, tt.tickSpacing, spacing)
		}
	}
}

func TestCreatePool(t *testing.T) {
	tests := []struct {
		name        string
		token0Path  string
		token1Path  string
		fee         uint32
		sqrtPrice   string
		shouldPanic bool
		panicMsg    string
		inOrder     bool
	}{
		{
			name:       "success - normal token pair",
			token0Path: barPath,
			token1Path: fooPath,
			fee:        3000,
			sqrtPrice:  "4295128740",
		},
		{
			name:        "fail - same tokens",
			token0Path:  barPath,
			token1Path:  barPath,
			fee:         3000,
			sqrtPrice:   "4295128740",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-POOL-011] same token used in single pool || expected token0Path(gno.land/r/onbloc/bar) != token1Path(gno.land/r/onbloc/bar",
		},
		{
			name:        "success - when tokens not in order, reverse order and create pool",
			token0Path:  fooPath,
			token1Path:  bazPath,
			fee:         3000,
			sqrtPrice:   "4295128740",
			shouldPanic: false,
			panicMsg:    "[GNOSWAP-POOL-012] tokens must be in lexicographical order || expected token0Path(gno.land/r/onbloc/foo) < token1Path(gno.land/r/onbloc/baz)",
			inOrder:     true,
		},
		{
			name:        "fail - pool already exists",
			token0Path:  barPath,
			token1Path:  fooPath,
			fee:         3000,
			sqrtPrice:   "4295128740",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-POOL-013] pool already created || expected poolPath(gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:3000) not to exist",
		},
	}

	std.TestSetRealm(std.NewUserRealm(consts.ADMIN))
	SetPoolCreationFeeByAdmin(0)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("expected panic but got none")
						return
					}
					errMsg := r.(string)
					if !strings.Contains(errMsg, tt.panicMsg) {
						t.Errorf("expected panic message containing %q but got %q", tt.panicMsg, errMsg)
					}
				}()
			}

			CreatePool(tt.token0Path, tt.token1Path, tt.fee, tt.sqrtPrice)

			if !tt.shouldPanic {
				// verify pool was created correctly
				poolPath := GetPoolPath(tt.token0Path, tt.token1Path, tt.fee)
				pool := mustGetPool(poolPath)

				// check if GNOT was properly wrapped
				expectedToken0 := tt.token0Path
				expectedToken1 := tt.token1Path
				if tt.inOrder {
					expectedToken0, expectedToken1 = expectedToken1, expectedToken0
				}

				if expectedToken0 == consts.GNOT {
					expectedToken0 = consts.WRAPPED_WUGNOT
				}
				if expectedToken1 == consts.GNOT {
					expectedToken1 = consts.WRAPPED_WUGNOT
				}

				if pool.token0Path != expectedToken0 || pool.token1Path != expectedToken1 {
					t.Errorf("incorrect token paths in pool. got %s,%s want %s,%s",
						pool.token0Path, pool.token1Path, expectedToken0, expectedToken1)
				}
			}
		})
	}

	resetObject(t)
}

func TestGetPool(t *testing.T) {
	tests := []struct {
		name        string
		setupFn     func(t *testing.T)
		action      func(t *testing.T)
		shouldPanic bool
		expected    string
	}{
		{
			name: "Panic - unregisterd poolPath ",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
			},
			action: func(t *testing.T) {
				GetPool(barPath, fooPath, fee500)
			},
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-008] requested data not found || expected poolPath(gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500) to exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					if tt.shouldPanic {
						t.Errorf(">>> %s: expected panic but got none", tt.name)
						return
					}
				} else {
					switch r.(type) {
					case string:
						if r.(string) != tt.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tt.name, r, tt.expected)
						}
					case error:
						if r.(error).Error() != tt.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tt.name, r.(error).Error(), tt.expected)
						}
					default:
						t.Errorf(">>> %s: got panic %v, want %v", tt.name, r, tt.expected)
					}
				}
			}()
			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			if tt.shouldPanic {
				tt.action(t)
			} else {
			}
		})
	}
}
