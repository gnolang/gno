package components

import (
	_ "embed"
	"encoding/json"
)

// Wallet describes an external wallet application capable of handling a
// GnoConnect launch link (`<scheme>://tx?...`). The registry is embedded at
// build time so gnoweb renders the wallet chooser offline, with no fetch.
//
// The registry stores the bare URL scheme (e.g. "land.gno.gnokey"); the
// frontend composes the full launch prefix "<scheme>://tx?..." so the registry
// stays reusable if the standard later adds other hosts (run, sign, ...).
type Wallet struct {
	Name       string   `json:"name"`
	ID         string   `json:"id"`
	Icon       string   `json:"icon"`      // data: URI, self-contained (offline-safe)
	Scheme     string   `json:"scheme"`    // bare URL scheme, e.g. "land.gno.gnokey"
	Platforms  []string `json:"platforms"` // e.g. ["ios", "android"]
	InstallURL string   `json:"install_url"`
}

//go:embed wallets.json
var walletsJSON []byte

// wallets is the parsed, embedded external-wallet registry, loaded once at
// package init. A malformed registry is a build-time authoring error, so we
// panic rather than fail silently at request time.
var wallets = func() []Wallet {
	var w []Wallet
	if err := json.Unmarshal(walletsJSON, &w); err != nil {
		panic("unable to parse embedded wallet registry: " + err.Error())
	}
	return w
}()

// Wallets returns the embedded external-wallet registry.
func Wallets() []Wallet {
	return wallets
}
