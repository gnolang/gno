package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"slices"
	"strings"
)

// WalletKind says how a registry entry is reached: an extension runs in the
// page and announces itself, an app is reached by GnoConnect launch link.
type WalletKind string

const (
	WalletKindExtension WalletKind = "extension"
	WalletKindApp       WalletKind = "app"
)

// InstallLink is one place a wallet can be obtained, per store or repository.
type InstallLink struct {
	Label string `json:"label"`
	URL   string `json:"url"` // https only
}

// Wallet is one entry of the embedded wallet registry (wallets.json).
type Wallet struct {
	Name      string        `json:"name"`
	ID        string        `json:"id"`
	Icon      string        `json:"icon"`   // data: URI (offline-safe)
	RDNS      string        `json:"rdns"`   // durable identity; matches an announcement
	Kind      WalletKind    `json:"kind"`   // extension | app
	Scheme    string        `json:"scheme"` // bare URL scheme; required for an app, empty otherwise
	Global    string        `json:"global"` // window key of a legacy extension, if it has one
	Platforms []string      `json:"platforms"`
	Install   []InstallLink `json:"install"`
}

// WalletPlatforms is the accepted `platforms` set, in the order the install
// page groups them.
var WalletPlatforms = []string{"ios", "android", "chrome", "firefox", "brave", "edge", "safari"}

var walletPlatformLabels = map[string]string{
	"ios":     "iOS",
	"android": "Android",
	"chrome":  "Chrome",
	"firefox": "Firefox",
	"brave":   "Brave",
	"edge":    "Edge",
	"safari":  "Safari",
}

// WalletPlatformLabel returns the display label for a platform key, or the key
// itself when it has none.
func WalletPlatformLabel(platform string) string {
	if label, ok := walletPlatformLabels[platform]; ok {
		return label
	}
	return platform
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

// A bare JS identifier: the frontend probes it as window[global].
var walletGlobalRe = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

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
	seenRDNS := make(map[string]bool, len(ws))
	seenSchemes := make(map[string]bool, len(ws))
	for _, w := range ws {
		if err := validateWallet(w); err != nil {
			return err
		}
		switch {
		case seenIDs[w.ID]:
			return fmt.Errorf("duplicate wallet id %q", w.ID)
		case seenRDNS[w.RDNS]:
			return fmt.Errorf("duplicate wallet rdns %q", w.RDNS)
		// Skip empties: two scheme-less extensions would otherwise collide.
		case w.Scheme != "" && seenSchemes[w.Scheme]:
			return fmt.Errorf("duplicate wallet scheme %q", w.Scheme)
		}
		seenIDs[w.ID] = true
		seenRDNS[w.RDNS] = true
		if w.Scheme != "" {
			seenSchemes[w.Scheme] = true
		}
	}
	return nil
}

func validateWallet(w Wallet) error {
	switch {
	case w.Name == "" || w.ID == "" || w.RDNS == "":
		return fmt.Errorf("entry %q/%q: name, id and rdns are required", w.Name, w.ID)
	case w.Kind != WalletKindExtension && w.Kind != WalletKindApp:
		return fmt.Errorf("wallet %q: kind must be %q or %q, got %q", w.ID, WalletKindExtension, WalletKindApp, w.Kind)
	case w.Kind == WalletKindApp && w.Scheme == "":
		return fmt.Errorf("wallet %q: an app entry requires a scheme", w.ID)
	case w.Scheme != "" && !walletSchemeRe.MatchString(w.Scheme):
		return fmt.Errorf("wallet %q: scheme %q is not a valid URL scheme", w.ID, w.Scheme)
	case w.Global != "" && w.Kind != WalletKindExtension:
		return fmt.Errorf("wallet %q: only an extension may declare a global", w.ID)
	case w.Global != "" && !walletGlobalRe.MatchString(w.Global):
		return fmt.Errorf("wallet %q: global %q is not a valid global name", w.ID, w.Global)
	case !strings.HasPrefix(w.Icon, "data:image/"):
		return fmt.Errorf("wallet %q: icon must be a data:image/ URI", w.ID)
	case len(w.Platforms) == 0:
		return fmt.Errorf("wallet %q: at least one platform is required", w.ID)
	case len(w.Install) == 0:
		return fmt.Errorf("wallet %q: at least one install link is required", w.ID)
	}
	for _, p := range w.Platforms {
		if !slices.Contains(WalletPlatforms, p) {
			return fmt.Errorf("wallet %q: unknown platform %q", w.ID, p)
		}
	}
	for _, link := range w.Install {
		if link.Label == "" {
			return fmt.Errorf("wallet %q: install link needs a label", w.ID)
		}
		if !strings.HasPrefix(link.URL, "https://") {
			return fmt.Errorf("wallet %q: install URL must be https, got %q", w.ID, link.URL)
		}
	}
	return nil
}

// Wallets returns the embedded wallet registry.
func Wallets() []Wallet {
	return wallets
}

// WalletsJSON returns the registry pre-marshaled for the frontend's
// <script type="application/json"> tag.
func WalletsJSON() template.JS {
	return walletsMarshalled
}
