package grc20

import (
	"errors"
	"math"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/nt/urequire/v0"
)

// newTestToken constructs a Token. The IIFE same-realm cross
// promotes the test's EOA-origin cur into a fresh /p/grc20
// CodeRealm cur so NewToken's IsCurrent + non-empty rlm.PkgPath()
// checks pass.
func newTestToken(name, symbol string, decimals int, rlm realm) (tok *Token, adm *PrivateLedger) {
	func(cur realm) {
		tok, adm = NewToken(name, symbol, decimals, cur)
	}(cross(rlm))
	return
}

func TestTestImpl(cur realm, t *testing.T) {
	bank, _ := newTestToken("Dummy", "DUMMY", 4, cur)
	urequire.False(t, bank == nil, "dummy should not be nil")
}

// Two tokens may share a symbol; they still keep separate ledgers. Their IDs
// differ too, but only once they have been persisted — the VM stamps an object
// ID at finalization, so both read "" here. See
// filetests/token_identity_filetest.gno.
func TestNewTokenAllowsDuplicateSymbolInSameRealm(cur realm, t *testing.T) {
	holder := testutils.TestAddress("holder")

	first, firstLedger := newTestToken("Same", "DUP", 6, cur)
	second, secondLedger := newTestToken("Same", "DUP", 6, cur)

	urequire.NoError(t, firstLedger.Mint(holder, 11))
	urequire.NoError(t, secondLedger.Mint(holder, 22))
	urequire.Equal(t, int64(11), first.BalanceOf(holder))
	urequire.Equal(t, int64(22), second.BalanceOf(holder))
}

func TestNewTokenValidation(cur realm, t *testing.T) {
	tests := []struct {
		label   string
		name    string
		symbol  string
		decimal int
		want    error
	}{
		{label: "empty name", name: "", symbol: "OK", decimal: 4, want: ErrInvalidName},
		{label: "empty symbol", name: "Name", symbol: "", decimal: 4, want: ErrInvalidSymbol},
		{label: "name too long", name: strings.Repeat("a", MaxNameLen+1), symbol: "OK", decimal: 4, want: ErrInvalidName},
		{label: "symbol too long", name: "Name", symbol: strings.Repeat("A", MaxSymbolLen+1), decimal: 4, want: ErrInvalidSymbol},
		{label: "name with control char", name: "bad\x01name", symbol: "OK", decimal: 4, want: ErrInvalidName},
		{label: "name with newline", name: "bad\nname", symbol: "OK", decimal: 4, want: ErrInvalidName},
		// Symbol charset — disallowed delimiters and whitespace. A symbol is
		// the trailing component of TokenPath(), which downstream splits on
		// `.`, so these must not get through.
		{label: "symbol with dot", name: "Name", symbol: "BA.D", decimal: 4, want: ErrInvalidSymbol},
		{label: "symbol with slash", name: "Name", symbol: "BA/D", decimal: 4, want: ErrInvalidSymbol},
		{label: "symbol with space", name: "Name", symbol: "BA D", decimal: 4, want: ErrInvalidSymbol},
		{label: "symbol with quote", name: "Name", symbol: "BA\"D", decimal: 4, want: ErrInvalidSymbol},
		{label: "negative decimals", name: "Name", symbol: "OK", decimal: -1, want: ErrInvalidDecimals},
		{label: "decimals over cap", name: "Name", symbol: "OK", decimal: MaxDecimals + 1, want: ErrInvalidDecimals},
		// Boundary positives — these must construct.
		{label: "boundary name+symbol+decimals", name: strings.Repeat("a", MaxNameLen), symbol: strings.Repeat("A", MaxSymbolLen), decimal: MaxDecimals},
		{label: "UTF-8 name", name: "Доллар", symbol: "RUB", decimal: 2},
	}

	for _, tt := range tests {
		// newTestToken wraps NewToken in cross(...), so a panic from
		// NewToken's validators crosses a realm boundary — use revive()
		// (defer-recover doesn't see cross-realm panics).
		var tok *Token
		r := revive(func() {
			tok, _ = newTestToken(tt.name, tt.symbol, tt.decimal, cur)
		})
		if tt.want == nil {
			if r != nil {
				t.Errorf("%s: expected no panic, got %v", tt.label, r)
				continue
			}
			uassert.True(t, tok != nil, tt.label+": should construct")
			continue
		}
		if r == nil {
			t.Errorf("%s: expected panic, got none", tt.label)
			continue
		}
		if r != tt.want {
			t.Errorf("%s: expected %v, got %v", tt.label, tt.want, r)
		}
	}
}

func TestToken(cur realm, t *testing.T) {
	var (
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
		carl  = testutils.TestAddress("carl")
	)

	bank, adm := newTestToken("Dummy", "DUMMY", 6, cur)

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

func TestMintOverflow(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, adm := newTestToken("Dummy", "DUMMY", 6, cur)

	safeValue := int64(1 << 62)
	urequire.NoError(t, adm.Mint(alice, safeValue))
	urequire.Equal(t, tok.BalanceOf(alice), safeValue)

	err := adm.Mint(bob, safeValue)
	uassert.Error(t, err, "expected ErrMintOverflow")
}

func TestTransferFromAtomicity(cur realm, t *testing.T) {
	var (
		owner   = testutils.TestAddress("owner")
		spender = testutils.TestAddress("spender")

		invalidRecipient = address("")
		recipient        = testutils.TestAddress("to")
	)

	token, admin := newTestToken("Test", "TEST", 6, cur)

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

	// Self-transfers fail in Transfer after SpendAllowance. TransferFrom must
	// reject them before it can mutate the allowance.
	err = admin.TransferFrom(owner, spender, owner, transferAmount)
	uassert.True(t, errors.Is(err, ErrCannotTransferToSelf), "self-transfer should fail")
	uassert.Equal(t, initialBalance, token.BalanceOf(owner),
		"owner balance should not change when self-transfer fails")
	uassert.Equal(t, initialAllowance, token.Allowance(owner, spender),
		"allowance should not be reduced when self-transfer fails")

	// transfer all tokens
	admin.Transfer(owner, recipient, 100)
	remainingBalance := token.BalanceOf(owner)
	uassert.Equal(t, remainingBalance, int64(0),
		"balance should be zero")

	err = admin.TransferFrom(owner, spender, recipient, transferAmount)
	uassert.Error(t, err, "transfer should fail due to insufficient balance")
}

func TestMintUntilOverflow(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, adm := newTestToken("Dummy", "DUMMY", 6, cur)

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
				if !errors.Is(err, tt.expectedError) {
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
