package auth

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Concrete implementation of AccountKeeper.
type AccountKeeper struct {
	// The (unexposed) key used to access the store from the Context.
	key store.StoreKey

	// The prototypical Account constructor.
	proto func() std.Account
}

// NewAccountKeeper returns a new AccountKeeper that uses go-amino to
// (binary) encode and decode concrete std.Accounts.
func NewAccountKeeper(
	key store.StoreKey, proto func() std.Account,
) AccountKeeper {
	return AccountKeeper{
		key:   key,
		proto: proto,
	}
}

// Logger returns a module-specific logger.
func (ak AccountKeeper) Logger(ctx sdk.Context) *slog.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("auth"))
}

// NewAccountWithAddress implements AccountKeeper.
func (ak AccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr crypto.Address) std.Account {
	acc := ak.proto()
	// acc.SetSequence(0) // start with 0.
	err := acc.SetAddress(addr)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	err = acc.SetAccountNumber(ak.GetNextAccountNumber(ctx))
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	return acc
}

// GetAccount implements AccountKeeper.
func (ak AccountKeeper) GetAccount(ctx sdk.Context, addr crypto.Address) std.Account {
	stor := ctx.Store(ak.key)
	bz := stor.Get(AddressStoreKey(addr))
	if bz == nil {
		return nil
	}
	acc := ak.decodeAccount(bz)
	return acc
}

// GetAllAccounts returns all accounts in the AccountKeeper.
func (ak AccountKeeper) GetAllAccounts(ctx sdk.Context) []std.Account {
	accounts := []std.Account{}
	appendAccount := func(acc std.Account) (stop bool) {
		accounts = append(accounts, acc)
		return false
	}
	ak.IterateAccounts(ctx, appendAccount)
	return accounts
}

// SetAccount implements AccountKeeper.
func (ak AccountKeeper) SetAccount(ctx sdk.Context, acc std.Account) {
	addr := acc.GetAddress()
	stor := ctx.Store(ak.key)
	bz, err := amino.MarshalAny(acc)
	if err != nil {
		panic(err)
	}
	stor.Set(AddressStoreKey(addr), bz)
}

// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (ak AccountKeeper) RemoveAccount(ctx sdk.Context, acc std.Account) {
	addr := acc.GetAddress()
	stor := ctx.Store(ak.key)
	stor.Delete(AddressStoreKey(addr))
}

// IterateAccounts implements AccountKeeper.
func (ak AccountKeeper) IterateAccounts(ctx sdk.Context, process func(std.Account) (stop bool)) {
	stor := ctx.Store(ak.key)
	iter := store.PrefixIterator(stor, []byte(AddressStoreKeyPrefix))
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		val := iter.Value()
		acc := ak.decodeAccount(val)
		if process(acc) {
			return
		}
		iter.Next()
	}
}

// GetPubKey Returns the PubKey of the account at address
func (ak AccountKeeper) GetPubKey(ctx sdk.Context, addr crypto.Address) (crypto.PubKey, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", addr))
	}
	return acc.GetPubKey(), nil
}

// GetSequence Returns the Sequence of the account at address
func (ak AccountKeeper) GetSequence(ctx sdk.Context, addr crypto.Address) (uint64, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return 0, std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", addr))
	}
	return acc.GetSequence(), nil
}

// GetNextAccountNumber Returns and increments the global account number counter
func (ak AccountKeeper) GetNextAccountNumber(ctx sdk.Context) uint64 {
	var accNumber uint64
	stor := ctx.Store(ak.key)
	bz := stor.Get([]byte(GlobalAccountNumberKey))
	if bz == nil {
		accNumber = 0 // start with 0.
	} else {
		err := amino.Unmarshal(bz, &accNumber)
		if err != nil {
			panic(err)
		}
	}

	bz = amino.MustMarshal(accNumber + 1)
	stor.Set([]byte(GlobalAccountNumberKey), bz)

	return accNumber
}

// -----------------------------------------------------------------------------
// Misc.

func (ak AccountKeeper) decodeAccount(bz []byte) (acc std.Account) {
	err := amino.Unmarshal(bz, &acc)
	if err != nil {
		panic(err)
	}
	return
}
