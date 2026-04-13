package ledger

import (
	"fmt"
	"math/big"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/internal/ledger"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

type (
	// PrivKeyLedgerSecp256k1 implements PrivKey, calling the ledger nano we
	// cache the PubKey from the first call to use it later.
	PrivKeyLedgerSecp256k1 struct {
		// CachedPubKey should be private, but we want to encode it via
		// go-amino so we can view the address later, even without having the
		// ledger attached.
		CachedPubKey crypto.PubKey
		Path         hd.BIP44Params
	}
)

// NewPrivKeyLedgerSecp256k1Unsafe will generate a new key and store the public key for later use.
//
// This function is marked as unsafe as it will retrieve a pubkey without user verification.
// It can only be used to verify a pubkey but never to create new accounts/keys. In that case,
// please refer to NewPrivKeyLedgerSecp256k1
func NewPrivKeyLedgerSecp256k1Unsafe(path hd.BIP44Params) (crypto.PrivKey, error) {
	device, err := getLedgerDevice()
	if err != nil {
		return nil, err
	}
	defer warnIfErrors(device.Close)

	pubKey, err := getPubKeyUnsafe(device, path)
	if err != nil {
		return nil, err
	}

	return PrivKeyLedgerSecp256k1{pubKey, path}, nil
}

// NewPrivKeyLedgerSecp256k1 will generate a new key and store the public key for later use.
// The request will require user confirmation and will show account and index in the device
func NewPrivKeyLedgerSecp256k1(path hd.BIP44Params, hrp string) (crypto.PrivKey, string, error) {
	device, err := getLedgerDevice()
	if err != nil {
		return nil, "", err
	}
	defer warnIfErrors(device.Close)

	pubKey, addr, err := getPubKeyAddrSafe(device, path, hrp)
	if err != nil {
		return nil, "", err
	}

	return PrivKeyLedgerSecp256k1{pubKey, path}, addr, nil
}

// PubKey returns the cached public key.
func (pkl PrivKeyLedgerSecp256k1) PubKey() crypto.PubKey {
	return pkl.CachedPubKey
}

// Sign returns a secp256k1 signature for the corresponding message
func (pkl PrivKeyLedgerSecp256k1) Sign(message []byte) ([]byte, error) {
	device, err := getLedgerDevice()
	if err != nil {
		return nil, err
	}
	defer warnIfErrors(device.Close)

	return sign(device, pkl, message)
}

// LedgerShowAddress triggers a ledger device to show the corresponding address.
func LedgerShowAddress(path hd.BIP44Params, expectedPubKey crypto.PubKey) error {
	device, err := getLedgerDevice()
	if err != nil {
		return err
	}
	defer warnIfErrors(device.Close)

	pubKey, err := getPubKeyUnsafe(device, path)
	if err != nil {
		return err
	}

	if pubKey != expectedPubKey {
		return fmt.Errorf("the key's pubkey does not match with the one retrieved from Ledger. Check that the HD path and device are the correct ones")
	}

	pubKey2, _, err := getPubKeyAddrSafe(device, path, crypto.Bech32AddrPrefix())
	if err != nil {
		return err
	}

	if pubKey2 != expectedPubKey {
		return fmt.Errorf("the key's pubkey does not match with the one retrieved from Ledger. Check that the HD path and device are the correct ones")
	}

	return nil
}

// ValidateKey allows us to verify the sanity of a public key after loading it
// from disk.
func (pkl PrivKeyLedgerSecp256k1) ValidateKey() error {
	device, err := getLedgerDevice()
	if err != nil {
		return err
	}
	defer warnIfErrors(device.Close)

	return validateKey(device, pkl)
}

// AssertIsPrivKeyInner implements the PrivKey interface. It performs a no-op.
func (pkl *PrivKeyLedgerSecp256k1) AssertIsPrivKeyInner() {}

// Bytes implements the PrivKey interface. It stores the cached public key so
// we can verify the same key when we reconnect to a ledger.
func (pkl PrivKeyLedgerSecp256k1) Bytes() []byte {
	return amino.MustMarshal(pkl)
}

// Equals implements the PrivKey interface. It makes sure two private keys
// refer to the same public key.
func (pkl PrivKeyLedgerSecp256k1) Equals(other crypto.PrivKey) bool {
	if otherKey, ok := other.(PrivKeyLedgerSecp256k1); ok {
		return pkl.CachedPubKey.Equals(otherKey.CachedPubKey)
	}
	return false
}

// warnIfErrors wraps a function and writes a warning to stderr. This is required
// to avoid ignoring errors when defer is used. Using defer may result in linter warnings.
func warnIfErrors(f func() error) {
	if err := f(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, "received error when closing ledger connection", err)
	}
}

func convertDERtoBER(signatureDER []byte) ([]byte, error) {
	sigDER, err := ecdsa.ParseDERSignature(signatureDER)
	if err != nil {
		return nil, err
	}

	sigStr := sigDER.Serialize()
	// The format of a DER encoded signature is as follows:
	// 0x30 <total length> 0x02 <length of R> <R> 0x02 <length of S> <S>
	r, s := new(big.Int), new(big.Int)
	r.SetBytes(sigStr[4 : 4+sigStr[3]])
	s.SetBytes(sigStr[4+sigStr[3]+2:])

	sModNScalar := new(secp.ModNScalar)
	sModNScalar.SetByteSlice(s.Bytes())
	if sModNScalar.IsOverHalfOrder() {
		s = new(big.Int).Sub(secp.S256().N, s)
	}

	sigBytes := make([]byte, 64)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(r.Bytes()):32], r.Bytes())
	copy(sigBytes[64-len(s.Bytes()):64], s.Bytes())

	return sigBytes, nil
}

func getLedgerDevice() (ledger.SECP256K1, error) {
	device, err := ledger.Discover()
	if err != nil {
		return nil, errors.Wrap(err, "ledger nano S")
	}

	return device, nil
}

func validateKey(device ledger.SECP256K1, pkl PrivKeyLedgerSecp256k1) error {
	pub, err := getPubKeyUnsafe(device, pkl.Path)
	if err != nil {
		return err
	}

	// verify this matches cached address
	if !pub.Equals(pkl.CachedPubKey) {
		return fmt.Errorf("cached key does not match retrieved key")
	}

	return nil
}

// Sign calls the ledger and stores the PubKey for future use.
//
// Communication is checked on NewPrivKeyLedger and PrivKeyFromBytes, returning
// an error, so this should only trigger if the private key is held in memory
// for a while before use.
func sign(device ledger.SECP256K1, pkl PrivKeyLedgerSecp256k1, msg []byte) ([]byte, error) {
	err := validateKey(device, pkl)
	if err != nil {
		return nil, err
	}
	sig, err := device.SignSECP256K1(pkl.Path.DerivationPath(), msg, 0)
	if err != nil {
		return nil, err
	}

	return convertDERtoBER(sig)
}

// getPubKeyUnsafe reads the pubkey from a ledger device
//
// This function is marked as unsafe as it will retrieve a pubkey without user verification
// It can only be used to verify a pubkey but never to create new accounts/keys. In that case,
// please refer to getPubKeyAddrSafe
//
// since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling
func getPubKeyUnsafe(device ledger.SECP256K1, path hd.BIP44Params) (crypto.PubKey, error) {
	publicKey, err := device.GetPublicKeySECP256K1(path.DerivationPath())
	if err != nil {
		return nil, fmt.Errorf("please open Cosmos app on the Ledger device - error: %w", err)
	}

	// re-serialize in the 33-byte compressed format
	cmp, err := btcec.ParsePubKey(publicKey[:])
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}

	var compressedPublicKey secp256k1.PubKeySecp256k1
	copy(compressedPublicKey[:], cmp.SerializeCompressed())

	return compressedPublicKey, nil
}

// getPubKeyAddr reads the pubkey and the address from a ledger device.
// This function is marked as Safe as it will require user confirmation and
// account and index will be shown in the device.
//
// Since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling.
func getPubKeyAddrSafe(device ledger.SECP256K1, path hd.BIP44Params, hrp string) (crypto.PubKey, string, error) {
	publicKey, addr, err := device.GetAddressPubKeySECP256K1(path.DerivationPath(), hrp)
	if err != nil {
		return nil, "", fmt.Errorf("address %s rejected", addr)
	}

	// re-serialize in the 33-byte compressed format
	cmp, err := btcec.ParsePubKey(publicKey[:])
	if err != nil {
		return nil, "", fmt.Errorf("error parsing public key: %w", err)
	}

	var compressedPublicKey secp256k1.PubKeySecp256k1
	copy(compressedPublicKey[:], cmp.SerializeCompressed())

	return compressedPublicKey, addr, nil
}
