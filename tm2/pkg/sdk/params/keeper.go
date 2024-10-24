package params

import (
	"log/slog"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/store"
)

const (
	ModuleName = "params"
	StoreKey   = ModuleName
)

type ParamsKeeperI interface {
	GetString(ctx sdk.Context, key string, ptr *string)
	GetInt64(ctx sdk.Context, key string, ptr *int64)
	GetUint64(ctx sdk.Context, key string, ptr *uint64)
	GetBool(ctx sdk.Context, key string, ptr *bool)
	GetBytes(ctx sdk.Context, key string, ptr *[]byte)

	SetString(ctx sdk.Context, key string, value string)
	SetInt64(ctx sdk.Context, key string, value int64)
	SetUint64(ctx sdk.Context, key string, value uint64)
	SetBool(ctx sdk.Context, key string, value bool)
	SetBytes(ctx sdk.Context, key string, value []byte)

	Has(ctx sdk.Context, key string) bool
	GetRaw(ctx sdk.Context, key string) []byte

	// XXX: ListKeys?
}

var _ ParamsKeeperI = ParamsKeeper{}

// global paramstore Keeper.
type ParamsKeeper struct {
	key    store.StoreKey
	prefix string
}

// NewParamsKeeper returns a new ParamsKeeper.
func NewParamsKeeper(key store.StoreKey, prefix string) ParamsKeeper {
	return ParamsKeeper{
		key:    key,
		prefix: prefix,
	}
}

// Logger returns a module-specific logger.
// XXX: why do we expose this?
func (pk ParamsKeeper) Logger(ctx sdk.Context) *slog.Logger {
	return ctx.Logger().With("module", ModuleName)
}

func (pk ParamsKeeper) Has(ctx sdk.Context, key string) bool {
	stor := ctx.Store(pk.key)
	return stor.Has([]byte(key))
}

func (pk ParamsKeeper) GetRaw(ctx sdk.Context, key string) []byte {
	stor := ctx.Store(pk.key)
	return stor.Get([]byte(key))
}

func (pk ParamsKeeper) GetString(ctx sdk.Context, key string, ptr *string) {
	checkSuffix(key, ".string")
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetBool(ctx sdk.Context, key string, ptr *bool) {
	checkSuffix(key, ".bool")
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetInt64(ctx sdk.Context, key string, ptr *int64) {
	checkSuffix(key, ".int64")
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetUint64(ctx sdk.Context, key string, ptr *uint64) {
	checkSuffix(key, ".uint64")
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetBytes(ctx sdk.Context, key string, ptr *[]byte) {
	checkSuffix(key, ".bytes")
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) SetString(ctx sdk.Context, key, value string) {
	checkSuffix(key, ".string")
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetBool(ctx sdk.Context, key string, value bool) {
	checkSuffix(key, ".bool")
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetInt64(ctx sdk.Context, key string, value int64) {
	checkSuffix(key, ".int64")
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetUint64(ctx sdk.Context, key string, value uint64) {
	checkSuffix(key, ".uint64")
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetBytes(ctx sdk.Context, key string, value []byte) {
	checkSuffix(key, ".bytes")
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) getIfExists(ctx sdk.Context, key string, ptr interface{}) {
	stor := ctx.Store(pk.key)
	bz := stor.Get([]byte(key))
	if bz == nil {
		return
	}
	err := amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (pk ParamsKeeper) get(ctx sdk.Context, key string, ptr interface{}) {
	stor := ctx.Store(pk.key)
	bz := stor.Get([]byte(key))
	err := amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (pk ParamsKeeper) set(ctx sdk.Context, key string, value interface{}) {
	stor := ctx.Store(pk.key)
	bz, err := amino.MarshalJSON(value)
	if err != nil {
		panic(err)
	}
	stor.Set([]byte(key), bz)
}

func checkSuffix(key, expectedSuffix string) {
	var (
		noSuffix = !strings.HasSuffix(key, expectedSuffix)
		noName   = len(key) == len(expectedSuffix)
		// XXX: additional sanity checks?
	)
	if noSuffix || noName {
		panic(`key should be like "<name>` + expectedSuffix + `" (` + key + `)`)
	}
}
