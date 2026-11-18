// Package migrations runs the state migrations of coordinated upgrades: the
// code registered under the version a halt proposal names in
// halt_min_version runs once, in the block after the halt
// (gno.land/adr/prxxxx_inplace_state_migration.md).
package migrations

import (
	"fmt"
	"slices"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Env is everything a handler may reach.
type Env struct {
	Params   params.ParamsKeeperI
	Account  auth.AccountKeeperI
	Bank     bank.BankKeeperI
	GasPrice auth.GasPriceKeeperI
	VM       *vm.VMKeeper

	// Store keys, for what the keepers do not expose, such as walking the VM
	// object graph in ctx.Store(BaseKey).
	MainKey store.StoreKey
	BaseKey store.StoreKey
}

// Handler performs one upgrade's state change. It runs inside the block after
// the halt, on every node, so it must be deterministic and bounded. An error
// stops the node at that block.
type Handler func(ctx sdk.Context, env Env) error

// Upgrade is the migration of one release.
type Upgrade struct {
	Version string   // the release tag; what halt_min_version carries
	Stdlibs []string // stdlibs the release adds, loaded before Handler runs
	Handler Handler  // the code of the migration; optional with Stdlibs
}

// Stdlibs returns the stdlibs every upgrade adds.
func Stdlibs(upgrades []Upgrade) []string {
	var paths []string
	for _, u := range upgrades {
		paths = append(paths, u.Stdlibs...)
	}
	return paths
}

// Find returns the upgrade registered under version.
func Find(upgrades []Upgrade, version string) (Upgrade, bool) {
	for _, u := range upgrades {
		if u.Version == version {
			return u, true
		}
	}
	return Upgrade{}, false
}

// Validate checks that every upgrade names a version and does something, that
// no version is registered twice, and that each stdlib added is a stdlib of
// this binary, added once.
func Validate(upgrades []Upgrade) error {
	seen := make(map[string]bool, len(upgrades))
	added := make(map[string]bool)
	for i, u := range upgrades {
		switch {
		case u.Version == "":
			return fmt.Errorf("upgrade %d: empty version", i)
		case u.Handler == nil && len(u.Stdlibs) == 0:
			return fmt.Errorf("upgrade %q: no handler and no stdlib", u.Version)
		case seen[u.Version]:
			return fmt.Errorf("upgrade %q: registered twice", u.Version)
		}
		seen[u.Version] = true
		for _, lib := range u.Stdlibs {
			switch {
			case !slices.Contains(stdlibs.InitOrder(), lib):
				return fmt.Errorf("upgrade %q: %q is not a stdlib", u.Version, lib)
			case added[lib]:
				return fmt.Errorf("upgrade %q: stdlib %q added twice", u.Version, lib)
			}
			added[lib] = true
		}
	}
	return nil
}
