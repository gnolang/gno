package main

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

func setupAddressBook(logger *slog.Logger, cfg *AppConfig) (*address.Book, error) {
	book := address.NewBook()

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
			"addr", defaultDeployerAddress.String(),
			"note", "test key, every gnodev shares this address")
		return book, nil
	}

	book.Add(defaultDeployerAddress, "")

	logger.Warn("default address tracked in-memory only; gnokey cannot sign with it",
		"addr", defaultDeployerAddress.String(),
	)

	return book, nil
}
