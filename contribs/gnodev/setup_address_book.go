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

// DevKeyName is what the well-known deployer account is called once imported;
// the docs print that name in every sample.
const DevKeyName = "devtest"

func setupAddressBook(logger *slog.Logger, cfg *AppConfig) (*address.Book, error) {
	book := address.NewBook()

	// The `I` key runs the same import while gnodev is up.
	if cfg.importDevKey {
		importDevKey(logger, cfg.home)
	}

	// Check for home folder
	if cfg.home == "" {
		logger.Warn("home not specified, no keybase will be loaded")
	} else if !osm.DirExists(cfg.home) {
		logger.Warn("keybase directory does not exist, no local keys will be imported",
			"path", cfg.home)
	} else if err := book.ImportKeybase(cfg.home); err != nil {
		return nil, fmt.Errorf("unable to import local keybase %q: %w", cfg.home, err)
	}

	// Add additional accounts to our keybase
	for acc := range cfg.premineAccounts {
		if _, ok := book.GetByName(acc); ok {
			continue // we already know this account from keybase
		}

		// Check if we have a valid bech32 address instead
		addr, err := crypto.AddressFromBech32(acc)
		if err != nil {
			return nil, fmt.Errorf("invalid bech32 address or unknown keyname %q", acc)
		}

		book.Add(addr, "") // add addr to the book with no name

		logger.Info("additional account added", "addr", addr.String())
	}

	// Ensure that we have a default address
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

	// Nameless: the keybase holds no name for it, and logAccounts prints "_".
	book.Add(defaultDeployerAddress, "")

	// The mnemonic stays out of the log: it is the public DefaultDeployerSeed.
	logger.Warn("default address tracked in-memory only; gnokey cannot sign with it",
		"addr", defaultDeployerAddress.String(),
	)
	logger.Info("start with -import-dev-key to sign as it, or press I in interactive mode",
		"name", DevKeyName,
	)

	return book, nil
}

// importDevKey writes the well-known deployer mnemonic into the keybase at
// home and reports whether that keybase can sign for the deployer address.
// Every failure is a warning rather than an error, so a degraded keybase never
// stops gnodev from booting.
func importDevKey(logger *slog.Logger, home string) bool {
	if home == "" {
		logger.Warn("dev key skipped: home not specified, cannot write to keybase")
		return false
	}
	// A fresh install has no ~/.config/gno yet, and openKeybase creates it the
	// way `gnokey add` does; any other missing -home is a typo, never created.
	if !osm.DirExists(home) && filepath.Clean(home) != filepath.Clean(gnoenv.HomeDir()) {
		logger.Warn("dev key skipped: home directory does not exist", "path", home)
		return false
	}

	kb, err := openKeybase(home)
	if err != nil {
		logger.Warn("dev key skipped: cannot open keybase", "path", home, "err", err)
		return false
	}

	addr := defaultDeployerAddress.String()

	// One name per address in the keybase, so importing over an address the
	// user already holds would drop their own name for it, commonly `test1`.
	if has, err := kb.HasByAddress(defaultDeployerAddress); err != nil {
		logger.Warn("dev key skipped: cannot read keybase", "err", err)
		return false
	} else if has {
		logger.Info("dev key already present in keybase, skipping", "addr", addr)
		return true
	}

	// The name may belong to a key of the user's own; leave it untouched.
	switch info, err := kb.GetByName(DevKeyName); {
	case err == nil:
		logger.Warn("dev key name exists in keybase with a different address, not overwriting",
			"existing", info.GetAddress().String(),
			"expected", addr)
		return false
	case !keyerror.IsErrKeyNotFound(err):
		logger.Warn("dev key skipped: cannot read the name", "name", DevKeyName, "err", err)
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
