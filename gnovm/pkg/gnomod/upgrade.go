package gnomod

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// CheckUpgradeFields checks the version and [upgrade] fields, on their own
// and against live, the gnomod.toml of the package a redeploy replaces, nil
// on a first deployment. isRealm says whether Module is a realm path, which
// this package cannot tell. The chain and `gno upgrade check` share it.
func (f *File) CheckUpgradeFields(live *File, isRealm bool) error {
	if f.Mutable() && !isRealm {
		return fmt.Errorf("private or upgradeable packages must be realm packages")
	}
	if f.Version < 0 {
		return fmt.Errorf("version in gnomod.toml must not be negative")
	}
	// A version needs a way to redeploy, except on the redeploy that drops
	// the authority: that one carries the state and freezes the realm.
	if f.Version > 0 && !f.Mutable() && live == nil {
		return fmt.Errorf("version in gnomod.toml requires private = true or an [upgrade] authority")
	}
	if f.Upgradeable() {
		if f.Version == 0 {
			return fmt.Errorf("an [upgrade] authority in gnomod.toml requires version")
		}
		if _, err := crypto.AddressFromBech32(f.Upgrade.Authority); err != nil {
			return fmt.Errorf("invalid [upgrade] authority in gnomod.toml: %w", err)
		}
	}
	if live == nil {
		return nil
	}
	if live.Private && !f.Private {
		return fmt.Errorf("a private package cannot be overridden by a public package")
	}
	if !live.Private && f.Private {
		return fmt.Errorf("an upgradeable package cannot become private")
	}
	// Version unset on both sides is the old private redeploy, which starts
	// the globals over; stepping it by exactly one carries them.
	if (f.Version != 0 || live.Version != 0) && f.Version != live.Version+1 {
		return fmt.Errorf("redeploying %s needs version = %d in gnomod.toml, got %d",
			f.Module, live.Version+1, f.Version)
	}
	return nil
}
