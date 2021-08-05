package params

import (
	"fmt"

	"github.com/tendermint/go-amino-x"

	sdk "github.com/tendermint/classic/sdk/types"
	"github.com/tendermint/classic/sdk/x/params/subspace"
	"github.com/tendermint/classic/sdk/x/params/types"

	"github.com/tendermint/classic/libs/log"
)

// Keeper of the global paramstore
type Keeper struct {
	key       sdk.StoreKey
	tkey      sdk.StoreKey
	codespace sdk.CodespaceType
	spaces    map[string]*Subspace
}

// NewKeeper constructs a params keeper
func NewKeeper(key *sdk.KVStoreKey, tkey *sdk.TransientStoreKey, codespace sdk.CodespaceType) (k Keeper) {
	k = Keeper{
		key:       key,
		tkey:      tkey,
		codespace: codespace,
		spaces:    make(map[string]*Subspace),
	}

	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Allocate subspace used for keepers
func (k Keeper) Subspace(s string) Subspace {
	_, ok := k.spaces[s]
	if ok {
		panic("subspace already occupied")
	}

	if s == "" {
		panic("cannot use empty string for subspace")
	}

	space := subspace.NewSubspace(k.key, k.tkey, s)
	k.spaces[s] = &space

	return space
}

// Get existing substore from keeper
func (k Keeper) GetSubspace(s string) (Subspace, bool) {
	space, ok := k.spaces[s]
	if !ok {
		return Subspace{}, false
	}
	return *space, ok
}
