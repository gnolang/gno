package bank

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

var (
	errEmptyDenom         = errors.New("denom is empty")
	errInsufficientSupply = errors.New("insufficient supply")
)

// SupplyKeeper is an interface for managing supply of a token
type SupplyKeeper interface {
	GetSupply(store types.Store, denom string) int64
	SetSupply(store types.Store, denom string, amount int64)
	AddSupply(store types.Store, denom string, amount int64)
	SubtractSupply(store types.Store, denom string, amount int64)
}

// SupplyStore manages the actual supply data
type SupplyStore struct {
	store types.Store
}

func NewSupplyStore(store types.Store) *SupplyStore {
	return &SupplyStore{store: store}
}

func generateSupplyKey(denom string) []byte {
	return fmt.Appendf(nil, "supply:%s", denom)
}

// GetSupply returns the total supply of a specific token
func (s *SupplyStore) GetSupply(store types.Store, denom string) (int64, error) {
	if denom == "" {
		return 0, errEmptyDenom
	}

	key := generateSupplyKey(denom)
	bz := store.Get(key)
	if bz == nil {
		return 0, nil
	}

	var amount int64
	amino.MustUnmarshal(bz, &amount)

	return amount, nil
}

// SetSupply sets the total supply of a specific token
func (s *SupplyStore) SetSupply(store types.Store, denom string, amount int64) error {
	if denom == "" {
		return errEmptyDenom
	}

	key := generateSupplyKey(denom)
	bz, err := amino.Marshal(amount)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// AddSupply increases the total supply of a specific token
func (s *SupplyStore) AddSupply(store types.Store, denom string, amount int64) error {
	current, err := s.GetSupply(store, denom)
	if err != nil {
		return err
	}
	return s.SetSupply(store, denom, current+amount)
}

// SubtractSupply decreases the total supply of a specific token
func (s *SupplyStore) SubtractSupply(store types.Store, denom string, amount int64) error {
	current, err := s.GetSupply(store, denom)
	if err != nil {
		return err
	}
	if current < amount {
		return errInsufficientSupply
	}
	return s.SetSupply(store, denom, current-amount)
}
