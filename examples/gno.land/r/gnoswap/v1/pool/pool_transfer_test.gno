package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
)

func TestTransferAndVerify(t *testing.T) {
	// Setup common test data
	pool := &Pool{
		balances: Balances{
			token0: u256.NewUint(1000),
			token1: u256.NewUint(1000),
		},
	}

	t.Run("validatePoolBalance", func(t *testing.T) {
		tests := []struct {
			name          string
			amount        *u256.Uint
			isToken0      bool
			expectedError bool
		}{
			{
				name:          "must success for negative amount",
				amount:        u256.NewUint(500),
				isToken0:      true,
				expectedError: false,
			},
			{
				name:          "must panic for insufficient token0 balance",
				amount:        u256.NewUint(1500),
				isToken0:      true,
				expectedError: true,
			},
			{
				name:          "must success for negative amount",
				amount:        u256.NewUint(500),
				isToken0:      false,
				expectedError: false,
			},
			{
				name:          "must panic for insufficient token1 balance",
				amount:        u256.NewUint(1500),
				isToken0:      false,
				expectedError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				token0 := pool.balances.token0
				token1 := pool.balances.token1

				err := validatePoolBalance(token0, token1, tt.amount, tt.isToken0)
				if err != nil {
					if !tt.expectedError {
						t.Errorf("unexpected error: %v", err)
					}
				}
			})
		}
	})
}

func TestTransferFromAndVerify(t *testing.T) {
	tests := []struct {
		name         string
		pool         *Pool
		from         std.Address
		to           std.Address
		tokenPath    string
		amount       *i256.Int
		isToken0     bool
		expectedBal0 *u256.Uint
		expectedBal1 *u256.Uint
	}{
		{
			name: "normal token0 transfer",
			pool: &Pool{
				balances: Balances{
					token0: u256.NewUint(1000),
					token1: u256.NewUint(2000),
				},
			},
			from:         testutils.TestAddress("from_addr"),
			to:           testutils.TestAddress("to_addr"),
			tokenPath:    fooPath,
			amount:       i256.NewInt(500),
			isToken0:     true,
			expectedBal0: u256.NewUint(1500), // 1000 + 500
			expectedBal1: u256.NewUint(2000), // unchanged
		},
		{
			name: "normal token1 transfer",
			pool: &Pool{
				balances: Balances{
					token0: u256.NewUint(1000),
					token1: u256.NewUint(2000),
				},
			},
			from:         testutils.TestAddress("from_addr"),
			to:           testutils.TestAddress("to_addr"),
			tokenPath:    fooPath,
			amount:       i256.NewInt(800),
			isToken0:     false,
			expectedBal0: u256.NewUint(1000), // unchanged
			expectedBal1: u256.NewUint(2800), // 2000 + 800
		},
		{
			name: "zero value transfer",
			pool: &Pool{
				balances: Balances{
					token0: u256.NewUint(1000),
					token1: u256.NewUint(2000),
				},
			},
			from:         testutils.TestAddress("from_addr"),
			to:           testutils.TestAddress("to_addr"),
			tokenPath:    fooPath,
			amount:       i256.NewInt(0),
			isToken0:     true,
			expectedBal0: u256.NewUint(1000), // unchanged
			expectedBal1: u256.NewUint(2000), // unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TokenFaucet(t, tt.tokenPath, tt.from)
			TokenApprove(t, tt.tokenPath, tt.from, pool, u256.MustFromDecimal(tt.amount.ToString()).Uint64())

			tt.pool.safeTransferFrom(tt.from, tt.to, tt.tokenPath, u256.MustFromDecimal(tt.amount.ToString()), tt.isToken0)

			if !tt.pool.balances.token0.Eq(tt.expectedBal0) {
				t.Errorf("token0 balance mismatch: expected %s, got %s",
					tt.expectedBal0.ToString(),
					tt.pool.balances.token0.ToString())
			}

			if !tt.pool.balances.token1.Eq(tt.expectedBal1) {
				t.Errorf("token1 balance mismatch: expected %s, got %s",
					tt.expectedBal1.ToString(),
					tt.pool.balances.token1.ToString())
			}
		})
	}

	t.Run("negative value handling", func(t *testing.T) {
		pool := &Pool{
			balances: Balances{
				token0: u256.NewUint(1000),
				token1: u256.NewUint(2000),
			},
		}

		negativeAmount := i256.NewInt(-500)

		TokenFaucet(t, fooPath, testutils.TestAddress("from_addr"))
		TokenApprove(t, fooPath, testutils.TestAddress("from_addr"), consts.POOL_ADDR, u256.MustFromDecimal(negativeAmount.Abs().ToString()).Uint64())
		pool.safeTransferFrom(
			testutils.TestAddress("from_addr"),
			testutils.TestAddress("to_addr"),
			fooPath,
			u256.MustFromDecimal(negativeAmount.Abs().ToString()),
			true,
		)

		expectedBal := u256.NewUint(1500) // 1000 + 500 (absolute value)
		if !pool.balances.token0.Eq(expectedBal) {
			t.Errorf("negative amount handling failed: expected %s, got %s",
				expectedBal.ToString(),
				pool.balances.token0.ToString())
		}
	})

	t.Run("uint64 overflow value", func(t *testing.T) {
		pool := &Pool{
			balances: Balances{
				token0: u256.NewUint(1000),
				token1: u256.NewUint(2000),
			},
		}

		hugeAmount := i256.FromUint256(u256.MustFromDecimal("18446744073709551616")) // 2^64

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for amount exceeding uint64 range")
			}
		}()

		pool.safeTransferFrom(
			testutils.TestAddress("from_addr"),
			testutils.TestAddress("to_addr"),
			fooPath,
			u256.MustFromDecimal(hugeAmount.ToString()),
			true,
		)
	})
}

func TestUpdatePoolBalance(t *testing.T) {
	tests := []struct {
		name          string
		initialToken0 *u256.Uint
		initialToken1 *u256.Uint
		amount        *u256.Uint
		isToken0      bool
		expectedBal   *u256.Uint
		expectErr     bool
	}{
		{
			name:          "normal token0 decrease",
			initialToken0: u256.NewUint(1000),
			initialToken1: u256.NewUint(2000),
			amount:        u256.NewUint(300),
			isToken0:      true,
			expectedBal:   u256.NewUint(700),
			expectErr:     false,
		},
		{
			name:          "normal token1 decrease",
			initialToken0: u256.NewUint(1000),
			initialToken1: u256.NewUint(2000),
			amount:        u256.NewUint(500),
			isToken0:      false,
			expectedBal:   u256.NewUint(1500),
			expectErr:     false,
		},
		{
			name:          "insufficient token0 balance",
			initialToken0: u256.NewUint(100),
			initialToken1: u256.NewUint(2000),
			amount:        u256.NewUint(200),
			isToken0:      true,
			expectedBal:   nil,
			expectErr:     true,
		},
		{
			name:          "insufficient token1 balance",
			initialToken0: u256.NewUint(1000),
			initialToken1: u256.NewUint(100),
			amount:        u256.NewUint(200),
			isToken0:      false,
			expectedBal:   nil,
			expectErr:     true,
		},
		{
			name:          "zero value handling",
			initialToken0: u256.NewUint(1000),
			initialToken1: u256.NewUint(2000),
			amount:        u256.NewUint(0),
			isToken0:      true,
			expectedBal:   u256.NewUint(1000),
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				balances: Balances{
					token0: tt.initialToken0,
					token1: tt.initialToken1,
				},
			}

			newBal, err := updatePoolBalance(tt.initialToken0, tt.initialToken1, tt.amount, tt.isToken0)

			if tt.expectErr {
				if err == nil {
					t.Errorf("%s: expected error but no error", tt.name)
				}
				return
			}
			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.name, err)
				return
			}

			if !newBal.Eq(tt.expectedBal) {
				t.Errorf("%s: balance mismatch, expected: %s, actual: %s",
					tt.name,
					tt.expectedBal.ToString(),
					newBal.ToString(),
				)
			}
		})
	}
}
