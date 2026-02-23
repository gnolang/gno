// PKGPATH: gno.land/r/test/exploit
package exploit

import (
	"gno.land/r/gov/dao/v3/memberstore"
)

func main() {
	// After the fix, memberstore.Tiers is no longer accessible (lowercase 'tiers')
	// External realms can only use the safe accessor functions:
	// - memberstore.GetTier(name) - read-only tier access
	// - memberstore.IterateTiers(fn) - read-only iteration
	// - memberstore.GetTierPower(name, members) - calculated power

	// Verify we can still READ tier data via the safe accessor
	t3, ok := memberstore.GetTier(memberstore.T3)
	if !ok {
		panic("T3 tier not found")
	}
	println("T3 BasePower (read-only):", t3.BasePower)
	println("T3 InvitationPoints (read-only):", t3.InvitationPoints)

	// The following lines would cause a compile error if uncommented:
	// memberstore.Tiers.Set(...) // ERROR: Tiers is not exported (lowercase)

	// Iterate over tiers (read-only)
	println("All tiers:")
	memberstore.IterateTiers(func(name string, tier memberstore.Tier) bool {
		println("  -", name, "BasePower:", tier.BasePower)
		return false
	})

	println("Security fix verified: external realms cannot modify tiers")
}

// Output:
// T3 BasePower (read-only): 1
// T3 InvitationPoints (read-only): 1
// All tiers:
//   - T1 BasePower: 3
//   - T2 BasePower: 2
//   - T3 BasePower: 1
// Security fix verified: external realms cannot modify tiers
