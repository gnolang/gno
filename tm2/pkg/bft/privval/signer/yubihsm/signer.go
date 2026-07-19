package yubihsm

import (
	"errors"
	"fmt"

	"github.com/certusone/yubihsm-go/commands"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// Signer implements types.Signer by delegating public key retrieval and
// signing to an Ed25519 key stored in a YubiHSM2 hardware security module.
// Unlike the file/AWS/GCP/Vault-backed signers, the private key never
// leaves the device: PubKey is fetched once at construction time and
// cached, and Sign sends the raw sign-bytes to the device (through the
// yubihsm-connector) for on-device signing.
type Signer struct {
	session hsmAPI
	keyID   uint16
	pubKey  ed25519.PubKeyEd25519
}

// Signer type implements types.Signer.
var _ types.Signer = (*Signer)(nil)

// PubKey implements types.Signer.
func (s *Signer) PubKey() crypto.PubKey {
	return s.pubKey
}

// Sign implements types.Signer.
func (s *Signer) Sign(signBytes []byte) ([]byte, error) {
	cmd, err := commands.CreateSignDataEddsaCommand(s.keyID, signBytes)
	if err != nil {
		return nil, fmt.Errorf("yubihsm signer: unable to build sign command: %w", err)
	}

	resp, err := s.session.SendEncryptedCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("yubihsm signer: sign failed: %w", err)
	}

	sigResp, ok := resp.(*commands.SignDataEddsaResponse)
	if !ok {
		return nil, errUnexpectedResponseType
	}

	return sigResp.Signature, nil
}

// Close implements types.Signer. It tears down the session with the device.
func (s *Signer) Close() error {
	s.session.Destroy()

	return nil
}

// Signer type implements fmt.Stringer.
var _ fmt.Stringer = (*Signer)(nil)

// String implements fmt.Stringer.
func (s *Signer) String() string {
	return fmt.Sprintf("{Type: YubiHSM2Signer, Addr: %s}", s.pubKey.Address())
}

// Config validation errors.
var (
	errDisabled               = errors.New("yubihsm signer: not enabled")
	errInvalidPubKeyLen       = errors.New("yubihsm signer: device returned an unexpected public key length")
	errUnexpectedResponseType = errors.New("yubihsm signer: unexpected response type from device")
)

// NewSignerFromConfig opens a session with the YubiHSM2 device behind cfg's
// yubihsm-connector and returns a ready-to-use Signer bound to cfg.KeyID.
func NewSignerFromConfig(cfg *Config) (*Signer, error) {
	if !cfg.IsEnabled() {
		return nil, errDisabled
	}

	session, err := newSession(cfg)
	if err != nil {
		return nil, err
	}

	return newSigner(session, cfg.KeyID)
}

// newSigner contains the constructor logic decoupled from the concrete
// YubiHSM2 session, so it can be exercised in tests against a mock hsmAPI
// without requiring a physical device.
func newSigner(session hsmAPI, keyID uint16) (*Signer, error) {
	cmd, err := commands.CreateGetPubKeyCommand(keyID)
	if err != nil {
		return nil, fmt.Errorf("yubihsm signer: unable to build get-pubkey command: %w", err)
	}

	resp, err := session.SendEncryptedCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("yubihsm signer: unable to retrieve public key: %w", err)
	}

	pubKeyResp, ok := resp.(*commands.GetPubKeyResponse)
	if !ok {
		return nil, errUnexpectedResponseType
	}

	if len(pubKeyResp.KeyData) != ed25519.PubKeyEd25519Size {
		return nil, errInvalidPubKeyLen
	}

	var pubKey ed25519.PubKeyEd25519
	copy(pubKey[:], pubKeyResp.KeyData)

	return &Signer{session: session, keyID: keyID, pubKey: pubKey}, nil
}
