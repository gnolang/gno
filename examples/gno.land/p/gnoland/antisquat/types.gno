package antisquat

import "std"

// NameRecord represents a registered namespace entry.
type NameRecord struct {
	Name         string
	Owner        std.Address
	RegisteredAt int64 // block height
	ExpiresAt    int64 // block height; 0 means no expiry
	RenewalFee   int64 // in ugnot
}

// IsExpired returns true if the name has passed its expiry block height.
func (r *NameRecord) IsExpired(currentHeight int64) bool {
	if r.ExpiresAt == 0 {
		return false
	}
	return currentHeight > r.ExpiresAt
}

// IsPastGrace returns true if the name is beyond the grace period after expiry.
func (r *NameRecord) IsPastGrace(currentHeight int64, gracePeriod int64) bool {
	if r.ExpiresAt == 0 {
		return false
	}
	return currentHeight > r.ExpiresAt+gracePeriod
}

// ReservedCategory indicates why a name is reserved.
type ReservedCategory int

const (
	ReservedSystem     ReservedCategory = iota + 1
	ReservedGovernance
	ReservedBrand
)

// String returns the string representation of a ReservedCategory.
func (c ReservedCategory) String() string {
	switch c {
	case ReservedSystem:
		return "system"
	case ReservedGovernance:
		return "governance"
	case ReservedBrand:
		return "brand"
	default:
		return "unknown"
	}
}

// rateLimitEntry tracks registration block heights for a single address.
type rateLimitEntry struct {
	Heights []int64
}
