package params

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/store"
)

var ErrMissingParamValue = errors.New("missing param value")

const (
	stringSuffix = ".string"
	boolSuffix   = ".bool"
	int64Suffix  = ".int64"
	uint64Suffix = ".uint64"
	bytesSuffix  = ".bytes"
)

// Keeper is the global param store keeper
// TODO: The keeper, and its functionality is not thread safe,
// because the underlying store is not thread safe. Check if this is expected behavior.
type Keeper struct {
	key    store.StoreKey
	prefix string
}

// NewParamsKeeper returns a new ParamsKeeper.
func NewParamsKeeper(key store.StoreKey, prefix string) Keeper {
	return Keeper{
		key:    key,
		prefix: prefix,
	}
}

func (pk Keeper) Has(ctx sdk.Context, key string) bool {
	s := ctx.Store(pk.key)

	return s.Has([]byte(key))
}

func (pk Keeper) GetRaw(ctx sdk.Context, key string) []byte {
	s := ctx.Store(pk.key)

	return s.Get([]byte(key))
}

func (pk Keeper) GetString(ctx sdk.Context, key string) (string, error) {
	if err := checkSuffix(key, stringSuffix); err != nil {
		return "", fmt.Errorf("invalid suffix, %w", err)
	}

	return get[string](ctx.Store(pk.key), key)
}

func (pk Keeper) GetBool(ctx sdk.Context, key string) (bool, error) {
	if err := checkSuffix(key, boolSuffix); err != nil {
		return false, fmt.Errorf("invalid suffix, %w", err)
	}

	return get[bool](ctx.Store(pk.key), key)
}

func (pk Keeper) GetInt64(ctx sdk.Context, key string) (int64, error) {
	if err := checkSuffix(key, int64Suffix); err != nil {
		return 0, fmt.Errorf("invalid suffix, %w", err)
	}

	return get[int64](ctx.Store(pk.key), key)
}

func (pk Keeper) GetUint64(ctx sdk.Context, key string) (uint64, error) {
	if err := checkSuffix(key, uint64Suffix); err != nil {
		return 0, fmt.Errorf("invalid suffix, %w", err)
	}

	return get[uint64](ctx.Store(pk.key), key)
}

func (pk Keeper) GetBytes(ctx sdk.Context, key string) ([]byte, error) {
	if err := checkSuffix(key, bytesSuffix); err != nil {
		return nil, fmt.Errorf("invalid suffix, %w", err)
	}

	return get[[]byte](ctx.Store(pk.key), key)
}

func (pk Keeper) SetString(ctx sdk.Context, key, value string) error {
	if err := checkSuffix(key, stringSuffix); err != nil {
		return fmt.Errorf("invalid suffix, %w", err)
	}

	return set(ctx.Store(pk.key), key, value)
}

func (pk Keeper) SetBool(ctx sdk.Context, key string, value bool) error {
	if err := checkSuffix(key, boolSuffix); err != nil {
		return fmt.Errorf("invalid suffix, %w", err)
	}

	return set(ctx.Store(pk.key), key, value)
}

func (pk Keeper) SetInt64(ctx sdk.Context, key string, value int64) error {
	if err := checkSuffix(key, int64Suffix); err != nil {
		return fmt.Errorf("invalid suffix, %w", err)
	}

	return set(ctx.Store(pk.key), key, value)
}

func (pk Keeper) SetUint64(ctx sdk.Context, key string, value uint64) error {
	if err := checkSuffix(key, uint64Suffix); err != nil {
		return fmt.Errorf("invalid suffix, %w", err)
	}

	return set(ctx.Store(pk.key), key, value)
}

func (pk Keeper) SetBytes(ctx sdk.Context, key string, value []byte) error {
	if err := checkSuffix(key, bytesSuffix); err != nil {
		return fmt.Errorf("invalid suffix, %w", err)
	}

	return set(ctx.Store(pk.key), key, value)
}

type keeperType interface {
	[]byte | uint64 | int64 | bool | string
}

// get fetches the value associated with the given key, if any
func get[T keeperType](
	s store.Store,
	key string,
) (T, error) {
	var ret T

	bz := s.Get([]byte(key))

	if bz == nil {
		return ret, fmt.Errorf("%w: %s", ErrMissingParamValue, key)
	}

	if err := amino.UnmarshalJSON(bz, &ret); err != nil {
		return ret, fmt.Errorf("unable to parse value, %w", err)
	}

	return ret, nil
}

// set saves the given key value pair in the store
func set[T keeperType](
	s store.Store,
	key string,
	value T,
) error {
	bz, err := amino.MarshalJSON(value)
	if err != nil {
		return fmt.Errorf("unable to marshal value, %w", err)
	}

	s.Set([]byte(key), bz)

	return nil
}

// checkSuffix checks if the full key has the given suffix,
// and if the key is empty
func checkSuffix(key, expectedSuffix string) error {
	var (
		noSuffix = !strings.HasSuffix(key, expectedSuffix)
		noName   = len(key) == len(expectedSuffix)
	)

	if noSuffix || noName {
		return fmt.Errorf("key should be like \"<name>%s\" (%s)", expectedSuffix, key)
	}

	return nil
}
