package main

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// DevKeyName is the name under which gnodev auto-imports the well-known
// deployer mnemonic into the user's local keybase. The derived address is
// funded in the dev chain genesis, so signing against this name works against
// gnodev with no further setup.
const DevKeyName = "dev"

func setupAddressBook(logger *slog.Logger, cfg *AppConfig) (*address.Book, error) {
	book := address.NewBook()

	if err := ensureDevKey(logger, cfg); err != nil {
		return nil, err
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

	// With auto-import we usually hit this; --no-dev-key or no writable home
	// fall through to tracking the address in-memory only.
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

	return book, nil
}

// ensureDevKey writes the well-known deployer mnemonic into the user's
// local gnokey keybase under DevKeyName, unless opted out or already
// present under that name with a different address.
func ensureDevKey(logger *slog.Logger, cfg *AppConfig) error {
	if cfg.noDevKey {
		logger.Info("dev key skipped (--no-dev-key)")
		return nil
	}
	if cfg.home == "" {
		logger.Warn("dev key skipped: home not specified, cannot write to keybase")
		return nil
	}
	if !osm.DirExists(cfg.home) {
		// Default home (~/.config/gno) doesn't exist on fresh installs;
		// create it so the auto-import actually fires for first-time users,
		// matching `gnokey add`'s behavior. A user-supplied -home that
		// doesn't exist is likely a typo — refuse to materialize it.
		if cfg.home != gnoenv.HomeDir() {
			logger.Warn("dev key skipped: home directory does not exist", "path", cfg.home)
			return nil
		}
		if err := osm.EnsureDir(cfg.home, 0o700); err != nil {
			logger.Warn("dev key skipped: cannot create default home", "path", cfg.home, "err", err)
			return nil
		}
	}

	kb, err := keys.NewKeyBaseFromDir(cfg.home)
	if err != nil {
		return fmt.Errorf("unable to open keybase at %q: %w", cfg.home, err)
	}

	addr := defaultDeployerAddress.String()
	info, err := kb.GetByName(DevKeyName)
	switch {
	case err == nil:
		if info.GetAddress() == defaultDeployerAddress {
			logger.Info("dev key already present, skipping", "addr", addr)
			return nil
		}
		logger.Warn("dev key exists in keybase with a different address, not overwriting",
			"existing", info.GetAddress().String(),
			"expected", addr)
		return nil
	case keyerror.IsErrKeyNotFound(err):
	default:
		return fmt.Errorf("unable to read %q from keybase: %w", DevKeyName, err)
	}

	if _, err := kb.CreateAccount(DevKeyName, DefaultDeployerSeed, "", "", 0, 0); err != nil {
		return fmt.Errorf("unable to import dev key: %w", err)
	}
	logger.Info("dev key imported", "name", DevKeyName, "addr", addr)
	return nil
}
