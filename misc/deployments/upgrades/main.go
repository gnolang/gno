// Command upgrades maintains a chain's upgrade ledger: the upgrades.json under
// misc/deployments/<chain>/ and the table in UPGRADES.md rendered from it.
// The rules live in gno.land/pkg/upgrades; this is the thin command the
// release tooling and CI call.
//
//	go run ./misc/deployments/upgrades validate misc/deployments/mainnet.gno.land
//	go run ./misc/deployments/upgrades render   misc/deployments/mainnet.gno.land
//	go run ./misc/deployments/upgrades check    misc/deployments/mainnet.gno.land
//	go run ./misc/deployments/upgrades has      misc/deployments/mainnet.gno.land v1.6.0
package main

import (
	"fmt"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/upgrades"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: upgrades <validate|render|check> <dir> | has <dir> <version>")
	}
	cmd, dir := args[0], args[1]
	switch cmd {
	case "validate":
		l, err := upgrades.Load(dir)
		if err != nil {
			return err
		}
		if err := l.Validate(); err != nil {
			return err
		}
		fmt.Printf("%s/%s: %d entries, valid\n", dir, upgrades.LedgerFile, len(l.Upgrades))
	case "render":
		if err := upgrades.Render(dir); err != nil {
			return err
		}
		fmt.Printf("%s/%s rendered from %s\n", dir, upgrades.DocFile, upgrades.LedgerFile)
	case "check":
		if err := upgrades.Check(dir); err != nil {
			return err
		}
		fmt.Printf("%s: ledger valid, %s up to date\n", dir, upgrades.DocFile)
	case "has":
		if len(args) != 3 {
			return fmt.Errorf("usage: upgrades has <dir> <version>")
		}
		l, err := upgrades.Load(dir)
		if err != nil {
			return err
		}
		if !l.Has(args[2]) {
			return fmt.Errorf("%s/%s has no entry for %s", dir, upgrades.LedgerFile, args[2])
		}
		fmt.Printf("%s/%s has an entry for %s\n", dir, upgrades.LedgerFile, args[2])
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
	return nil
}
