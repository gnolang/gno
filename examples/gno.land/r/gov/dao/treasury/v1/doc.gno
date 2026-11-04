// Package treasury (v1) is the GovDAO treasury built on
// gno.land/p/nt/treasury/v1: its GRC20 banker pays gno.land/p/nt/grc20/v1
// tokens listed in gno.land/r/nt/grc20reg/v1, and every amount is a uint256.
// Coin payments are range-checked down to int64 when they are built.
//
// It is a separate realm from r/gov/dao/treasury/v0, with its own address:
// funds held by v0 stay there until a v0 payment proposal moves them here.
// Only realms in r/gov/dao's AllowedDAOs may send, which after the impl/v1
// install are impl/v0 and impl/v1.
package treasury
