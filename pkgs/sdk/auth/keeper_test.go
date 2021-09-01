package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/pkgs/crypto"
)

func TestAccountMapperGetSet(t *testing.T) {
	input := setupTestInput()
	addr := crypto.AddressFromPreimage([]byte("some-address"))

	// no account before its created
	acc := input.acck.GetAccount(input.ctx, addr)
	require.Nil(t, acc)

	// create account and check default values
	acc = input.acck.NewAccountWithAddress(input.ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, addr, acc.GetAddress())
	require.EqualValues(t, nil, acc.GetPubKey())
	require.EqualValues(t, 0, acc.GetSequence())

	// NewAccount doesn't call Set, so it's still nil
	require.Nil(t, input.acck.GetAccount(input.ctx, addr))

	// set some values on the account and save it
	newSequence := uint64(20)
	acc.SetSequence(newSequence)
	input.acck.SetAccount(input.ctx, acc)

	// check the new values
	acc = input.acck.GetAccount(input.ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, newSequence, acc.GetSequence())
}

func TestAccountMapperRemoveAccount(t *testing.T) {
	input := setupTestInput()
	addr1 := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))

	// create accounts
	acc1 := input.acck.NewAccountWithAddress(input.ctx, addr1)
	acc2 := input.acck.NewAccountWithAddress(input.ctx, addr2)

	accSeq1 := uint64(20)
	accSeq2 := uint64(40)

	acc1.SetSequence(accSeq1)
	acc2.SetSequence(accSeq2)
	input.acck.SetAccount(input.ctx, acc1)
	input.acck.SetAccount(input.ctx, acc2)

	acc1 = input.acck.GetAccount(input.ctx, addr1)
	require.NotNil(t, acc1)
	require.Equal(t, accSeq1, acc1.GetSequence())

	// remove one account
	input.acck.RemoveAccount(input.ctx, acc1)
	acc1 = input.acck.GetAccount(input.ctx, addr1)
	require.Nil(t, acc1)

	acc2 = input.acck.GetAccount(input.ctx, addr2)
	require.NotNil(t, acc2)
	require.Equal(t, accSeq2, acc2.GetSequence())
}
