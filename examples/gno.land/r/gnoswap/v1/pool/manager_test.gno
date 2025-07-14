package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
)

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
			name:       "success - wugnot <> gns pool",
			token0Path: "gno.land/r/demo/wugnot",
			token1Path: "gno.land/r/gnoswap/v1/gns",
			fee:        3000,
			sqrtPrice:  "79228162514264337593543950337",
		},
		{
			name:        "fail - same tokens",
			token0Path:  barPath,
			token1Path:  barPath,
			fee:         3000,
			sqrtPrice:   "4295128740",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-POOL-011] same token used in single pool || expected token0Path(gno.land/r/onbloc/bar) != token1Path(gno.land/r/onbloc/bar)",
		},
		{
			name:        "success - when tokens not in order, reverse order and create pool",
			token0Path:  fooPath,
			token1Path:  bazPath,
			fee:         3000,
			sqrtPrice:   "4295343490",
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
			panicMsg:    "[GNOSWAP-POOL-003] pool already created || expected poolPath(gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:3000) not to exist",
		},
		{
			name:        "fail - invalid fee amount",
			token0Path:  barPath,
			token1Path:  fooPath,
			fee:         1234,
			sqrtPrice:   "4295128740",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-POOL-002] unsupported fee tier || expected fee(1234) to be one of 100, 500, 3000, 10000",
		},
		{
			name:        "fail - invalid sqrtPrice (too small)",
			token0Path:  bazPath,
			token1Path:  barPath,
			fee:         3000,
			sqrtPrice:   "0",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-POOL-005] out of range for numeric value || sqrtPriceX96(0) is out of range",
		},
	}

	testing.SetRealm(std.NewUserRealm(adminAddr))
	SetPoolCreationFeeByAdmin(cross, 0)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				uassert.AbortsWithMessage(t, tt.panicMsg, func() {
					CreatePool(cross, tt.token0Path, tt.token1Path, tt.fee, tt.sqrtPrice)
				})
			} else {
				CreatePool(cross, tt.token0Path, tt.token1Path, tt.fee, tt.sqrtPrice)

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
		action      func(t *testing.T) *Pool
		shouldPanic bool
		expected    string
		verifyFn    func(t *testing.T, pool *Pool)
	}{
		{
			name: "Panic - unregistered poolPath",
			setupFn: func(t *testing.T) {
				testing.SetRealm(adminRealm)
				SetPoolCreationFeeByAdmin(cross, 0)
			},
			action: func(t *testing.T) *Pool {
				return GetPool(barPath, fooPath, fee500)
			},
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-008] requested data not found || expected poolPath(gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500) to exist",
		},
		{
			name: "success - get pool (bar:foo:500)",
			setupFn: func(t *testing.T) {
				testing.SetRealm(adminRealm)
				SetPoolCreationFeeByAdmin(cross, 0)
				CreatePool(cross, barPath, fooPath, fee500, "4295128740")
			},
			action: func(t *testing.T) *Pool {
				return GetPool(barPath, fooPath, fee500)
			},
			verifyFn: func(t *testing.T, pool *Pool) {
				if pool.token0Path != barPath {
					t.Errorf("expected token0Path %s, got %s", barPath, pool.token0Path)
				}
				if pool.token1Path != fooPath {
					t.Errorf("expected token1Path %s, got %s", fooPath, pool.token1Path)
				}
				if pool.fee != fee500 {
					t.Errorf("expected fee %d, got %d", fee500, pool.fee)
				}
				if pool.slot0.sqrtPriceX96.ToString() != "4295128740" {
					t.Errorf("expected sqrtPriceX96 %s, got %s", "4295128740", pool.slot0.sqrtPriceX96.ToString())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					if tt.shouldPanic {
						t.Errorf(">>> %s: expected panic but got none", tt.name)
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
			}

			if !tt.shouldPanic {
				pool := tt.action(t)
				if tt.verifyFn != nil {
					tt.verifyFn(t, pool)
				}
			}
		})
	}
}
