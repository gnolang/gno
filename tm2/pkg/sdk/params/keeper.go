package params

// XXX Rename ParamsKeeper to ParamKeeper, like AccountKeeper is singular.

import (
	"fmt"
	"log/slog"
	"reflect"
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
	// GetXxx writes the stored value (if any) into *ptr and returns
	// whether the key existed. A return of false leaves *ptr at its
	// zero value, distinguishing "never set" from "set to zero" —
	// which the in-memory backed types alone could not.
	GetString(ctx sdk.Context, key string, ptr *string) bool
	GetInt64(ctx sdk.Context, key string, ptr *int64) bool
	GetUint64(ctx sdk.Context, key string, ptr *uint64) bool
	GetBool(ctx sdk.Context, key string, ptr *bool) bool
	GetBytes(ctx sdk.Context, key string, ptr *[]byte) bool
	GetStrings(ctx sdk.Context, key string, ptr *[]string) bool

	// SetXxx writes value under key and returns the byte delta
	// (newSize - oldSize, with key bytes added on first-create or
	// subtracted on delete). Callers that need the delta for storage-
	// deposit accounting use the return value; callers that don't can
	// ignore it (Go discards return values of expression statements).
	// The internal Get to resolve oldSize is unmetered — its leaf-find
	// is amortized into the Set's set-read-depth charge that fires at
	// commit time. Saves ~ReadCostFlat × set-read-depth gas vs. callers
	// that did `GetBytes; SetBytes` for the same intent.
	SetString(ctx sdk.Context, key string, value string) int
	SetInt64(ctx sdk.Context, key string, value int64) int
	SetUint64(ctx sdk.Context, key string, value uint64) int
	SetBool(ctx sdk.Context, key string, value bool) int
	SetBytes(ctx sdk.Context, key string, value []byte) int
	SetStrings(ctx sdk.Context, key string, value []string) int

	Has(ctx sdk.Context, key string) bool

	GetStruct(ctx sdk.Context, key string, strctPtr any)
	SetStruct(ctx sdk.Context, key string, strct any)

	// NOTE: GetAny and SetAny don't work on structs.
	GetAny(ctx sdk.Context, key string) any
	SetAny(ctx sdk.Context, key string, value any)
}

type ParamfulKeeper interface {
	WillSetParam(ctx sdk.Context, key string, value any)
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

func (pk ParamsKeeper) GetRegisteredKeeper(moduleName string) (ParamfulKeeper, bool) {
	rk, ok := pk.kprs[moduleName]
	return rk, ok
}

func (pk ParamsKeeper) Register(moduleName string, pmk ParamfulKeeper) {
	if pmk == nil {
		panic("cannot register nil keeper")
	}
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
	gctx := ctx.GasContext()
	stor := ctx.Store(pk.key)
	return stor.Has(gctx, storeKey(key))
}

func (pk ParamsKeeper) GetString(ctx sdk.Context, key string, ptr *string) bool {
	return pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetBool(ctx sdk.Context, key string, ptr *bool) bool {
	return pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetInt64(ctx sdk.Context, key string, ptr *int64) bool {
	return pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetUint64(ctx sdk.Context, key string, ptr *uint64) bool {
	return pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) GetBytes(ctx sdk.Context, key string, ptr *[]byte) bool {
	gctx := ctx.GasContext()
	stor := ctx.Store(pk.key)
	bz := stor.Get(gctx, storeKey(key))
	if bz == nil {
		return false
	}
	*ptr = bz
	return true
}

func (pk ParamsKeeper) GetStrings(ctx sdk.Context, key string, ptr *[]string) bool {
	return pk.getIfExists(ctx, key, ptr)
}

func (pk ParamsKeeper) SetString(ctx sdk.Context, key, value string) int {
	return pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetBool(ctx sdk.Context, key string, value bool) int {
	return pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetInt64(ctx sdk.Context, key string, value int64) int {
	return pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetUint64(ctx sdk.Context, key string, value uint64) int {
	return pk.set(ctx, key, value)
}

func (pk ParamsKeeper) SetBytes(ctx sdk.Context, key string, value []byte) int {
	// H1: route validation through the same hook as `set`, so module
	// keepers' WillSetParam fires for byte-typed writes too. Storage
	// stays raw (GetBytes reads raw, no amino unmarshal) — we only
	// borrow the validate step from `set`, not the JSON encoding.
	pk.validate(ctx, key, value)
	gctx := ctx.GasContext()
	stor := ctx.Store(pk.key)
	skey := storeKey(key)
	// Free Get — Set's set-read-depth charge at commit covers the
	// leaf-find traversal, so charging the read again is redundant.
	oldbz := stor.Get(nil, skey)
	if value == nil {
		stor.Delete(gctx, skey)
		if oldbz == nil {
			return 0 // no-op delete
		}
		return -(len(key) + len(oldbz)) // delete: key+val gone
	}
	// Copy to avoid altering the input bytes
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	stor.Set(gctx, skey, valueCopy)
	diff := len(value) - len(oldbz)
	if oldbz == nil {
		diff += len(key) // first-create: key bytes count too
	}
	return diff
}

func (pk ParamsKeeper) SetStrings(ctx sdk.Context, key string, value []string) int {
	return pk.set(ctx, key, value)
}

func (pk ParamsKeeper) GetStruct(ctx sdk.Context, key string, strctPtr any) {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		panic("struct key expected format <module>:<struct name>")
	}
	moduleName := parts[0]
	structName := parts[1] // <submodule>
	if !pk.IsRegistered(moduleName) {
		panic("unregistered module name")
	}
	if structName != "p" {
		panic("the only supported struct name is 'p'")
	}
	gctx := ctx.GasContext()
	stor := ctx.Store(pk.key)
	kvz := getStructFieldsFromStore(gctx, strctPtr, stor, storeKey(key))
	decodeStructFields(strctPtr, kvz)
}

func (pk ParamsKeeper) SetStruct(ctx sdk.Context, key string, strct any) {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		panic("struct key expected format <module>:<struct name>")
	}
	moduleName := parts[0]
	structName := parts[1] // <submodule>
	if !pk.IsRegistered(moduleName) {
		panic("unregistered module name")
	}
	if structName != "p" {
		panic("the only supported struct name is 'p'")
	}
	gctx := ctx.GasContext()
	stor := ctx.Store(pk.key)
	kvz := encodeStructFields(strct)
	for _, kv := range kvz {
		stor.Set(gctx, storeKey(key+":"+string(kv.Key)), kv.Value)
	}
}

func (pk ParamsKeeper) GetAny(ctx sdk.Context, key string) any {
	panic("not yet implemented")
}

func (pk ParamsKeeper) SetAny(ctx sdk.Context, key string, value any) {
	switch value := value.(type) {
	case string:
		pk.SetString(ctx, key, value)
	case int64:
		pk.SetInt64(ctx, key, value)
	case uint64:
		pk.SetUint64(ctx, key, value)
	case bool:
		pk.SetBool(ctx, key, value)
	case []byte:
		pk.SetBytes(ctx, key, value)
	case []string:
		pk.SetStrings(ctx, key, value)
	default:
		panic(fmt.Sprintf("unexected value type for SetAny: %v", reflect.TypeOf(value)))
	}
}

func (pk ParamsKeeper) getIfExists(ctx sdk.Context, key string, ptr any) bool {
	gctx := ctx.GasContext()
	stor := ctx.Store(pk.key)
	bz := stor.Get(gctx, storeKey(key))
	if bz == nil {
		return false
	}
	amino.MustUnmarshalJSON(bz, ptr)
	return true
}

// validate runs the registered module keeper's WillSetParam hook for
// the prefix in key. Extracted from `set` so that SetBytes (which
// keeps raw storage) can also enforce module-level validation.
func (pk ParamsKeeper) validate(ctx sdk.Context, key string, value any) {
	module, rawKey := parsePrefix(key)
	if module == "" {
		return
	}
	kpr, ok := pk.GetRegisteredKeeper(module)
	if !ok {
		panic("module not registered: " + module)
	}
	kpr.WillSetParam(ctx, rawKey, value)
}

// set marshals value, writes it under key, and returns the byte delta
// (newSize - oldSize, with key bytes added on first-create). The Get
// to resolve oldSize uses nil gctx — Set's set-read-depth charge at
// commit time covers the leaf-find traversal, so charging the read
// again would be redundant. Uses len(key) (the unprefixed key passed
// by the caller) for first-create accounting, matching the convention
// used by gno.land/pkg/sdk/vm/params_deposit.go's recordParamsDelta.
func (pk ParamsKeeper) set(ctx sdk.Context, key string, value any) int {
	pk.validate(ctx, key, value)
	gctx := ctx.GasContext()
	stor := ctx.Store(pk.key)
	newbz := amino.MustMarshalJSON(value)
	skey := storeKey(key)
	oldbz := stor.Get(nil, skey)
	diff := len(newbz) - len(oldbz)
	if oldbz == nil {
		diff += len(key)
	}
	stor.Set(gctx, skey, newbz)
	return diff
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

func (ppk prefixParamsKeeper) GetString(ctx sdk.Context, key string, ptr *string) bool {
	return ppk.pk.GetString(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) GetInt64(ctx sdk.Context, key string, ptr *int64) bool {
	return ppk.pk.GetInt64(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) GetUint64(ctx sdk.Context, key string, ptr *uint64) bool {
	return ppk.pk.GetUint64(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) GetBool(ctx sdk.Context, key string, ptr *bool) bool {
	return ppk.pk.GetBool(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) GetBytes(ctx sdk.Context, key string, ptr *[]byte) bool {
	return ppk.pk.GetBytes(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) GetStrings(ctx sdk.Context, key string, ptr *[]string) bool {
	return ppk.pk.GetStrings(ctx, ppk.prefixed(key), ptr)
}

func (ppk prefixParamsKeeper) SetString(ctx sdk.Context, key string, value string) int {
	return ppk.pk.SetString(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) SetInt64(ctx sdk.Context, key string, value int64) int {
	return ppk.pk.SetInt64(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) SetUint64(ctx sdk.Context, key string, value uint64) int {
	return ppk.pk.SetUint64(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) SetBool(ctx sdk.Context, key string, value bool) int {
	return ppk.pk.SetBool(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) SetBytes(ctx sdk.Context, key string, value []byte) int {
	return ppk.pk.SetBytes(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) SetStrings(ctx sdk.Context, key string, value []string) int {
	return ppk.pk.SetStrings(ctx, ppk.prefixed(key), value)
}

func (ppk prefixParamsKeeper) Has(ctx sdk.Context, key string) bool {
	return ppk.pk.Has(ctx, ppk.prefixed(key))
}

func (ppk prefixParamsKeeper) GetStruct(ctx sdk.Context, key string, paramPtr any) {
	ppk.pk.GetStruct(ctx, ppk.prefixed(key), paramPtr)
}

func (ppk prefixParamsKeeper) SetStruct(ctx sdk.Context, key string, param any) {
	ppk.pk.SetStruct(ctx, ppk.prefixed(key), param)
}

func (ppk prefixParamsKeeper) GetAny(ctx sdk.Context, key string) any {
	return ppk.pk.GetAny(ctx, ppk.prefixed(key))
}

func (ppk prefixParamsKeeper) SetAny(ctx sdk.Context, key string, value any) {
	ppk.pk.SetAny(ctx, ppk.prefixed(key), value)
}
