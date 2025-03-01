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

	// StoreKey = ModuleName
	StoreKeyPrefix = "/pv/"
)

func storeKey(key string) []byte {
	return append([]byte(StoreKeyPrefix), []byte(key)...)
}

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
	SetRaw(ctx sdk.Context, key string, value []byte)

	GetStruct(ctx sdk.Context, key string, paramPtr interface{})
	SetStruct(ctx sdk.Context, key string, param interface{})
}

type ParamfulKeeper interface {
	WillSetParam(ctx sdk.Context, key string, value interface{})
}

var _ ParamsKeeperI = ParamsKeeper{}

// global paramstore Keeper.
type ParamsKeeper struct {
	key  store.StoreKey
	kprs map[string]ParamfulKeeper // Register a prefix for module parameter keys.
}

// NewParamsKeeper returns a new ParamsKeeper.
func NewParamsKeeper(key store.StoreKey) ParamsKeeper {
	return ParamsKeeper{
		key:  key,
		kprs: map[string]ParamfulKeeper{},
	}
}

func (pk ParamsKeeper) ForModule(moduleName string) prefixParamsKeeper {
	ppk := newPrefixParamsKeeper(pk, moduleName+":")
	return ppk
}

func (pk ParamsKeeper) GetRegisteredKeeper(moduleName string) ParamfulKeeper {
	rk, ok := pk.kprs[moduleName]
	if !ok {
		panic("keeper for module " + moduleName + " not registered")
	}
	return rk
}

func (pk ParamsKeeper) Register(moduleName string, pmk ParamfulKeeper) {
	if _, exists := pk.kprs[moduleName]; exists {
		panic("keeper for module " + moduleName + " already registered")
	}
	pk.kprs[moduleName] = pmk
}

func (pk ParamsKeeper) IsRegistered(moduleName string) bool {
	_, ok := pk.kprs[moduleName]
	return ok
}

func (pk ParamsKeeper) ModuleExists(moduleName string) bool {
	return pk.IsRegistered(moduleName)
}

// XXX: why do we expose this?
func (pk ParamsKeeper) Logger(ctx sdk.Context) *slog.Logger {
	return ctx.Logger().With("module", ModuleName)
}

func (pk ParamsKeeper) Has(ctx sdk.Context, key string) bool {
	stor := ctx.Store(pk.key)
	return stor.Has(storeKey(key))
}

func (pk ParamsKeeper) GetString(ctx sdk.Context, key string, ptr *string) {
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetBool(ctx sdk.Context, key string, ptr *bool) {
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetInt64(ctx sdk.Context, key string, ptr *int64) {
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetUint64(ctx sdk.Context, key string, ptr *uint64) {
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetBytes(ctx sdk.Context, key string, ptr *[]byte) {
	pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) SetString(ctx sdk.Context, key, value string) {
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetBool(ctx sdk.Context, key string, value bool) {
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetInt64(ctx sdk.Context, key string, value int64) {
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetUint64(ctx sdk.Context, key string, value uint64) {
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetBytes(ctx sdk.Context, key string, value []byte) {
	pk.set(ctx, key, value)
}

func (pk ParamsKeeper) GetRaw(ctx sdk.Context, key string) []byte {
	stor := ctx.Store(pk.key)
	return stor.Get(storeKey(key))
}

func (pk ParamsKeeper) SetRaw(ctx sdk.Context, key string, value []byte) {
	stor := ctx.Store(pk.key)
	stor.Set(storeKey(key), value)
}

func (pk ParamsKeeper) GetStruct(ctx sdk.Context, key string, paramPtr interface{}) {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		panic("struct key expected format <module>:<struct name>")
	}
	moduleName := parts[0]
	structName := parts[1]
	if !pk.IsRegistered(moduleName) {
		panic("unregistered module name")
	}
	if structName != "_" {
		panic("the only supported struct name is _")
	}
	stor := ctx.Store(pk.key)
	kvz := getStructFieldsFromStore(paramPtr, stor, storeKey(key))
	decodeStructFields(paramPtr, kvz)
}

func (pk ParamsKeeper) SetStruct(ctx sdk.Context, key string, param interface{}) {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		panic("struct key expected format <module>:<struct name>")
	}
	moduleName := parts[0]
	structName := parts[1]
	if !pk.IsRegistered(moduleName) {
		panic("unregistered module name")
	}
	if structName != "_" {
		panic("the only supported struct name is _")
	}
	stor := ctx.Store(pk.key)
	kvz := encodeStructFields(param)
	for _, kv := range kvz {
		stor.Set(storeKey(key+":"+string(kv.Key)), kv.Value)
	}
}

func (pk ParamsKeeper) getIfExists(ctx sdk.Context, key string, ptr interface{}) {
	stor := ctx.Store(pk.key)
	bz := stor.Get(storeKey(key))
	if bz == nil {
		return
	}
	amino.MustUnmarshalJSON(bz, ptr)
}

func (pk ParamsKeeper) set(ctx sdk.Context, key string, value interface{}) {
	module, rawKey := parsePrefix(key)
	if module != "" {
		kpr := pk.GetRegisteredKeeper(module)
		if kpr != nil {
			kpr.WillSetParam(ctx, rawKey, value)
		}
	}
	stor := ctx.Store(pk.key)
	bz := amino.MustMarshalJSON(value)
	stor.Set(storeKey(key), bz)
}

func parsePrefix(key string) (prefix, rawKey string) {
	// Look for the first colon.
	colonIndex := strings.Index(key, ":")

	if colonIndex != -1 {
		// colon found: the key has a module prefix.
		prefix = key[:colonIndex]
		rawKey = key[colonIndex+1:]

		return
	}
	return "", key
}

//----------------------------------------

type prefixParamsKeeper struct {
	prefix string
	pk     ParamsKeeper
}

func newPrefixParamsKeeper(pk ParamsKeeper, prefix string) prefixParamsKeeper {
	return prefixParamsKeeper{
		prefix: prefix,
		pk:     pk,
	}
}

func (ppk prefixParamsKeeper) prefixed(key string) string {
	return ppk.prefix + key
}

func (ppk prefixParamsKeeper) GetString(ctx sdk.Context, key string, ptr *string) {
	ppk.pk.GetString(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) GetInt64(ctx sdk.Context, key string, ptr *int64) {
	ppk.pk.GetInt64(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) GetUint64(ctx sdk.Context, key string, ptr *uint64) {
	ppk.pk.GetUint64(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) GetBool(ctx sdk.Context, key string, ptr *bool) {
	ppk.pk.GetBool(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) GetBytes(ctx sdk.Context, key string, ptr *[]byte) {
	ppk.pk.GetBytes(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) SetString(ctx sdk.Context, key string, value string) {
	ppk.pk.SetString(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) SetInt64(ctx sdk.Context, key string, value int64) {
	ppk.pk.SetInt64(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) SetUint64(ctx sdk.Context, key string, value uint64) {
	ppk.pk.SetUint64(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) SetBool(ctx sdk.Context, key string, value bool) {
	ppk.pk.SetBool(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) SetBytes(ctx sdk.Context, key string, value []byte) {
	ppk.pk.SetBytes(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) Has(ctx sdk.Context, key string) bool {
	return ppk.pk.Has(ctx, ppk.prefixed(key))
}

func (ppk prefixParamsKeeper) GetRaw(ctx sdk.Context, key string) []byte {
	return ppk.pk.GetRaw(ctx, ppk.prefixed(key))
}

func (ppk prefixParamsKeeper) SetRaw(ctx sdk.Context, key string, value []byte) {
	ppk.pk.SetRaw(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) GetStruct(ctx sdk.Context, key string, paramPtr interface{}) {
	ppk.pk.GetStruct(ctx, ppk.prefixed(key), paramPtr)
}

func (ppk prefixParamsKeeper) SetStruct(ctx sdk.Context, key string, param interface{}) {
	ppk.pk.SetStruct(ctx, ppk.prefixed(key), param)
}
