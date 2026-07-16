package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"
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

// RFC 3986 scheme grammar. It cannot contain ":", so a registry entry can
// never smuggle a payload (e.g. "javascript:...") into the launch link the
// frontend assigns to window.location.
var walletSchemeRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*$`)

func init() {
	if err := json.Unmarshal(walletsRaw, &wallets); err != nil {
		panic("unable to parse embedded wallet registry: " + err.Error())
	}
	if err := validateWallets(wallets); err != nil {
		panic("invalid embedded wallet registry: " + err.Error())
	}
	raw, err := json.Marshal(wallets)
	if err != nil {
		panic("unable to marshal wallet registry: " + err.Error())
	}
	walletsMarshalled = template.JS(raw) //nolint:gosec // JSON object intended for <script type="application/json"> embed
}

func validateWallets(ws []Wallet) error {
	seenIDs := make(map[string]bool, len(ws))
	seenSchemes := make(map[string]bool, len(ws))
	for _, w := range ws {
		switch {
		case w.Name == "" || w.ID == "":
			return fmt.Errorf("entry %q/%q: name and id are required", w.Name, w.ID)
		case !walletSchemeRe.MatchString(w.Scheme):
			return fmt.Errorf("wallet %q: scheme %q is not a valid URL scheme", w.ID, w.Scheme)
		case !strings.HasPrefix(w.Icon, "data:image/"):
			return fmt.Errorf("wallet %q: icon must be a data:image/ URI", w.ID)
		case seenIDs[w.ID]:
			return fmt.Errorf("duplicate wallet id %q", w.ID)
		case seenSchemes[w.Scheme]:
			return fmt.Errorf("duplicate wallet scheme %q", w.Scheme)
		}
		seenIDs[w.ID] = true
		seenSchemes[w.Scheme] = true
	}
	return nil
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
