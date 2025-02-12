package consts

import (
	"std"
)

const (
	// Protocol configuration
	PROTOCOL_NAME = "GnoLend"
	VERSION      = "1.0.0"

	// Risk parameters
	MAX_LTV_RATIO              = 0.75 // 75% maximum loan-to-value ratio
	LIQUIDATION_THRESHOLD      = 0.80 // 80% liquidation threshold
	LIQUIDATION_BONUS         = 0.05 // 5% bonus for liquidators
	MAX_LIQUIDATION_CLOSE     = 0.50 // Maximum 50% of a position can be liquidated at once
	
	// Interest rate parameters
	MIN_SUPPLY_RATE           = 0.01 // 1% minimum supply rate
	MAX_SUPPLY_RATE           = 0.50 // 50% maximum supply rate
	MIN_BORROW_RATE          = 0.02 // 2% minimum borrow rate
	MAX_BORROW_RATE          = 1.00 // 100% maximum borrow rate
	
	// Protocol fees
	PROTOCOL_FEE             = 0.10 // 10% of interest goes to protocol
	
	// Time constants
	BLOCKS_PER_YEAR         = 31_536_000 // Assuming 1s block time
	
	// Math constants
	PRECISION              = 1e18
)

var (
	// Protocol admin address
	ADMIN std.Address = "g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8"
) 