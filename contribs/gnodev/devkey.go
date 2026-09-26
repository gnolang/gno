package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// DevKeyName is what the well-known deployer account is called once imported;
// the docs print that name in every sample.
const DevKeyName = "devtest"

// errDevKeyPresent reports that the keybase already signs for the deployer
// address, under whatever name its owner gave it.
var errDevKeyPresent = errors.New("the keybase already holds this address")

// askDevKey offers the import before gnodev prints anything else, and writes
// nothing unless the answer starts with y. It runs before the terminal goes
// raw, so the answer is an ordinary line and Enter declines.
func askDevKey(io commands.IO, home string) {
	if home == "" || devKeySignable(home) {
		return
	}

	answer, err := io.GetString(fmt.Sprintf(
		"gnodev has a test account %s and your keybase holds no key for it.\n"+
			"Add its key as %q? [y/N]", defaultDeployerAddress, DevKeyName))
	if err != nil || !strings.HasPrefix(strings.ToLower(answer), "y") {
		return
	}

	switch created, err := importDevKey(home); {
	case err != nil:
		io.ErrPrintfln("dev key not added: %v", err)
	case created:
		io.ErrPrintfln("dev key added as %q, address %s", DevKeyName, defaultDeployerAddress)
	}
}

// devKeySignable reports whether the keybase at home already holds a key for
// the deployer address, so a reader who has one is never asked again.
func devKeySignable(home string) bool {
	if !osm.DirExists(home) {
		return false
	}

	kb, err := openKeybase(home)
	if err != nil {
		return false
	}
	has, err := kb.HasByAddress(defaultDeployerAddress)
	return err == nil && has
}

// importDevKey writes the well-known deployer mnemonic into the keybase at home
// and reports whether this call created the entry. It refuses rather than
// replacing anything the user holds, and its errors are for a caller that
// carries on booting.
func importDevKey(home string) (bool, error) {
	if home == "" {
		return false, errors.New("no home directory to write a keybase to")
	}
	// A fresh install has no ~/.config/gno yet, and openKeybase creates it the
	// way `gnokey add` does; any other missing -home is a typo, never created.
	if !osm.DirExists(home) && filepath.Clean(home) != filepath.Clean(gnoenv.HomeDir()) {
		return false, fmt.Errorf("home directory %q does not exist", home)
	}

	kb, err := openKeybase(home)
	if err != nil {
		return false, err
	}

	// One name per address in the keybase, so importing over an address the
	// user already holds would drop their own name for it, commonly `test1`.
	switch has, err := kb.HasByAddress(defaultDeployerAddress); {
	case err != nil:
		return false, fmt.Errorf("cannot read the keybase: %w", err)
	case has:
		return false, errDevKeyPresent
	}

	// The name may belong to a key of the user's own; leave it untouched.
	switch info, err := kb.GetByName(DevKeyName); {
	case err == nil:
		return false, fmt.Errorf("%q already names %s in the keybase", DevKeyName, info.GetAddress())
	case !keyerror.IsErrKeyNotFound(err):
		return false, fmt.Errorf("cannot read %q in the keybase: %w", DevKeyName, err)
	}

	if _, err := kb.CreateAccount(DevKeyName, DefaultDeployerSeed, "", "", 0, 0); err != nil {
		return false, fmt.Errorf("cannot write the key: %w", err)
	}
	return true, nil
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
