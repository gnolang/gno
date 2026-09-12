package conn

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/async"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"golang.org/x/crypto/chacha20poly1305"
)

// EXPLORATION: scheme-agnostic SecretConnection variant.
//
// The original MakeSecretConnection in secret_connection.go is hard-typed
// to ed25519 because that is what the tm2 p2p transport assumes for node
// identities. The privval transport uses the same code but has no such
// constraint: the operator controls both ends and can pick any signing
// scheme they like.
//
// MakeSecretConnectionAny is a parallel entry point that accepts any
// crypto.PrivKey (currently ed25519 or secp256k1) and exchanges the auth
// signature using an amino-polymorphic message (authSigMessageAny). The
// DH + ChaCha20-Poly1305 wire layer is shared with the original path; only
// the final challenge-signature handshake differs.
//
// This intentionally diverges from the original on-wire format for the
// auth step. A peer using MakeSecretConnection cannot talk to a peer using
// MakeSecretConnectionAny. That is deliberate: privval is a private channel
// between an operator's signer and their validator node, never a p2p
// gossip peer, so the wire-break is contained.
//
// See tm2/adr/prxxxx_secp_validator.md.

// authSigMessageAny is the scheme-agnostic variant of authSigMessage.
// The Key field is amino-polymorphic so either ed25519.PubKeyEd25519 or
// secp256k1.PubKeySecp256k1 (or any future registered scheme) can travel
// on the wire.
type authSigMessageAny struct {
	Key crypto.PubKey
	Sig []byte
}

// MakeSecretConnectionAny is the scheme-agnostic counterpart to
// MakeSecretConnection. It accepts any crypto.PrivKey and authenticates
// the remote peer using an amino-polymorphic auth signature exchange.
//
// The returned *SecretConnection has its remPubKey set to a zero value
// (the field is concretely ed25519-typed for legacy reasons); call
// RemotePubKeyAny on the returned connection to retrieve the authenticated
// remote key as a crypto.PubKey.
func MakeSecretConnectionAny(conn io.ReadWriteCloser, locPrivKey crypto.PrivKey) (*SecretConnection, crypto.PubKey, error) {
	if locPrivKey == nil {
		return nil, nil, errors.New("local private key is nil")
	}
	locPubKey := locPrivKey.PubKey()

	// Generate ephemeral keys for perfect forward secrecy.
	locEphPub, locEphPriv := genEphKeys()

	// Write local ephemeral pubkey and receive one too.
	remEphPub, err := shareEphPubKey(conn, locEphPub)
	if err != nil {
		return nil, nil, err
	}

	// Sort by lexical order.
	loEphPub, _ := sort32(locEphPub, remEphPub)

	// Check if the local ephemeral public key was the least, lexicographically sorted.
	locIsLeast := bytes.Equal(locEphPub[:], loEphPub[:])

	// Compute common diffie hellman secret using X25519.
	dhSecret, err := computeDHSecret(remEphPub, locEphPriv)
	if err != nil {
		return nil, nil, err
	}

	// Generate per-direction AEAD keys and the challenge bytes.
	recvSecret, sendSecret, challenge := deriveSecretAndChallenge(dhSecret, locIsLeast)

	sendAead, err := chacha20poly1305.New(sendSecret[:])
	if err != nil {
		return nil, nil, errors.New("invalid send SecretConnection Key")
	}
	recvAead, err := chacha20poly1305.New(recvSecret[:])
	if err != nil {
		return nil, nil, errors.New("invalid receive SecretConnection Key")
	}

	// Construct SecretConnection. remPubKey is left at its zero value;
	// the authenticated key is returned separately via the second return
	// value so callers don't tangle with the legacy ed25519-typed field.
	sc := &SecretConnection{
		conn:       conn,
		recvBuffer: nil,
		recvNonce:  new([aeadNonceSize]byte),
		sendNonce:  new([aeadNonceSize]byte),
		recvAead:   recvAead,
		sendAead:   sendAead,
	}

	// Sign the challenge bytes for authentication.
	locSignature, err := locPrivKey.Sign(challenge[:])
	if err != nil {
		return nil, nil, fmt.Errorf("unable to sign challenge: %w", err)
	}

	// Share each other's pubkey & challenge signature, polymorphically.
	authMsg, err := shareAuthSignatureAny(sc, locPubKey, locSignature)
	if err != nil {
		return nil, nil, err
	}

	remPubKey, remSignature := authMsg.Key, authMsg.Sig
	if remPubKey == nil {
		return nil, nil, errors.New("remote did not send a public key")
	}

	if !remPubKey.VerifyBytes(challenge[:], remSignature) {
		return nil, nil, errors.New("challenge verification failed")
	}

	return sc, remPubKey, nil
}

func shareAuthSignatureAny(sc *SecretConnection, pubKey crypto.PubKey, signature []byte) (authSigMessageAny, error) {
	var recvMsg authSigMessageAny

	trs, _ := async.Parallel(
		func(_ int) (val any, err error, abort bool) {
			_, err1 := amino.MarshalSizedWriter(sc, authSigMessageAny{Key: pubKey, Sig: signature})
			if err1 != nil {
				return nil, err1, true
			}
			return nil, nil, false
		},
		func(_ int) (val any, err error, abort bool) {
			var msg authSigMessageAny
			_, err2 := amino.UnmarshalSizedReader(sc, &msg, 1024*1024)
			if err2 != nil {
				return nil, err2, true
			}
			return msg, nil, false
		},
	)

	if trs.FirstError() != nil {
		return recvMsg, trs.FirstError()
	}

	recvMsg = trs.FirstValue().(authSigMessageAny)
	return recvMsg, nil
}
