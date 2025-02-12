package launchpad

const (
	minimumGnsAmount = 1_000_000
)

const (
	TIMESTAMP_180DAYS = uint64(180 * 24 * 60 * 60) // 15552000
	TIMESTAMP_14DAYS  = uint64(14 * 24 * 60 * 60)  // 1209600
	TIMESTAMP_90DAYS  = uint64(90 * 24 * 60 * 60)  // 7776000
	TIMESTAMP_7DAYS   = uint64(7 * 24 * 60 * 60)   // 604800
	TIMESTAMP_30DAYS  = uint64(30 * 24 * 60 * 60)  // 2592000
	TIMESTAMP_3DAYS   = uint64(3 * 24 * 60 * 60)   // 259200
	TIMESTAMP_DAY     = uint64(24 * 60 * 60)       // 86400

	TIER30      = uint64(30)
	TIER90      = uint64(90)
	TIER180     = uint64(180)
	TOTAL_TIERS = 3

	PAD_SEP = "*PAD*"

	MINIMUM_DEPOSIT_AMOUNT = uint64(1_000_000)
)
