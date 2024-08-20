package params

import (
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/store"
)

const (
	ModuleName = "params"

	StoreKey = "params"

	// ValueStorePrevfix is "/pv/" for param value.
	ValueStoreKeyPrefix = "/pv/"
)

func ValueStoreKey(key string) []byte {
	return append([]byte(ValueStoreKeyPrefix), []byte(key)...)
}

// Keeper of the global param store.
type Keeper struct {
	key       store.StoreKey
	keyMapper KeyMapper
}

// NewKeeper constructs a params keeper
func NewKeeper(key store.StoreKey, keyMapper KeyMapper) Keeper {
	return Keeper{
		key:       key,
		keyMapper: keyMapper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) *slog.Logger {
	return ctx.Logger().With("module", ModuleName)
}

// GetParam gets a param value from the global param store.
func (k Keeper) GetParam(ctx sdk.Context, key string, target interface{}) (bool, error) {
	stor := ctx.Store(k.key)
	if k.keyMapper != nil {
		key = k.keyMapper.Map(key)
	}

	bz := stor.Get(ValueStoreKey(key))
	if bz == nil {
		return false, nil
	}

	return true, amino.Unmarshal(bz, target)
}

// SetParam sets a param value to the global param store.
func (k Keeper) SetParam(ctx sdk.Context, key string, param interface{}) error {
	bz, err := amino.Marshal(param)
	if err != nil {
		return err
	}

	stor := ctx.Store(k.key)
	if k.keyMapper != nil {
		key = k.keyMapper.Map(key)
	}

	stor.Set(ValueStoreKey(key), bz)
	return nil
}
