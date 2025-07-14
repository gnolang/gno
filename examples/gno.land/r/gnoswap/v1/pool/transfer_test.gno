package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestValidatePoolBalance(t *testing.T) {
	tests := []struct {
		name          string
		token0        *u256.Uint
		token1        *u256.Uint
		amount        *u256.Uint
		isToken0      bool
		expectedError bool
	}{
		{
			name:          "successful token0 transfer",
			token0:        u256.NewUint(1000),
			token1:        u256.NewUint(1000),
			amount:        u256.NewUint(500),
			isToken0:      true,
			expectedError: false,
		},
		{
			name:          "successful token1 transfer",
			token0:        u256.NewUint(1000),
			token1:        u256.NewUint(1000),
			amount:        u256.NewUint(500),
			isToken0:      false,
			expectedError: false,
		},
		{
			name:          "insufficient token0 balance",
			token0:        u256.NewUint(100),
			token1:        u256.NewUint(1000),
			amount:        u256.NewUint(500),
			isToken0:      true,
			expectedError: true,
		},
		{
			name:          "insufficient token1 balance",
			token0:        u256.NewUint(1000),
			token1:        u256.NewUint(100),
			amount:        u256.NewUint(500),
			isToken0:      false,
			expectedError: true,
		},
		{
			name:          "nil token0",
			token0:        nil,
			token1:        u256.NewUint(1000),
			amount:        u256.NewUint(500),
			isToken0:      true,
			expectedError: true,
		},
		{
			name:          "nil token1",
			token0:        u256.NewUint(1000),
			token1:        nil,
			amount:        u256.NewUint(500),
			isToken0:      false,
			expectedError: true,
		},
		{
			name:          "nil amount",
			token0:        u256.NewUint(1000),
			token1:        u256.NewUint(1000),
			amount:        nil,
			isToken0:      true,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePoolBalance(tt.token0, tt.token1, tt.amount, tt.isToken0)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestUpdatePoolBalance(t *testing.T) {
	tests := []struct {
		name          string
		token0        *u256.Uint
		token1        *u256.Uint
		amount        *u256.Uint
		isToken0      bool
		expectedBal   *u256.Uint
		expectedError bool
	}{
		{
			name:          "successful token0 decrease",
			token0:        u256.NewUint(1000),
			token1:        u256.NewUint(1000),
			amount:        u256.NewUint(500),
			isToken0:      true,
			expectedBal:   u256.NewUint(500),
			expectedError: false,
		},
		{
			name:          "successful token1 decrease",
			token0:        u256.NewUint(1000),
			token1:        u256.NewUint(1000),
			amount:        u256.NewUint(500),
			isToken0:      false,
			expectedBal:   u256.NewUint(500),
			expectedError: false,
		},
		{
			name:          "insufficient token0 balance",
			token0:        u256.NewUint(100),
			token1:        u256.NewUint(1000),
			amount:        u256.NewUint(500),
			isToken0:      true,
			expectedBal:   nil,
			expectedError: true,
		},
		{
			name:          "insufficient token1 balance",
			token0:        u256.NewUint(1000),
			token1:        u256.NewUint(100),
			amount:        u256.NewUint(500),
			isToken0:      false,
			expectedBal:   nil,
			expectedError: true,
		},
		{
			name:          "zero amount",
			token0:        u256.NewUint(1000),
			token1:        u256.NewUint(1000),
			amount:        u256.NewUint(0),
			isToken0:      true,
			expectedBal:   u256.NewUint(1000),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newBal, err := updatePoolBalance(tt.token0, tt.token1, tt.amount, tt.isToken0)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !newBal.Eq(tt.expectedBal) {
				t.Errorf("balance mismatch: expected %s, got %s",
					tt.expectedBal.ToString(),
					newBal.ToString())
			}
		})
	}
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
			TokenApprove(t, tt.tokenPath, tt.from, poolAddr, u256.MustFromDecimal(tt.amount.ToString()).Int64())

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
		TokenApprove(t, fooPath, testutils.TestAddress("from_addr"), poolAddr, u256.MustFromDecimal(negativeAmount.Abs().ToString()).Int64())
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
