package params

import (
	"fmt"
	"log/slog"
	"maps"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/store"
)

type ParamsKeeperI interface {
	Get(ctx sdk.Context, key string, ptr interface{})
	Set(ctx sdk.Context, key string, value interface{})
}

var _ ParamsKeeperI = ParamsKeeper{}

// global paramstore Keeper.
type ParamsKeeper struct {
	key    store.StoreKey
	table  KeyTable
	prefix string
}

// NewParamsKeeper returns a new ParamsKeeper.
func NewParamsKeeper(key store.StoreKey, prefix string) ParamsKeeper {
	return ParamsKeeper{
		key:    key,
		table:  NewKeyTable(),
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

func (pk ParamsKeeper) Get(ctx sdk.Context, key string, ptr interface{}) {
	pk.checkType(key, ptr)
	stor := ctx.Store(pk.key)
	bz := stor.Get([]byte(key))
	err := amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (pk ParamsKeeper) GetIfExists(ctx sdk.Context, key string, ptr interface{}) {
	stor := ctx.Store(pk.key)
	bz := stor.Get([]byte(key))
	if bz == nil {
		return
	}
	pk.checkType(key, ptr)
	err := amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (pk ParamsKeeper) GetRaw(ctx sdk.Context, key string) []byte {
	stor := ctx.Store(pk.key)
	return stor.Get([]byte(key))
}

func (pk ParamsKeeper) Set(ctx sdk.Context, key string, value interface{}) {
	pk.checkType(key, value)
	stor := ctx.Store(pk.key)
	bz, err := amino.MarshalJSON(value)
	if err != nil {
		panic(err)
	}
	stor.Set([]byte(key), bz)
}

func (pk ParamsKeeper) Update(ctx sdk.Context, key string, value []byte) error {
	attr, ok := pk.table.m[key]
	if !ok {
		panic(fmt.Sprintf("parameter %s not registered", key))
	}

	ty := attr.ty
	dest := reflect.New(ty).Interface()
	pk.GetIfExists(ctx, key, dest)

	if err := amino.UnmarshalJSON(value, dest); err != nil {
		return err
	}

	destValue := reflect.Indirect(reflect.ValueOf(dest)).Interface()
	if err := pk.Validate(ctx, key, destValue); err != nil {
		return err
	}

	pk.Set(ctx, key, dest)
	return nil
}

func (pk ParamsKeeper) Validate(ctx sdk.Context, key string, value interface{}) error {
	attr, ok := pk.table.m[key]
	if !ok {
		return fmt.Errorf("parameter %s not registered", key)
	}

	if err := attr.vfn(value); err != nil {
		return fmt.Errorf("invalid parameter value: %w", err)
	}

	return nil
}

func (pk ParamsKeeper) checkType(key string, value interface{}) {
	attr, ok := pk.table.m[key]
	if !ok {
		panic(fmt.Sprintf("parameter %s is not registered", key))
	}

	ty := attr.ty
	pty := reflect.TypeOf(value)
	if pty.Kind() == reflect.Ptr {
		pty = pty.Elem()
	}

	if pty != ty {
		panic("type mismatch with registered table")
	}
}

func (pk ParamsKeeper) HasKeyTable() bool {
	return len(pk.table.m) > 0
}

func (pk ParamsKeeper) WithKeyTable(table KeyTable) ParamsKeeper {
	if table.m == nil {
		panic("WithKeyTable() called with nil KeyTable")
	}
	if len(pk.table.m) != 0 {
		panic("WithKeyTable() called on already initialized Keeper")
	}

	maps.Copy(pk.table.m, table.m)
	return pk
}

// XXX: added, should we remove?
func (pk ParamsKeeper) RegisterType(psp ParamSetPair) {
	pk.table.RegisterType(psp)
}

// XXX: added, should we remove?
func (pk ParamsKeeper) HasTypeKey(key string) bool {
	return pk.table.HasKey(key)
}

// XXX: GetAllKeys
// XXX: GetAllParams
// XXX: ViewKeeper
// XXX: ModuleKeeper
