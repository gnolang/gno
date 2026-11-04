package grc20

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// newTestToken constructs a Token. The IIFE same-realm cross
// promotes the test's EOA-origin cur into a fresh /p/grc20
// CodeRealm cur so NewToken's IsCurrent + non-empty rlm.PkgPath()
// checks pass.
func newTestToken(name, symbol string, decimals int, id seqid.ID, rlm realm) (tok *Token, adm *PrivateLedger) {
	func(cur realm) {
		tok, adm = NewToken(name, symbol, decimals, id, cur)
	}(cross(rlm))
	return
}

func TestTestImpl(cur realm, t *testing.T) {
	bank, _ := newTestToken("Dummy", "DUMMY", 4, 0, cur)
	urequire.False(t, bank == nil, "dummy should not be nil")
}

func TestNewTokenAllowsDuplicateSymbolInSameRealm(cur realm, t *testing.T) {
	holder := testutils.TestAddress("holder")

	first, firstLedger := newTestToken("Same", "DUP", 6, 1, cur)
	second, secondLedger := newTestToken("Same", "DUP", 6, 2, cur)

	urequire.True(t, first.ID() != second.ID(), "duplicate symbols should not force duplicate IDs")

	urequire.NoError(t, firstLedger.Mint(holder, 11))
	urequire.NoError(t, secondLedger.Mint(holder, 22))
	require256(t, uint256(11), first.BalanceOf(holder), "first balance")
	require256(t, uint256(22), second.BalanceOf(holder), "second balance")
}

func TestNewTokenValidation(cur realm, t *testing.T) {
	// Wrap NewToken to satisfy IsCurrent and pkgPath requirements,
	// then validate name, symbol, and decimals.
	mustPanic := func(name, sym string, dec int, want error, label string) {
		t.Helper()
		// newTestToken wraps NewToken in cross(...), so the panic from
		// NewToken's validators crosses a realm boundary — use revive()
		// (defer-recover doesn't see cross-realm panics).
		r := revive(func() {
			newTestToken(name, sym, dec, 0, cur)
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
	tok, _ := newTestToken(strings.Repeat("a", MaxNameLen), strings.Repeat("A", MaxSymbolLen), MaxDecimals, 0, cur)
	urequire.True(t, tok != nil, "boundary name+symbol+decimals should succeed")

	// UTF-8 name with non-ASCII is allowed.
	tok2, _ := newTestToken("Доллар", "RUB", 2, 0, cur)
	urequire.True(t, tok2 != nil, "UTF-8 name should be allowed")
}

func TestToken(cur realm, t *testing.T) {
	var (
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
		carl  = testutils.TestAddress("carl")
	)

	bank, adm := newTestToken("Dummy", "DUMMY", 6, 0, cur)

	checkBalances := func(aliceEB, bobEB, carlEB uint256) {
		t.Helper()
		exp := fmtAmounts("alice", aliceEB, "bob", bobEB, "carl", carlEB)
		got := fmtAmounts("alice", bank.BalanceOf(alice), "bob", bank.BalanceOf(bob), "carl", bank.BalanceOf(carl))
		uassert.Equal(t, got, exp, "invalid balances")
	}
	checkAllowances := func(abEB, acEB, baEB, bcEB, caEB, cbEB uint256) {
		t.Helper()
		exp := fmtAmounts("ab", abEB, "ac", acEB, "ba", baEB, "bc", bcEB, "ca", caEB, "cb", cbEB)
		got := fmtAmounts(
			"ab", bank.Allowance(alice, bob), "ac", bank.Allowance(alice, carl),
			"ba", bank.Allowance(bob, alice), "bc", bank.Allowance(bob, carl),
			"ca", bank.Allowance(carl, alice), "cb", bank.Allowance(carl, bob),
		)
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
	tok, adm := newTestToken("Dummy", "DUMMY", 6, 0, cur)

	safeValue := uint256(1 << 255)
	urequire.NoError(t, adm.Mint(alice, safeValue))
	require256(t, uint256(safeValue), tok.BalanceOf(alice), "alice balance")

	err := adm.Mint(bob, safeValue)
	uassert.True(t, errors.Is(err, ErrMintOverflow), "expected ErrMintOverflow")
	require256(t, uint256(safeValue), tok.TotalSupply(), "supply unchanged by a refused mint")
	require256(t, uint256(0), tok.BalanceOf(bob), "bob unchanged by a refused mint")
}

func TestTransferFromAtomicity(cur realm, t *testing.T) {
	var (
		owner   = testutils.TestAddress("owner")
		spender = testutils.TestAddress("spender")

		invalidRecipient = address("")
		recipient        = testutils.TestAddress("to")
	)

	token, admin := newTestToken("Test", "TEST", 6, 0, cur)

	// owner has 100 tokens, spender has 50 allowance
	initialBalance := uint256(100)
	initialAllowance := uint256(50)

	urequire.NoError(t, admin.Mint(owner, initialBalance))
	urequire.NoError(t, admin.Approve(owner, spender, initialAllowance))

	// transfer to an invalid address to force a transfer failure
	transferAmount := uint256(30)
	err := admin.TransferFrom(owner, spender, invalidRecipient, transferAmount)
	uassert.Error(t, err, "transfer should fail due to invalid address")

	ownerBalance := token.BalanceOf(owner)
	require256(t, uint256(initialBalance), ownerBalance, "owner balance should remain unchanged")

	// check if allowance was incorrectly reduced
	remainingAllowance := token.Allowance(owner, spender)
	require256(t, uint256(initialAllowance), remainingAllowance, "allowance should not be reduced when transfer fails")

	// Self-transfers fail in Transfer after SpendAllowance. TransferFrom must
	// reject them before it can mutate the allowance.
	err = admin.TransferFrom(owner, spender, owner, transferAmount)
	uassert.True(t, errors.Is(err, ErrCannotTransferToSelf), "self-transfer should fail")
	require256(t, uint256(initialBalance), token.BalanceOf(owner), "owner balance should not change when self-transfer fails")
	require256(t, uint256(initialAllowance), token.Allowance(owner, spender), "allowance should not be reduced when self-transfer fails")

	// transfer all tokens
	admin.Transfer(owner, recipient, 100)
	remainingBalance := token.BalanceOf(owner)
	require256(t, uint256(0), remainingBalance, "balance should be zero")
	uassert.False(t, token.HasAddr(owner), "a zero balance is not stored")

	err = admin.TransferFrom(owner, spender, recipient, transferAmount)
	uassert.Error(t, err, "transfer should fail due to insufficient balance")
}

func TestMintUntilOverflow(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, adm := newTestToken("Dummy", "DUMMY", 6, 0, cur)

	tests := []struct {
		name           string
		addr           address
		amount         uint256
		expectedError  error
		expectedSupply uint256
		description    string
	}{
		{
			name:           "mint zero",
			addr:           alice,
			amount:         0,
			expectedError:  nil,
			expectedSupply: 0,
			description:    "minting zero is a no-op",
		},
		{
			name:           "mint almost MaxUint256",
			addr:           alice,
			amount:         math.MaxUint256 - 1000,
			expectedError:  nil,
			expectedSupply: math.MaxUint256 - 1000,
			description:    "minting almost MaxUint256 should succeed",
		},
		{
			name:           "mint value that would exceed MaxUint256",
			addr:           bob,
			amount:         1001,
			expectedError:  ErrMintOverflow,
			expectedSupply: math.MaxUint256 - 1000,
			description:    "minting past MaxUint256 should fail with ErrMintOverflow",
		},
		{
			name:           "mint small value",
			addr:           bob,
			amount:         1000,
			expectedError:  nil,
			expectedSupply: math.MaxUint256,
			description:    "minting up to exactly MaxUint256 should succeed",
		},
		{
			name:           "mint one past MaxUint256",
			addr:           bob,
			amount:         1,
			expectedError:  ErrMintOverflow,
			expectedSupply: math.MaxUint256,
			description:    "minting any value when at MaxUint256 should fail with ErrMintOverflow",
		},
		{
			name:           "mint MaxUint256 at MaxUint256",
			addr:           bob,
			amount:         math.MaxUint256,
			expectedError:  ErrMintOverflow,
			expectedSupply: math.MaxUint256,
			description:    "an amount that would wrap the supply to a small value must be refused",
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

			require256(t, uint256(tt.expectedSupply), tok.TotalSupply(), "totalSupply should match expected value")
		})
	}

	require256(t, uint256(math.MaxUint256-1000), tok.BalanceOf(alice), "alice balance")
	require256(t, uint256(1000), tok.BalanceOf(bob), "bob balance")
}

// TestMaxSupplyMovements moves the whole MaxUint256 supply around: every
// balance and the total stay exact at the top of the range.
func TestMaxSupplyMovements(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")
	tok, adm := newTestToken("Dummy", "DUMMY", 18, 0, cur)

	urequire.NoError(t, adm.Mint(alice, math.MaxUint256))
	require256(t, uint256(math.MaxUint256), tok.TotalSupply(), "supply")

	// Transfer everything but one.
	urequire.NoError(t, adm.Transfer(alice, bob, math.MaxUint256-1))
	require256(t, uint256(1), tok.BalanceOf(alice), "alice")
	require256(t, uint256(math.MaxUint256-1), tok.BalanceOf(bob), "bob")

	// Overdraft by one is refused and changes nothing.
	err := adm.Transfer(alice, bob, 2)
	uassert.True(t, errors.Is(err, ErrInsufficientBalance), "overdraft must be refused")
	require256(t, uint256(1), tok.BalanceOf(alice), "alice after refused overdraft")

	// The last unit lands on bob, who now holds MaxUint256.
	urequire.NoError(t, adm.Transfer(alice, bob, 1))
	require256(t, uint256(math.MaxUint256), tok.BalanceOf(bob), "bob holds everything")
	uassert.False(t, tok.HasAddr(alice), "alice's zero balance is removed")
	uassert.Equal(t, 1, tok.KnownAccounts())

	// A transfer above MaxUint256 - balance is refused rather than wrapped.
	err = adm.Transfer(bob, carl, math.MaxUint256)
	urequire.NoError(t, err)
	err = adm.Transfer(bob, carl, 1)
	uassert.True(t, errors.Is(err, ErrInsufficientBalance), "empty account must not underflow")

	// Burn everything.
	urequire.NoError(t, adm.Burn(carl, math.MaxUint256))
	require256(t, uint256(0), tok.TotalSupply(), "supply after burn")
	uassert.Equal(t, 0, tok.KnownAccounts())

	err = adm.Burn(carl, 1)
	uassert.True(t, errors.Is(err, ErrInsufficientBalance), "burning from zero must be refused")
}

// TestAbove2Pow63 moves amounts no int64 token could hold.
func TestAbove2Pow63(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")
	tok, adm := newTestToken("Dummy", "DUMMY", 18, 0, cur)

	big := uint256(1<<63) + 7 // 9223372036854775815
	urequire.NoError(t, adm.Mint(alice, big*3))
	require256(t, uint256(big*3), tok.TotalSupply(), "supply")

	urequire.NoError(t, adm.Transfer(alice, bob, big))
	require256(t, uint256(big*2), tok.BalanceOf(alice), "alice")
	require256(t, uint256(big), tok.BalanceOf(bob), "bob")

	urequire.NoError(t, adm.Approve(alice, carl, big))
	urequire.NoError(t, adm.TransferFrom(alice, carl, carl, big))
	require256(t, uint256(big), tok.BalanceOf(alice), "alice after TransferFrom")
	require256(t, uint256(big), tok.BalanceOf(carl), "carl")
	require256(t, uint256(0), tok.Allowance(alice, carl), "allowance used up")

	uassert.Equal(t, "9223372036854775815", strconv.FormatUint256(tok.BalanceOf(bob), 10))
}

// TestAllowanceNearMax exercises an "infinite" (MaxUint256) approval and
// spending against it.
func TestAllowanceNearMax(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")
	tok, adm := newTestToken("Dummy", "DUMMY", 18, 0, cur)

	urequire.NoError(t, adm.Mint(alice, 1000))
	urequire.NoError(t, adm.Approve(alice, bob, math.MaxUint256))
	require256(t, uint256(math.MaxUint256), tok.Allowance(alice, bob), "max allowance")

	// Spending decreases even a max allowance (no "infinite" special case).
	urequire.NoError(t, adm.TransferFrom(alice, bob, carl, 400))
	require256(t, uint256(math.MaxUint256-400), tok.Allowance(alice, bob), "allowance after spend")

	// Spending more than the allowance is refused without wrapping.
	urequire.NoError(t, adm.Approve(alice, bob, 10))
	err := adm.SpendAllowance(alice, bob, 11)
	uassert.True(t, errors.Is(err, ErrInsufficientAllowance), "over-allowance must be refused")
	require256(t, uint256(10), tok.Allowance(alice, bob), "allowance unchanged")

	// A max-size spend against a max allowance leaves zero, which is removed.
	urequire.NoError(t, adm.Approve(alice, bob, math.MaxUint256))
	urequire.NoError(t, adm.SpendAllowance(alice, bob, math.MaxUint256))
	require256(t, uint256(0), tok.Allowance(alice, bob), "allowance spent")

	// Approving zero clears an allowance.
	urequire.NoError(t, adm.Approve(alice, bob, 5))
	urequire.NoError(t, adm.Approve(alice, bob, 0))
	require256(t, uint256(0), tok.Allowance(alice, bob), "allowance cleared")

	// TransferFrom with a max allowance but insufficient balance is refused
	// before the allowance is touched.
	urequire.NoError(t, adm.Approve(alice, bob, math.MaxUint256))
	err = adm.TransferFrom(alice, bob, carl, 601)
	uassert.True(t, errors.Is(err, ErrInsufficientBalance), "insufficient balance")
	require256(t, uint256(math.MaxUint256), tok.Allowance(alice, bob), "allowance untouched")
}

func TestFormatAmount(t *testing.T) {
	cases := []struct {
		amount   uint256
		decimals int
		want     string
	}{
		{0, 0, "0"},
		{0, 4, "0.0000"},
		{12, 4, "0.0012"},
		{12345, 4, "1.2345"},
		{10000, 4, "1.0000"},
		{math.MaxUint256, 0, "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
		{math.MaxUint256, 18, "115792089237316195423570985008687907853269984665640564039457.584007913129639935"},
	}
	for _, c := range cases {
		uassert.Equal(t, c.want, FormatAmount(c.amount, c.decimals))
	}
}

func TestRenderHomeMaxSupply(cur realm, t *testing.T) {
	tok, adm := newTestToken("Dummy", "DUMMY", 6, 0, cur)
	urequire.NoError(t, adm.Mint(testutils.TestAddress("alice"), math.MaxUint256))
	exp := "# Dummy ($DUMMY)\n\n" +
		"* **Decimals**: 6\n" +
		"* **Total supply**: 115792089237316195423570985008687907853269984665640564039457584007913129639935\n" +
		"* **Known accounts**: 1\n"
	uassert.Equal(t, exp, tok.RenderHome())
}

// A token name is display text and may contain markdown. RenderHome must not
// turn it into a live link (or any other markup) on the token's page.
func TestRenderHomeEscapesName(cur realm, t *testing.T) {
	const evil = "[Claim reward](https://evil.example)"
	tok, _ := newTestToken(evil, "EVIL", 6, 0, cur)
	out := tok.RenderHome()
	uassert.False(t, strings.Contains(out, evil), "name rendered as a live link: "+out)
	uassert.True(t, strings.Contains(out, `\[Claim reward\]\(https://evil\.example\)`), "name not escaped: "+out)
}
