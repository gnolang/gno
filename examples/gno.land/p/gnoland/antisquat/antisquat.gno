// Package antisquat provides a pluggable anti-squatting library for gno.land
// namespace registration. It implements rate limiting, protected name management,
// length-based pricing, and name lifecycle management.
//
// This package is designed to be imported by any realm that needs namespace
// management. It handles state and logic, while the importing realm handles
// authorization (GovDAO gating) and coin collection.
//
// Design decisions addressing review feedback:
//   - O(log n) lookups via per-address AVL tree indexing
//   - Per-address pruning only (no global iteration)
//   - Dynamic protected names with add/remove functions
//   - Correct ugnot pricing (1 GNOT = 1,000,000 ugnot)
//   - Explicit fee handling (CalculateFee returns fee, realm collects)
package antisquat

import (
	"std"

	"gno.land/p/nt/avl"
)

// AntiSquat is the main anti-squatting system.
type AntiSquat struct {
	config    *Config
	names     *avl.Tree // name -> *NameRecord
	addresses *avl.Tree // addr.String() -> *avl.Tree (name -> bool)
	rateLimit *avl.Tree // addr.String() -> *rateLimitEntry
	reserved  *avl.Tree // name -> ReservedCategory
}

// New creates a new AntiSquat instance with the given configuration.
// If config is nil, DefaultConfig() is used.
func New(config *Config) *AntiSquat {
	if config == nil {
		config = DefaultConfig()
	}
	return &AntiSquat{
		config:    config,
		names:     avl.NewTree(),
		addresses: avl.NewTree(),
		rateLimit: avl.NewTree(),
		reserved:  avl.NewTree(),
	}
}

// CanRegister checks whether a name can be registered by the given caller.
// This method does not mutate any state.
func (as *AntiSquat) CanRegister(name string, caller std.Address) error {
	normalized := NormalizeName(name)

	if err := ValidateName(normalized, as.config); err != nil {
		return err
	}

	if as.IsReserved(normalized) {
		return ErrNameReserved
	}

	if record := as.GetRecord(normalized); record != nil {
		currentHeight := std.GetHeight()
		if !record.IsPastGrace(currentHeight, as.config.GracePeriodBlocks) {
			return ErrNameTaken
		}
	}

	if as.isRateLimited(caller) {
		return ErrRateLimited
	}

	return nil
}

// Register registers a name for the given owner.
// The calling realm is responsible for collecting the registration fee
// (use CalculateFee to determine the amount).
func (as *AntiSquat) Register(name string, owner std.Address) (*NameRecord, error) {
	normalized := NormalizeName(name)

	if err := as.CanRegister(normalized, owner); err != nil {
		return nil, err
	}

	currentHeight := std.GetHeight()

	// If the name existed but is past grace, clean up the old record.
	if oldRecord := as.GetRecord(normalized); oldRecord != nil {
		as.removeNameFromAddress(oldRecord.Owner, normalized)
		as.names.Remove(normalized)
	}

	record := &NameRecord{
		Name:         normalized,
		Owner:        owner,
		RegisteredAt: currentHeight,
		ExpiresAt:    currentHeight + as.config.RenewalPeriodBlocks,
		RenewalFee:   as.CalculateFee(normalized),
	}

	as.names.Set(normalized, record)
	as.addNameToAddress(owner, normalized)
	as.recordRegistration(owner, currentHeight)

	return record, nil
}

// Renew extends the expiry of a name owned by the caller.
// Returns the renewal fee in ugnot.
func (as *AntiSquat) Renew(name string, caller std.Address) (int64, error) {
	normalized := NormalizeName(name)

	record := as.GetRecord(normalized)
	if record == nil {
		return 0, ErrNameNotFound
	}

	if record.Owner != caller {
		return 0, ErrNotOwner
	}

	if record.ExpiresAt == 0 {
		return 0, ErrNameNotExpirable
	}

	currentHeight := std.GetHeight()

	if record.IsPastGrace(currentHeight, as.config.GracePeriodBlocks) {
		return 0, ErrNameExpired
	}

	if record.IsExpired(currentHeight) {
		record.ExpiresAt = currentHeight + as.config.RenewalPeriodBlocks
	} else {
		record.ExpiresAt = record.ExpiresAt + as.config.RenewalPeriodBlocks
	}

	as.names.Set(normalized, record)

	return record.RenewalFee, nil
}

// Resolve returns the owner address for a registered, non-expired name.
func (as *AntiSquat) Resolve(name string) std.Address {
	normalized := NormalizeName(name)
	record := as.GetRecord(normalized)
	if record == nil {
		return std.Address("")
	}

	currentHeight := std.GetHeight()
	if record.IsExpired(currentHeight) {
		return std.Address("")
	}

	return record.Owner
}

// GetRecord returns the NameRecord for a given name, or nil if not found.
func (as *AntiSquat) GetRecord(name string) *NameRecord {
	normalized := NormalizeName(name)
	v, found := as.names.Get(normalized)
	if !found {
		return nil
	}
	return v.(*NameRecord)
}

// GetNamesForAddress returns all names registered to the given address.
func (as *AntiSquat) GetNamesForAddress(addr std.Address) []string {
	addrKey := string(addr)
	v, found := as.addresses.Get(addrKey)
	if !found {
		return nil
	}

	addrTree := v.(*avl.Tree)
	names := make([]string, 0, addrTree.Size())
	addrTree.Iterate("", "", func(key string, _ any) bool {
		names = append(names, key)
		return false
	})
	return names
}

// CalculateFee returns the registration fee in ugnot for a given name.
func (as *AntiSquat) CalculateFee(name string) int64 {
	normalized := NormalizeName(name)
	nameLen := len(normalized)

	for _, tier := range as.config.PricingTiers {
		if nameLen >= tier.MinLength && nameLen <= tier.MaxLength {
			return tier.Fee
		}
	}

	return 0
}

// IsReserved returns true if the given name is in the reserved names list.
func (as *AntiSquat) IsReserved(name string) bool {
	normalized := NormalizeName(name)
	return as.reserved.Has(normalized)
}

// GetReservedCategory returns the ReservedCategory for a reserved name.
func (as *AntiSquat) GetReservedCategory(name string) ReservedCategory {
	normalized := NormalizeName(name)
	v, found := as.reserved.Get(normalized)
	if !found {
		return 0
	}
	return v.(ReservedCategory)
}

// AddReservedName adds a name to the reserved list.
// The calling realm is responsible for gating this behind GovDAO authorization.
func (as *AntiSquat) AddReservedName(name string, category ReservedCategory) {
	normalized := NormalizeName(name)
	as.reserved.Set(normalized, category)
}

// RemoveReservedName removes a name from the reserved list.
// The calling realm is responsible for gating this behind GovDAO authorization.
func (as *AntiSquat) RemoveReservedName(name string) {
	normalized := NormalizeName(name)
	as.reserved.Remove(normalized)
}

// TotalRegistered returns the total number of registered names.
func (as *AntiSquat) TotalRegistered() int {
	return as.names.Size()
}

// GetConfig returns the current configuration.
func (as *AntiSquat) GetConfig() *Config {
	return as.config
}

// isRateLimited checks if the given address has exceeded the registration
// rate limit within the current window.
//
// Complexity: O(log n) AVL lookup + O(k) scan where k <= MaxRegistrationsPerWindow.
func (as *AntiSquat) isRateLimited(addr std.Address) bool {
	addrKey := string(addr)
	v, found := as.rateLimit.Get(addrKey)
	if !found {
		return false
	}

	entry := v.(*rateLimitEntry)
	currentHeight := std.GetHeight()
	windowStart := currentHeight - as.config.RateLimitWindowBlocks

	count := 0
	for i := len(entry.Heights) - 1; i >= 0; i-- {
		if entry.Heights[i] >= windowStart {
			count++
		} else {
			break
		}
	}

	return count >= as.config.MaxRegistrationsPerWindow
}

// recordRegistration records a new registration for the given address
// and prunes old entries outside the rate limit window for THIS address only.
//
// Complexity: O(log n) AVL lookup + O(k) prune where k <= MaxRegistrationsPerWindow.
func (as *AntiSquat) recordRegistration(addr std.Address, height int64) {
	addrKey := string(addr)
	v, found := as.rateLimit.Get(addrKey)

	var entry *rateLimitEntry
	if found {
		entry = v.(*rateLimitEntry)
	} else {
		entry = &rateLimitEntry{}
	}

	windowStart := height - as.config.RateLimitWindowBlocks
	pruneIdx := 0
	for pruneIdx < len(entry.Heights) {
		if entry.Heights[pruneIdx] >= windowStart {
			break
		}
		pruneIdx++
	}
	if pruneIdx > 0 {
		entry.Heights = entry.Heights[pruneIdx:]
	}

	entry.Heights = append(entry.Heights, height)
	as.rateLimit.Set(addrKey, entry)
}

// addNameToAddress adds a name to the per-address index.
func (as *AntiSquat) addNameToAddress(addr std.Address, name string) {
	addrKey := string(addr)
	v, found := as.addresses.Get(addrKey)

	var addrTree *avl.Tree
	if found {
		addrTree = v.(*avl.Tree)
	} else {
		addrTree = avl.NewTree()
	}

	addrTree.Set(name, true)
	as.addresses.Set(addrKey, addrTree)
}

// removeNameFromAddress removes a name from the per-address index.
func (as *AntiSquat) removeNameFromAddress(addr std.Address, name string) {
	addrKey := string(addr)
	v, found := as.addresses.Get(addrKey)
	if !found {
		return
	}

	addrTree := v.(*avl.Tree)
	addrTree.Remove(name)

	if addrTree.Size() == 0 {
		as.addresses.Remove(addrKey)
	} else {
		as.addresses.Set(addrKey, addrTree)
	}
}
