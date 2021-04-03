package keeper

import (
	"fmt"

	"github.com/tendermint/classic/libs/log"
	"github.com/tendermint/go-amino-x"

	sdk "github.com/tendermint/classic/sdk/types"
	"github.com/tendermint/classic/sdk/x/supply/exported"
	"github.com/tendermint/classic/sdk/x/supply/internal/types"
)

// Keeper of the supply store
type Keeper struct {
	storeKey  sdk.StoreKey
	ak        types.AccountKeeper
	bk        types.BankKeeper
	permAddrs map[string]types.PermissionsForAddress
}

// NewKeeper creates a new Keeper instance
func NewKeeper(key sdk.StoreKey, ak types.AccountKeeper, bk types.BankKeeper, maccPerms map[string][]string) Keeper {
	// set the addresses
	permAddrs := make(map[string]types.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = types.NewPermissionsForAddress(name, perms)
	}

	return Keeper{
		storeKey:  key,
		ak:        ak,
		bk:        bk,
		permAddrs: permAddrs,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetSupply retrieves the Supply from store
func (k Keeper) GetSupply(ctx sdk.Context) (supply exported.SupplyI) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(SupplyKey)
	if b == nil {
		panic("stored supply should not have been nil")
	}
	amino.MustUnmarshalSized(b, &supply)
	return
}

// SetSupply sets the Supply to store
func (k Keeper) SetSupply(ctx sdk.Context, supply exported.SupplyI) {
	store := ctx.KVStore(k.storeKey)
	b := amino.MustMarshalSized(supply)
	store.Set(SupplyKey, b)
}

// ValidatePermissions validates that the module account has been granted
// permissions within its set of allowed permissions.
func (k Keeper) ValidatePermissions(macc exported.ModuleAccountI) error {
	permAddr := k.permAddrs[macc.GetName()]
	for _, perm := range macc.GetPermissions() {
		if !permAddr.HasPermission(perm) {
			return fmt.Errorf("invalid module permission %s", perm)
		}
	}
	return nil
}
