package grc20

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestTestImpl(t *testing.T) {
	bank, _ := NewToken("Dummy", "DUMMY", 4)
	urequire.False(t, bank == nil, "dummy should not be nil")
}

func TestToken(t *testing.T) {
	var (
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
		carl  = testutils.TestAddress("carl")
	)

	bank, adm := NewToken("Dummy", "DUMMY", 6)

	checkBalances := func(aliceEB, bobEB, carlEB uint64) {
		t.Helper()
		exp := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceEB, bobEB, carlEB)
		aliceGB := bank.BalanceOf(alice)
		bobGB := bank.BalanceOf(bob)
		carlGB := bank.BalanceOf(carl)
		got := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceGB, bobGB, carlGB)
		uassert.Equal(t, got, exp, "invalid balances")
	}
	checkAllowances := func(abEB, acEB, baEB, bcEB, caEB, cbEB uint64) {
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

func TestOverflow(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, adm := NewToken("Dummy", "DUMMY", 6)

	safeValue := uint64(1 << 62)
	urequire.NoError(t, adm.Mint(alice, safeValue))
	urequire.Equal(t, tok.BalanceOf(alice), safeValue)

	err := adm.Mint(bob, safeValue)
	uassert.Error(t, err, "expected ErrOverflow")
}

func TestTransferFromAtomicity(t *testing.T) {
	var (
		owner     = testutils.TestAddress("owner")
		spender   = testutils.TestAddress("spender")
		recipient = testutils.TestAddress("recipient")

		invalidRecipient = std.Address("")
	)

	token, admin := NewToken("Test", "TEST", 6)

	// owner has 100 tokens, spender has 50 allowance
	initialBalance := uint64(100)
	initialAllowance := uint64(50)

	urequire.NoError(t, admin.Mint(owner, initialBalance))
	urequire.NoError(t, admin.Approve(owner, spender, initialAllowance))

	// transfer to an invalid address to force a transfer failure
	transferAmount := uint64(30)
	err := admin.TransferFrom(owner, spender, invalidRecipient, transferAmount)
	uassert.Error(t, err, "transfer should fail due to invalid address")

	ownerBalance := token.BalanceOf(owner)
	uassert.Equal(t, ownerBalance, initialBalance, "owner balance should remain unchanged")

	// check if allowance was incorrectly reduced
	remainingAllowance := token.Allowance(owner, spender)
	uassert.Equal(t, remainingAllowance, initialAllowance,
		"allowance should not be reduced when transfer fails")
}

func TestMintOverflow(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tok, adm := NewToken("Dummy", "DUMMY", 6)

	tests := []struct {
		name           string
		address        std.Address
		amount         uint64
		expectedError  error
		expectedSupply uint64
		description    string
	}{
		{
			name:           "mint value larger than MaxUint64",
			address:        alice,
			amount:         MaxUint64 + 1000,
			expectedError:  ErrOverflow,
			expectedSupply: 0,
			description:    "minting a value larger than MaxUint64 should fail with ErrOverflow",
		},
		{
			name:           "mint safe value close to MaxUint64",
			address:        alice,
			amount:         MaxUint64 - 1000,
			expectedError:  nil,
			expectedSupply: MaxUint64 - 1000,
			description:    "minting a value close to MaxUint64 should succeed",
		},
		{
			name:           "mint small value",
			address:        bob,
			amount:         1000,
			expectedError:  nil,
			expectedSupply: MaxUint64,
			description:    "minting a small value when close to MaxUint64 should succeed",
		},
		{
			name:           "mint value that would exceed MaxUint64",
			address:        bob,
			amount:         1,
			expectedError:  ErrOverflow,
			expectedSupply: MaxUint64,
			description:    "minting any value when at MaxUint64 should fail with ErrOverflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adm.Mint(tt.address, tt.amount)

			if tt.expectedError != nil {
				uassert.Error(t, err, tt.description)
				if err.Error() != tt.expectedError.Error() {
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
