package privval

import (
	"context"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	rsclient "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/client"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// PrivValidatorConfig defines the configuration for the PrivValidator, with a local
// signer (file-based key), a gnokms remote signer (tm2-native protocol), or a tmkms
// listener (upstream Tendermint privval protocol — the validator listens for tmkms
// to dial in). At most one of RemoteSigner or TmkmsListener may be enabled.
type PrivValidatorConfig struct {
	// File path configuration.
	RootDir     string `json:"home" toml:"home"`
	SignState   string `json:"sign_state" toml:"sign_state" comment:"Path to the JSON file containing the last validator state to prevent double-signing"`
	LocalSigner string `json:"local_signer" toml:"local_signer" comment:"Path to the JSON file containing the private key to use for signing using a local signer"`

	// Remote Signer configuration (tm2-native protocol; validator dials gnokms).
	RemoteSigner *rsclient.RemoteSignerClientConfig `json:"remote_signer" toml:"remote_signer" comment:"Configuration for the remote signer client (gnokms)"`

	// TmkmsListener configures the upstream-Tendermint-protocol listener (validator
	// listens for tmkms / Horcrux to dial in). Mutually exclusive with RemoteSigner;
	// see upstream_config.go.
	TmkmsListener *TmkmsListenerConfig `json:"tmkms_listener" toml:"tmkms_listener" comment:"Configuration for upstream-Tendermint-protocol signer (tmkms / Horcrux). Empty listen_addr disables this mode."`
}

// PrivValidatorConfig validation errors.
var (
	errInvalidSignStatePath   = errors.New("invalid private validator sign state file path")
	errInvalidLocalSignerPath = errors.New("invalid private validator local signer file path")
	errNilRemoteSignerConfig  = errors.New("remote signer configuration cannot be nil")
)

// DefaultPrivValidatorConfig returns a default configuration for the PrivValidator.
func DefaultPrivValidatorConfig() *PrivValidatorConfig {
	return &PrivValidatorConfig{
		SignState:     "priv_validator_state.json",
		LocalSigner:   "priv_validator_key.json",
		RemoteSigner:  rsclient.DefaultRemoteSignerClientConfig(),
		TmkmsListener: DefaultTmkmsListenerConfig(),
	}
}

// TestPrivValidatorConfig returns a configuration for testing the PrivValidator.
func TestPrivValidatorConfig() *PrivValidatorConfig {
	return DefaultPrivValidatorConfig()
}

// SignStatePath returns the complete path for the sign state file.
func (cfg *PrivValidatorConfig) SignStatePath() string {
	return filepath.Join(cfg.RootDir, cfg.SignState)
}

// LocalSignerPath returns the complete path for the local signer file.
func (cfg *PrivValidatorConfig) LocalSignerPath() string {
	return filepath.Join(cfg.RootDir, cfg.LocalSigner)
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *PrivValidatorConfig) ValidateBasic() error {
	// Verify the validator sign state file path is set.
	if cfg.SignState == "" {
		return errInvalidSignStatePath
	}

	// Verify the validator local signer file path is set.
	if cfg.LocalSigner == "" {
		return errInvalidLocalSignerPath
	}

	// Verify the remote signer configuration is not nil.
	if cfg.RemoteSigner == nil {
		return errNilRemoteSignerConfig
	}

	// Validate the remote signer client configuration.
	if err := cfg.RemoteSigner.ValidateBasic(); err != nil {
		return err
	}

	// Validate the tmkms listener configuration if enabled.
	if cfg.TmkmsListener != nil {
		if err := cfg.TmkmsListener.ValidateBasic(); err != nil {
			return err
		}
	}

	// Mutual exclusion: at most one external-signer mode may be enabled.
	if cfg.RemoteSigner.ServerAddress != "" && cfg.TmkmsListener.IsEnabled() {
		return errBothExternalSignersEnabled
	}

	return nil
}

// NewSignerFromConfig returns a new Signer instance based on the configuration.
// The ctx and clientLogger are only used for the remote signer client.
// The clientPrivKey is only used for the remote signer client using a TCP connection.
func NewSignerFromConfig(
	ctx context.Context,
	config *PrivValidatorConfig,
	clientPrivKey ed25519.PrivKeyEd25519,
	clientLogger *slog.Logger,
) (types.Signer, error) {
	// Initialize the signer based on the configuration.
	// If the remote signer address is set, use a remote signer client.
	if config.RemoteSigner != nil && config.RemoteSigner.ServerAddress != "" {
		return rsclient.NewRemoteSignerClientFromConfig(
			ctx,
			config.RemoteSigner,
			clientPrivKey,
			clientLogger,
		)
	}

	// Otherwise, use a local signer.
	return local.LoadOrMakeLocalSigner(config.LocalSignerPath())
}

// NewPrivValidatorFromConfig returns a types.PrivValidator chosen by config:
//   - if TmkmsListener is enabled, build the upstream-protocol listener stack
//     (TCPListener → SignerListenerEndpoint → SignerClient → RetrySignerClient);
//   - otherwise return the existing local-or-remote-signer concrete *PrivValidator
//     (with FileState-backed HRS gating).
//
// Return type is the interface so callers don't need to special-case the
// listener path. The concrete *PrivValidator type is still accessible via type
// assertion for paths that need it (e.g., tests poking at the inner signer).
//
// The clientPrivKey is the validator's node identity key; it's used as the
// SecretConnection identity for both remote-signer-client and tmkms-listener
// modes.
func NewPrivValidatorFromConfig(
	config *PrivValidatorConfig,
	clientPrivKey ed25519.PrivKeyEd25519,
	clientLogger *slog.Logger,
) (types.PrivValidator, error) {
	// Mutual exclusion is also enforced in ValidateBasic, but defend in
	// depth here in case callers skip validation.
	if config.RemoteSigner != nil && config.RemoteSigner.ServerAddress != "" &&
		config.TmkmsListener.IsEnabled() {
		return nil, errBothExternalSignersEnabled
	}

	// tmkms-compat path. tmkms holds HRS authority; we don't wrap the
	// returned validator with a FileState — that'd be a redundant gate
	// at best and a misconfiguration footgun at worst.
	if config.TmkmsListener.IsEnabled() {
		return newTmkmsListenerPrivValidator(config.TmkmsListener, clientPrivKey, clientLogger)
	}

	// Local or remote-client signer path: existing flow, wrapped with
	// the in-tree FileState HRS gate.
	signer, err := NewSignerFromConfig(context.Background(), config, clientPrivKey, clientLogger)
	if err != nil {
		return nil, err
	}
	return NewPrivValidator(signer, config.SignStatePath())
}

// newTmkmsListenerPrivValidator wires the upstream package's listener
// stack and returns it as a types.PrivValidator. Init() blocks until
// the signer dials in (or WaitForConnectionTimeout elapses) — the
// validator's pubkey is fetched and cached during Init so subsequent
// PubKey() calls don't need network I/O.
func newTmkmsListenerPrivValidator(
	cfg *TmkmsListenerConfig,
	clientPrivKey ed25519.PrivKeyEd25519,
	logger *slog.Logger,
) (types.PrivValidator, error) {
	allowlist, err := cfg.ParseAllowlist()
	if err != nil {
		return nil, err
	}

	protocol, address := osm.ProtocolAndAddress(cfg.ListenAddr)
	rawLn, err := net.Listen(protocol, address)
	if err != nil {
		return nil, err
	}

	var compatLn net.Listener
	switch t := rawLn.(type) {
	case *net.TCPListener:
		compatLn = upstream.NewTCPListener(t, clientPrivKey, allowlist,
			upstream.TCPListenerTimeoutReadWrite(cfg.TimeoutReadWrite))
	case *net.UnixListener:
		// Default umask leaves the socket world-writable on most distros;
		// any local user could connect and (with an SecretConnection handshake
		// that we can't gate by allowlist on UDS) become the signer. Tighten
		// to owner-only RW. Best-effort: ignore the error (some filesystems
		// don't honor chmod on sockets) but log it.
		if err := os.Chmod(address, 0o600); err != nil {
			logger.Warn("tmkms_listener: chmod 0600 on unix socket failed; falling back to default perms",
				"path", address, "err", err)
		}
		compatLn = upstream.NewUnixListener(t,
			upstream.UnixListenerTimeoutReadWrite(cfg.TimeoutReadWrite))
	default:
		_ = rawLn.Close()
		return nil, errors.New("tmkms_listener: unsupported listener type %T", rawLn)
	}

	endpoint := upstream.NewSignerListenerEndpoint(logger, compatLn,
		upstream.SignerListenerEndpointTimeoutReadWrite(cfg.TimeoutReadWrite))

	sc, err := upstream.NewSignerClient(endpoint, cfg.ChainID)
	if err != nil {
		_ = endpoint.Stop()
		return nil, err
	}

	if err := sc.Init(cfg.WaitForConnectionTimeout); err != nil {
		// sc.Close() only drops the conn — the endpoint goroutines and
		// listener stay live and the port stays held. Stop the endpoint
		// to drain serviceLoop/pingLoop and close the listener.
		_ = endpoint.Stop()
		return nil, err
	}

	return upstream.NewRetrySignerClient(sc, cfg.Retries, cfg.RetryTimeout), nil
}
