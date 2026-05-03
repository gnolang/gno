package privval

// upstream_config.go: configuration for the upstream-Tendermint-protocol
// listener mode. Used when the validator wants to accept inbound from
// tmkms (or Horcrux, or another upstream-protocol signer) instead of
// using a local signing key or dialing the gnokms native protocol.
//
// Operators set this in config.toml under [priv_validator.tmkms_listener].
// When ListenAddr is non-empty the validator binds, listens for the
// signer to dial in, and uses the upstream package's SignerClient
// (which holds tmkms's authority over HRS state).

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// Errors returned by TmkmsListenerConfig.
var (
	errInvalidTmkmsListenAddr      = errors.New("invalid tmkms_listener.listen_addr")
	errInvalidTmkmsAllowedPubkeys  = errors.New("invalid tmkms_listener.allowed_kms_pubkeys")
	errEmptyTmkmsChainID           = errors.New("tmkms_listener.chain_id must not be empty")
	errEmptyTmkmsAllowedPubkeys    = errors.New("tmkms_listener.allowed_kms_pubkeys must not be empty (an empty list accepts any peer that completes the SecretConnection handshake — set explicitly only for dev/test)")
	errUnsupportedProtocolVersion  = errors.New("tmkms_listener.protocol_version must match the supported upstream Tendermint privval dialect")
	errBothExternalSignersEnabled  = errors.New("only one of remote_signer or tmkms_listener may be configured")
)

// TmkmsListenerConfig configures the upstream-Tendermint-protocol listener.
// Field types and JSON/TOML tags chosen to match cometbft's config style
// for operator familiarity.
type TmkmsListenerConfig struct {
	// ListenAddr is the validator's listen address for inbound signer
	// connections. Format: "tcp://0.0.0.0:26659" or
	// "unix:///var/run/gnoland/privval.sock". An empty value disables
	// this mode (the existing local/remote-client paths apply).
	ListenAddr string `json:"listen_addr" toml:"listen_addr" comment:"Address to listen on for upstream-protocol signer connections (e.g. tmkms). Empty to disable."`

	// AllowedKMSPubKeys is the allowlist of expected signer pubkeys.
	// Each entry is a hex-encoded ed25519 pubkey (32 bytes → 64 hex
	// chars), optionally with an "ed25519:" prefix. An empty list
	// accepts any peer that completes the SecretConnection handshake —
	// fail-open mode, dev/test only.
	AllowedKMSPubKeys []string `json:"allowed_kms_pubkeys" toml:"allowed_kms_pubkeys" comment:"Allowlist of expected signer pubkeys (hex-encoded ed25519). Empty = accept any peer (dev only)."`

	// ChainID is sent in PubKeyRequest/SignVoteRequest/SignProposalRequest.
	// tmkms verifies its configured chain_id matches, refusing to sign
	// for a different network. Required for production; equivalent to
	// the chain_id field in tmkms's [[validator]] block.
	ChainID string `json:"chain_id" toml:"chain_id" comment:"Chain ID sent to the signer; must match tmkms.toml's [[validator]] chain_id."`

	// ProtocolVersion pins the upstream Tendermint privval dialect.
	// Today only "v0.34" is supported — the canonical sign-bytes in
	// upstreampb are wired to v0.34's protobuf shape. Mirrors tmkms's
	// [[validator]].protocol_version; both sides MUST agree. We refuse
	// any other value at startup rather than silently misencode votes.
	ProtocolVersion string `json:"protocol_version" toml:"protocol_version" comment:"Upstream Tendermint privval dialect to speak (only \"v0.34\" supported). Must match tmkms.toml's [[validator]] protocol_version."`

	// TimeoutReadWrite is the read/write deadline applied to the held
	// signer connection. Default 5s, matching cometbft.
	TimeoutReadWrite time.Duration `json:"timeout_read_write" toml:"timeout_read_write" comment:"Read/write deadline for signer connections (default 5s)."`

	// WaitForConnectionTimeout caps how long Init() blocks waiting for
	// the signer to dial in at startup. Default 60s.
	WaitForConnectionTimeout time.Duration `json:"wait_for_connection_timeout" toml:"wait_for_connection_timeout" comment:"Max time to wait for signer to dial in at startup (default 60s)."`

	// Retries is the per-Sign retry budget on transient errors. Zero
	// means retry forever (matches cometbft's RetrySignerClient
	// convention). Default 5.
	Retries int `json:"retries" toml:"retries" comment:"Per-Sign retry attempts on transient errors. 0 = infinite (default 5)."`

	// RetryTimeout is the sleep between retry attempts. Default 1s.
	RetryTimeout time.Duration `json:"retry_timeout" toml:"retry_timeout" comment:"Sleep between retry attempts (default 1s)."`
}

// DefaultTmkmsListenerConfig returns a config with reasonable defaults
// but ListenAddr unset (so the mode is OFF by default — operator must
// opt in). AllowedKMSPubKeys is an empty slice (not nil) so TOML round-
// trip is byte-stable; matches the convention in
// rsclient.DefaultRemoteSignerClientConfig.
func DefaultTmkmsListenerConfig() *TmkmsListenerConfig {
	return &TmkmsListenerConfig{
		ListenAddr:               "",
		AllowedKMSPubKeys:        []string{},
		ChainID:                  "",
		ProtocolVersion:          upstream.ProtocolVersion,
		TimeoutReadWrite:         5 * time.Second,
		WaitForConnectionTimeout: 60 * time.Second,
		Retries:                  5,
		RetryTimeout:             1 * time.Second,
	}
}

// IsEnabled reports whether the listener mode is configured (i.e.,
// ListenAddr is set). Used by NewPrivValidatorFromConfig to decide
// which factory branch to take.
func (c *TmkmsListenerConfig) IsEnabled() bool {
	return c != nil && c.ListenAddr != ""
}

// ValidateBasic is invoked by the parent PrivValidatorConfig.ValidateBasic.
// Only checks fields when the mode is enabled.
//
// Refuses an empty AllowedKMSPubKeys when the mode is enabled. The
// underlying TCPListener treats an empty allowlist as "accept any peer
// that completes the SecretConnection handshake" — useful for dev/test
// but a footgun in production: a misconfigured firewall plus an attacker
// who can mint an ed25519 keypair would be enough to substitute the
// signer. We force the operator to put their tmkms identity in the
// allowlist explicitly.
func (c *TmkmsListenerConfig) ValidateBasic() error {
	if !c.IsEnabled() {
		return nil
	}
	if c.ChainID == "" {
		return errEmptyTmkmsChainID
	}
	if len(c.AllowedKMSPubKeys) == 0 {
		return errEmptyTmkmsAllowedPubkeys
	}
	if _, err := c.ParseAllowlist(); err != nil {
		return fmt.Errorf("%w: %v", errInvalidTmkmsAllowedPubkeys, err)
	}
	if c.ProtocolVersion != upstream.ProtocolVersion {
		return fmt.Errorf("%w: got %q, supported: %q",
			errUnsupportedProtocolVersion, c.ProtocolVersion, upstream.ProtocolVersion)
	}
	return nil
}

// ParseAllowlist decodes AllowedKMSPubKeys into typed ed25519 pubkeys.
// An entry may be a bare 64-hex-char string or prefixed with "ed25519:".
// An empty list returns nil — the caller treats nil as "accept any".
func (c *TmkmsListenerConfig) ParseAllowlist() ([]ed25519.PubKeyEd25519, error) {
	if len(c.AllowedKMSPubKeys) == 0 {
		return nil, nil
	}
	out := make([]ed25519.PubKeyEd25519, 0, len(c.AllowedKMSPubKeys))
	for i, raw := range c.AllowedKMSPubKeys {
		s := strings.TrimPrefix(raw, "ed25519:")
		bz, err := hex.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("entry %d: hex decode: %w", i, err)
		}
		if len(bz) != ed25519.PubKeyEd25519Size {
			return nil, fmt.Errorf("entry %d: pubkey length %d, expected %d", i, len(bz), ed25519.PubKeyEd25519Size)
		}
		var pk ed25519.PubKeyEd25519
		copy(pk[:], bz)
		out = append(out, pk)
	}
	return out, nil
}
