package consts

import (
	u256 "gno.land/p/gnoswap/uint256"
)

// WAD represents a decimal with 18 decimals (i.e. 1e18)
var WAD = u256.NewUint(1000000000000000000)

// Q96 represents 2^96
var Q96 = u256.MustFromDecimal("79228162514264337593543950336") // 2^96

// Q192 represents 2^192
var Q192 = u256.MustFromDecimal("6277101735386680763835789423207666416102355444464034512896") // 2^192

// ORACLE_PRICE_SCALE represents 1e36 for oracle price precision
var ORACLE_PRICE_SCALE = u256.MustFromDecimal("1000000000000000000000000000000000000") // 1e36

// MAX_FEE represents maximum fee (30% in WAD)
var MAX_FEE = u256.NewUint(300000000000000000) // 30%

// LIQUIDATION_CURSOR represents 0.3 in WAD for liquidation incentive calculation
var LIQUIDATION_CURSOR = u256.NewUint(300000000000000000) // 0.3 in WAD

// MAX_LIQUIDATION_INCENTIVE_FACTOR represents maximum liquidation incentive (1.15 in WAD)
var MAX_LIQUIDATION_INCENTIVE_FACTOR = u256.NewUint(1150000000000000000) // 1.15 in WAD
