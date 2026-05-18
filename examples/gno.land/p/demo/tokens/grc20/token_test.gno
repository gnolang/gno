package grc20

import (
	"math"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/nt/urequire/v0"
)

// newTestToken constructs a Token with a fresh crossing-frame cur so
// NewToken's IsCurrent check + rlm.PkgPath() derivation works in p/ test
// scope. The closure's package path becomes the Token's origRealm.
func newTestToken(name, symbol string, decimals int) (tok *Token, adm *PrivateLedger) {
	func(cur realm) {
		tok, adm = NewToken(0, cur, name, symbol, decimals)
	}(cross)
	return
}

func TestTestImpl(t *testing.T) {
	bank, _ := newTestToken("Dummy", "DUMMY", 4)
	urequire.False(t, bank == nil, "dummy should not be nil")
}

func TestNewTokenValidation(t *testing.T) {
	// Each case wraps NewToken in a crossing closure (same as newTestToken)
	// to satisfy IsCurrent + non-empty pkgPath. The validation under test
	// is the name/symbol/decimals checks that run after.
	mustPanic := func(name, sym string, dec int, want error, label string) {
		t.Helper()
		// newTestToken wraps NewToken in cross(...), so the panic from
		// NewToken's validators crosses a realm boundary — use revive()
		// (defer-recover doesn't see cross-realm panics).
		r := revive(func() {
			newTestToken(name, sym, dec)
		})
		if r == nil {
			t.Errorf("%s: expected panic, got none", label)
			return
		}
		if r != want {
			t.Errorf("%s: expected %v, got %v", label, want, r)
		}
	}

	// Empty name / symbol.
	mustPanic("", "OK", 4, ErrInvalidName, "empty name")
	mustPanic("Name", "", 4, ErrInvalidSymbol, "empty symbol")

	// Length caps.
	mustPanic(strings.Repeat("a", MaxNameLen+1), "OK", 4, ErrInvalidName, "name too long")
	mustPanic("Name", strings.Repeat("A", MaxSymbolLen+1), 4, ErrInvalidSymbol, "symbol too long")

	// Name control characters.
	mustPanic("bad\x01name", "OK", 4, ErrInvalidName, "name with control char")
	mustPanic("bad\nname", "OK", 4, ErrInvalidName, "name with newline")

	// Symbol charset — disallowed delimiters and whitespace.
	mustPanic("Name", "BA.D", 4, ErrInvalidSymbol, "symbol with dot")
	mustPanic("Name", "BA/D", 4, ErrInvalidSymbol, "symbol with slash")
	mustPanic("Name", "BA D", 4, ErrInvalidSymbol, "symbol with space")
	mustPanic("Name", "BA\"D", 4, ErrInvalidSymbol, "symbol with quote")

	// Decimals out of range.
	mustPanic("Name", "OK", -1, ErrInvalidDecimals, "negative decimals")
	mustPanic("Name", "OK", MaxDecimals+1, ErrInvalidDecimals, "decimals over cap")

	// Boundary positives — should NOT panic.
	tok, _ := newTestToken(strings.Repeat("a", MaxNameLen), strings.Repeat("A", MaxSymbolLen), MaxDecimals)
	urequire.True(t, tok != nil, "boundary name+symbol+decimals should succeed")

	// UTF-8 name with non-ASCII is allowed.
	tok2, _ := newTestToken("Доллар", "RUB", 2)
	urequire.True(t, tok2 != nil, "UTF-8 name should be allowed")
}

func TestToken(t *testing.T) {
	var (
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
		carl  = testutils.TestAddress("carl")
	)

	bank, adm := newTestToken("Dummy", "DUMMY", 6)

	checkBalances := func(aliceEB, bobEB, carlEB int64) {
		t.Helper()
		exp := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceEB, bobEB, carlEB)
		aliceGB := bank.BalanceOf(alice)
		bobGB := bank.BalanceOf(bob)
		carlGB := bank.BalanceOf(carl)
		got := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceGB, bobGB, carlGB)
		uassert.Equal(t, got, exp, "invalid balances")
	}
	checkAllowances := func(abEB, acEB, baEB, bcEB, caEB, cbEB int64) {
		t.Helper()
		exp := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abEB, acEB, baEB, bcEB, caEB, cbEB)
		abGB := bank.Allowance(alice, bob)
		acGB := bank.Allowance(alice, carl)
		baGB := bank.Allowance(bob, alice)
		bcGB := bank.Allowance(bob, carl)
		caGB := bank.Allowance(carl, alice)
		cbGB := bank.Allowance(carl, bob)
		got := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abGB, acGB, baGB, bcGB, caGB, cbGB)
		uassert.Equal(t, got, exp, "invalid allowances")
	}

	checkBalances(0, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.Mint(alice, 1000))
	urequire.NoError(t, adm.Mint(alice, 100))
	checkBalances(1100, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.Approve(alice, bob, 99999999))
	checkBalances(1100, 0, 0)
	checkAllowances(99999999, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.Approve(alice, bob, 400))
	checkBalances(1100, 0, 0)
	checkAllowances(400, 0, 0, 0, 0, 0)

	urequire.Error(t, adm.TransferFrom(alice, bob, carl, 100000000))
	checkBalances(1100, 0, 0)
	checkAllowances(400, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.TransferFrom(alice, bob, carl, 100))
	checkBalances(1000, 0, 100)
	checkAllowances(300, 0, 0, 0, 0, 0)

	urequire.Error(t, adm.SpendAllowance(alice, bob, 2000000))
	checkBalances(1000, 0, 100)
	checkAllowances(300, 0, 0, 0, 0, 0)

	urequire.NoError(t, adm.SpendAllowance(alice, bob, 100))
	checkBalances(1000, 0, 100)
	checkAllowances(200, 0, 0, 0, 0, 0)
}

func TestMintOverflow(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, adm := newTestToken("Dummy", "DUMMY", 6)

	safeValue := int64(1 << 62)
	urequire.NoError(t, adm.Mint(alice, safeValue))
	urequire.Equal(t, tok.BalanceOf(alice), safeValue)

	err := adm.Mint(bob, safeValue)
	uassert.Error(t, err, "expected ErrMintOverflow")
}

func TestTransferFromAtomicity(t *testing.T) {
	var (
		owner   = testutils.TestAddress("owner")
		spender = testutils.TestAddress("spender")

		invalidRecipient = address("")
		recipient        = testutils.TestAddress("to")
	)

	token, admin := newTestToken("Test", "TEST", 6)

	// owner has 100 tokens, spender has 50 allowance
	initialBalance := int64(100)
	initialAllowance := int64(50)

	urequire.NoError(t, admin.Mint(owner, initialBalance))
	urequire.NoError(t, admin.Approve(owner, spender, initialAllowance))

	// transfer to an invalid address to force a transfer failure
	transferAmount := int64(30)
	err := admin.TransferFrom(owner, spender, invalidRecipient, transferAmount)
	uassert.Error(t, err, "transfer should fail due to invalid address")

	ownerBalance := token.BalanceOf(owner)
	uassert.Equal(t, ownerBalance, initialBalance, "owner balance should remain unchanged")

	// check if allowance was incorrectly reduced
	remainingAllowance := token.Allowance(owner, spender)
	uassert.Equal(t, remainingAllowance, initialAllowance,
		"allowance should not be reduced when transfer fails")

	// transfer all tokens
	admin.Transfer(owner, recipient, 100)
	remainingBalance := token.BalanceOf(owner)
	uassert.Equal(t, remainingBalance, int64(0),
		"balance should be zero")

	err = admin.TransferFrom(owner, spender, recipient, transferAmount)
	uassert.Error(t, err, "transfer should fail due to insufficient balance")
}

func TestMintUntilOverflow(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, adm := newTestToken("Dummy", "DUMMY", 6)

	tests := []struct {
		name           string
		addr           address
		amount         int64
		expectedError  error
		expectedSupply int64
		description    string
	}{
		{
			name:           "mint negative value",
			addr:           alice,
			amount:         -1,
			expectedError:  ErrInvalidAmount,
			expectedSupply: 0,
			description:    "minting a negative number should fail with ErrInvalidAmount",
		},
		{
			name:           "mint MaxInt64",
			addr:           alice,
			amount:         math.MaxInt64 - 1000,
			expectedError:  nil,
			expectedSupply: math.MaxInt64 - 1000,
			description:    "minting almost MaxInt64 should succeed",
		},
		{
			name:           "mint small value",
			addr:           bob,
			amount:         1000,
			expectedError:  nil,
			expectedSupply: math.MaxInt64,
			description:    "minting a small value when close to MaxInt64 should succeed",
		},
		{
			name:           "mint value that would exceed MaxInt64",
			addr:           bob,
			amount:         1,
			expectedError:  ErrMintOverflow,
			expectedSupply: math.MaxInt64,
			description:    "minting any value when at MaxInt64 should fail with ErrMintOverflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adm.Mint(tt.addr, tt.amount)

			if tt.expectedError != nil {
				uassert.Error(t, err, tt.description)
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				uassert.NoError(t, err, tt.description)
			}

			totalSupply := tok.TotalSupply()
			uassert.Equal(t, totalSupply, tt.expectedSupply, "totalSupply should match expected value")
		})
	}
}
