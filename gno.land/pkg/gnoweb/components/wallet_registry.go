package components

import (
	_ "embed"
	"encoding/json"
	"html/template"
)

// Wallet is one entry of the embedded external-wallet registry (wallets.json),
// describing an app that handles GnoConnect launch links ("<scheme>://tx?...").
type Wallet struct {
	Name       string   `json:"name"`
	ID         string   `json:"id"`
	Icon       string   `json:"icon"`        // data: URI (offline-safe)
	Scheme     string   `json:"scheme"`      // bare URL scheme, e.g. "land.gno.gnokey"; the frontend appends "://tx?..."
	Platforms  []string `json:"platforms"`   // informational for now
	InstallURL string   `json:"install_url"` // informational for now
}

//go:embed wallets.json
var walletsRaw []byte

// Parsed and re-marshaled once at init (json.Marshal HTML-escapes, the raw
// file may not). A malformed registry is an authoring error, so panic.
var (
	wallets           []Wallet
	walletsMarshalled template.JS
)

func init() {
	if err := json.Unmarshal(walletsRaw, &wallets); err != nil {
		panic("unable to parse embedded wallet registry: " + err.Error())
	}
	raw, err := json.Marshal(wallets)
	if err != nil {
		panic("unable to marshal wallet registry: " + err.Error())
	}
	walletsMarshalled = template.JS(raw)
}

// Wallets returns the embedded external-wallet registry.
func Wallets() []Wallet {
	return wallets
}

// WalletsJSON returns the registry pre-marshaled for the frontend's
// <script type="application/json"> tag.
func WalletsJSON() template.JS {
	return walletsMarshalled
}
