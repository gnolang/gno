// Package migrations runs the state migrations of coordinated upgrades: the
// code registered under the version a halt proposal names in
// halt_min_version runs once, in the block after the halt
// (gno.land/adr/pr6507_inplace_state_migration.md).
package migrations

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
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
	Version string  // the release tag; what halt_min_version carries
	Handler Handler // the code of the migration
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

// Validate checks that every upgrade names a version and a handler, and that
// no version is registered twice.
func Validate(upgrades []Upgrade) error {
	seen := make(map[string]bool, len(upgrades))
	for i, u := range upgrades {
		switch {
		case u.Version == "":
			return fmt.Errorf("upgrade %d: empty version", i)
		case u.Handler == nil:
			return fmt.Errorf("upgrade %q: nil handler", u.Version)
		case seen[u.Version]:
			return fmt.Errorf("upgrade %q: registered twice", u.Version)
		}
		seen[u.Version] = true
	}
	return nil
}
