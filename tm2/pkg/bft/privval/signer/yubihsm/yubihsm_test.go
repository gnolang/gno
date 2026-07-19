package yubihsm

import (
	"errors"
	"testing"

	"github.com/certusone/yubihsm-go/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getPubKeyCmdType and signCmdType are the CommandType values used by the
// real GetPubKey/SignDataEddsa commands, derived from the real constructors
// (rather than hardcoded protocol byte values) so the mock stays correct if
// the underlying wire format ever changes.
var (
	getPubKeyCmdType = mustCommandType(commands.CreateGetPubKeyCommand(1))
	signCmdType      = mustCommandType(commands.CreateSignDataEddsaCommand(1, []byte("x")))
)

func mustCommandType(cmd *commands.CommandMessage, err error) commands.CommandType {
	if err != nil {
		panic(err)
	}

	return cmd.CommandType
}

// mockSession is a fake implementing hsmAPI, used so tests never require a
// physical YubiHSM2 device or a running yubihsm-connector.
type mockSession struct {
	pubKeyData []byte
	signature  []byte

	sendErr   error
	destroyed bool
}

func (m *mockSession) SendEncryptedCommand(c *commands.CommandMessage) (commands.Response, error) {
	if m.sendErr != nil {
		return nil, m.sendErr
	}

	switch c.CommandType {
	case getPubKeyCmdType:
		return &commands.GetPubKeyResponse{KeyData: m.pubKeyData}, nil
	case signCmdType:
		return &commands.SignDataEddsaResponse{Signature: m.signature}, nil
	default:
		return nil, errors.New("mockSession: unexpected command type")
	}
}

func (m *mockSession) Destroy() {
	m.destroyed = true
}

func TestNewSigner_ValidDevice(t *testing.T) {
	t.Parallel()

	session := &mockSession{
		pubKeyData: make([]byte, 32),
		signature:  make([]byte, 64),
	}
	session.pubKeyData[0] = 0xAB
	session.signature[0] = 0xCD

	signer, err := newSigner(session, 42)
	require.NoError(t, err)
	require.NotNil(t, signer)

	assert.Equal(t, session.pubKeyData, signer.pubKey[:])
	assert.Contains(t, signer.String(), "YubiHSM2Signer")

	sig, err := signer.Sign([]byte("sign-bytes"))
	require.NoError(t, err)
	assert.Equal(t, session.signature, sig)

	assert.NoError(t, signer.Close())
	assert.True(t, session.destroyed)
}

func TestNewSigner_SendCommandError(t *testing.T) {
	t.Parallel()

	session := &mockSession{sendErr: errors.New("connector unreachable")}

	signer, err := newSigner(session, 42)
	require.Error(t, err)
	assert.Nil(t, signer)
}

func TestNewSigner_InvalidPubKeyLength(t *testing.T) {
	t.Parallel()

	session := &mockSession{pubKeyData: make([]byte, 16)} // wrong length

	signer, err := newSigner(session, 42)
	require.ErrorIs(t, err, errInvalidPubKeyLen)
	assert.Nil(t, signer)
}

func TestNewSigner_SignError(t *testing.T) {
	t.Parallel()

	session := &mockSession{
		pubKeyData: make([]byte, 32),
		sendErr:    nil,
	}

	signer, err := newSigner(session, 42)
	require.NoError(t, err)

	// Now make subsequent calls fail (the Sign call).
	session.sendErr = errors.New("device error")

	sig, err := signer.Sign([]byte("sign-bytes"))
	require.Error(t, err)
	assert.Nil(t, sig)
}

func TestConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	assert.False(t, (&Config{}).IsEnabled())
	assert.False(t, (*Config)(nil).IsEnabled())
	assert.True(t, (&Config{ConnectorURL: "http://127.0.0.1:12345"}).IsEnabled())
}

func TestConfig_ValidateBasic(t *testing.T) {
	t.Parallel()

	assert.NoError(t, (&Config{}).ValidateBasic())

	assert.ErrorIs(t, (&Config{
		ConnectorURL: "http://127.0.0.1:12345",
	}).ValidateBasic(), errZeroAuthKeyID)

	assert.ErrorIs(t, (&Config{
		ConnectorURL: "http://127.0.0.1:12345",
		AuthKeyID:    1,
	}).ValidateBasic(), errZeroKeyID)

	assert.NoError(t, (&Config{
		ConnectorURL: "http://127.0.0.1:12345",
		AuthKeyID:    1,
		KeyID:        2,
	}).ValidateBasic())
}

func TestNewSignerFromConfig_Disabled(t *testing.T) {
	t.Parallel()

	_, err := NewSignerFromConfig(DefaultConfig())
	require.ErrorIs(t, err, errDisabled)
}
