package antisquat

// PricingTier defines the registration fee for a range of name lengths.
type PricingTier struct {
	MinLength int
	MaxLength int
	Fee       int64 // in ugnot
}

// Config holds all configuration for the anti-squatting system.
type Config struct {
	MinNameLength             int
	MaxNameLength             int
	MaxRegistrationsPerWindow int
	RateLimitWindowBlocks     int64
	RenewalPeriodBlocks       int64
	GracePeriodBlocks         int64
	PricingTiers              []PricingTier
}

// DefaultConfig returns sensible defaults for mainnet deployment.
//
// Pricing (in ugnot, where 1 GNOT = 1,000,000 ugnot):
//   - 1-3 chars:  1,000,000,000 ugnot (1000 GNOT)
//   - 4 chars:      100,000,000 ugnot  (100 GNOT)
//   - 5 chars:       50,000,000 ugnot   (50 GNOT)
//   - 6-10 chars:    10,000,000 ugnot   (10 GNOT)
//   - 11-64 chars:    5,000,000 ugnot    (5 GNOT)
//
// Rate limit: 3 registrations per 432,000 blocks (~30 days at 6s/block).
// Renewal period: 5,256,000 blocks (~1 year at 6s/block).
// Grace period: 432,000 blocks (~30 days at 6s/block).
func DefaultConfig() *Config {
	return &Config{
		MinNameLength:             1,
		MaxNameLength:             64,
		MaxRegistrationsPerWindow: 3,
		RateLimitWindowBlocks:     432_000,   // ~30 days
		RenewalPeriodBlocks:       5_256_000, // ~1 year
		GracePeriodBlocks:         432_000,   // ~30 days
		PricingTiers: []PricingTier{
			{MinLength: 1, MaxLength: 3, Fee: 1_000_000_000},  // 1000 GNOT
			{MinLength: 4, MaxLength: 4, Fee: 100_000_000},    // 100 GNOT
			{MinLength: 5, MaxLength: 5, Fee: 50_000_000},     // 50 GNOT
			{MinLength: 6, MaxLength: 10, Fee: 10_000_000},    // 10 GNOT
			{MinLength: 11, MaxLength: 64, Fee: 5_000_000},    // 5 GNOT
		},
	}
}
