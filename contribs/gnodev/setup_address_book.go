package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// DevKeyName is the name under which gnodev imports the well-known deployer
// mnemonic into the user's local keybase, on request. The derived address is
// funded in the dev chain genesis, so signing against this name works as soon
// as the key is there.
const DevKeyName = "devtest"

func setupAddressBook(logger *slog.Logger, cfg *AppConfig) (*address.Book, error) {
	book := address.NewBook()

	// Only on request: gnodev writes to the user's keybase when asked, by
	// -import-dev-key here or by the `I` key while it runs. Best-effort
	// either way, so a degraded keybase never stops gnodev from booting.
	if cfg.importDevKey {
		importDevKey(logger, cfg.home)
	}

	if cfg.home == "" {
		logger.Warn("home not specified, no keybase will be loaded")
	} else if !osm.DirExists(cfg.home) {
		logger.Warn("keybase directory does not exist, no local keys will be imported",
			"path", cfg.home)
	} else if err := book.ImportKeybase(cfg.home); err != nil {
		return nil, fmt.Errorf("unable to import local keybase %q: %w", cfg.home, err)
	}

	for acc := range cfg.premineAccounts {
		if _, ok := book.GetByName(acc); ok {
			continue
		}

		addr, err := crypto.AddressFromBech32(acc)
		if err != nil {
			return nil, fmt.Errorf("invalid bech32 address or unknown keyname %q", acc)
		}

		book.Add(addr, "")

		logger.Info("additional account added", "addr", addr.String())
	}

	// Reached once the key is in the keybase, by an earlier import or by the
	// user's own `gnokey add -recover`; otherwise the address is tracked
	// in-memory only.
	if names, ok := book.GetByAddress(defaultDeployerAddress); ok {
		var name string
		if len(names) > 0 {
			name = names[0]
		}
		logger.Info("default address resolved from keybase",
			"name", name,
			"addr", defaultDeployerAddress.String())
		return book, nil
	}

	creatorName := fmt.Sprintf("_default#%.6s", defaultDeployerAddress.String())
	book.Add(defaultDeployerAddress, creatorName)

	// Mnemonic intentionally omitted: it's the public DefaultDeployerSeed
	// constant; users who need it can recover it via gnokey or the source.
	logger.Warn("default address tracked in-memory only; gnokey cannot sign with it",
		"name", creatorName,
		"addr", defaultDeployerAddress.String(),
	)
	logger.Info("press I to import it as a local key, or start with -import-dev-key",
		"name", DevKeyName,
	)

	return book, nil
}

// importDevKey writes the well-known deployer mnemonic into the user's local
// gnokey keybase under DevKeyName, unless the address is already there, or the
// name is taken by a different address. It reports whether the keybase can
// sign for that address when it returns.
//
// Every failure degrades to a logged warning rather than an error: the import
// is a convenience, and a missing, unwritable, locked, or corrupt keybase must
// never prevent gnodev from starting. The deployer address is still tracked
// in-memory by setupAddressBook's fallback when the import is skipped.
func importDevKey(logger *slog.Logger, home string) bool {
	if home == "" {
		logger.Warn("dev key skipped: home not specified, cannot write to keybase")
		return false
	}
	if !osm.DirExists(home) {
		// Default home (~/.config/gno) doesn't exist on fresh installs;
		// create it so a requested import actually lands, matching
		// `gnokey add`'s behavior. A user-supplied -home that doesn't exist
		// is likely a typo, so refuse to materialize it. Clean both paths so
		// a path-equivalent -home (e.g. a trailing slash) still counts as the
		// default.
		if filepath.Clean(home) != filepath.Clean(gnoenv.HomeDir()) {
			logger.Warn("dev key skipped: home directory does not exist", "path", home)
			return false
		}
		if err := osm.EnsureDir(home, 0o700); err != nil {
			logger.Warn("dev key skipped: cannot create default home", "path", home, "err", err)
			return false
		}
	}

	kb, err := openKeybase(home)
	if err != nil {
		logger.Warn("dev key skipped: cannot open keybase", "path", home, "err", err)
		return false
	}

	addr := defaultDeployerAddress.String()

	// If the deployer address is already in the keybase under any name, it is
	// already signable; do not add a second name. The keybase enforces one
	// name per address, so importing `devtest` here would silently drop the
	// user's existing entry (commonly `test1`).
	if has, err := kb.HasByAddress(defaultDeployerAddress); err != nil {
		logger.Warn("dev key skipped: cannot read keybase", "err", err)
		return false
	} else if has {
		logger.Info("dev key already present in keybase, skipping", "addr", addr)
		return true
	}

	// The address is not present, but the name `devtest` might belong to an
	// unrelated key. Leave any such entry untouched.
	switch info, err := kb.GetByName(DevKeyName); {
	case err == nil:
		logger.Warn("dev key name exists in keybase with a different address, not overwriting",
			"existing", info.GetAddress().String(),
			"expected", addr)
		return false
	case keyerror.IsErrKeyNotFound(err):
	default:
		logger.Warn("dev key skipped: cannot read keybase", "name", DevKeyName, "err", err)
		return false
	}

	if _, err := kb.CreateAccount(DevKeyName, DefaultDeployerSeed, "", "", 0, 0); err != nil {
		logger.Warn("dev key skipped: import failed", "err", err)
		return false
	}
	logger.Info("dev key imported", "name", DevKeyName, "addr", addr)
	return true
}

// openKeybase opens (creating the data dir if needed) the keybase at home.
// keys.NewKeyBaseFromDir panics instead of returning an error when it cannot
// create that dir (e.g. an unwritable home), so recover here and surface it as
// a normal error for the best-effort caller.
func openKeybase(home string) (kb keys.Keybase, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot initialize keybase: %v", r)
		}
	}()
	return keys.NewKeyBaseFromDir(home)
}
