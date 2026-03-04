package crypto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
)

// ----------------------------------------
// Bech32Address

type Bech32Address string

func (b32 Bech32Address) String() string {
	return string(b32)
}

// ----------------------------------------
// Address

const (
	// AddressSize is the size of a pubkey address.
	AddressSize = tmhash.TruncatedSize
)

// (truncated) hash of some preimage (typically of a pubkey).
type Address [AddressSize]byte

func AddressFromString(str string) (addr Address, err error) {
	err = addr.DecodeString(str)
	return
}

func MustAddressFromString(str string) (addr Address) {
	err := addr.DecodeString(str)
	if err != nil {
		panic(fmt.Errorf("invalid address string representation: %v, error: %w", str, err))
	}
	return
}

func AddressFromPreimage(bz []byte) Address {
	addr := AddressFromBytes(tmhash.SumTruncated(bz))
	return addr
}

func AddressFromBytes(bz []byte) (ret Address) {
	if len(bz) != AddressSize {
		panic(fmt.Errorf("unexpected address byte length. expected %v, got %v", AddressSize, len(bz)))
	}
	copy(ret[:], bz)
	return
}

func (addr Address) MarshalJSON() ([]byte, error) {
	b := AddressToBech32(addr)
	return []byte(`"` + b + `"`), nil
}

func (addr *Address) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return addr.UnmarshalAmino(s)
}

func (addr Address) MarshalAmino() (string, error) {
	return AddressToBech32(addr), nil
}

func (addr *Address) UnmarshalAmino(b32str string) (err error) {
	// NOTE: also used to unmarshal normal JSON, through UnmarshalJSON.
	if b32str == "" {
		return nil // leave addr as zero.
	}
	addr2, err := AddressFromBech32(b32str)
	if err != nil {
		return err
	}
	copy(addr[:], addr2[:])
	return nil
}

func (addr Address) Compare(other Address) int {
	bz1 := make([]byte, len(addr))
	bz2 := make([]byte, len(other))
	copy(bz1, addr[:])
	copy(bz2, other[:])
	return bytes.Compare(bz1, bz2)
}

func (addr Address) IsZero() bool {
	return addr == Address{}
}

func (addr Address) String() string {
	return AddressToBech32(addr)
}

func (addr Address) Bech32() Bech32Address {
	return Bech32Address(AddressToBech32(addr))
}

func (addr Address) Bytes() []byte {
	res := make([]byte, AddressSize)
	copy(res, addr[:])
	return res
}

func (addr *Address) DecodeString(str string) error {
	pre, bz, err := bech32.Decode(str)
	if err != nil {
		return err
	}
	if pre != Bech32AddrPrefix() {
		return fmt.Errorf("unexpected bech32 prefix for address. expected %q, got %q", Bech32AddrPrefix(), pre)
	}
	if len(bz) != AddressSize {
		return fmt.Errorf("unexpected address byte length. expected %v, got %v", AddressSize, len(bz))
	}
	copy((*addr)[:], bz)
	return nil
}

// ----------------------------------------
// ID

var ErrZeroID = errors.New("address ID is zero")

// The bech32 representation w/ bech32 prefix.
type ID string

func (id ID) IsZero() bool {
	return id == ""
}

func (id ID) String() string {
	return string(id)
}

func (id ID) Validate() error {
	if id.IsZero() {
		return ErrZeroID
	}

	var addr Address

	return addr.DecodeID(id)
}

func (addr Address) ID() ID {
	return ID(addr.String())
}

func (addr *Address) DecodeID(id ID) error {
	return addr.DecodeString(string(id))
}

// ----------------------------------------
// PubKey

// All operations must be deterministic.
type PubKey interface {
	// Stable
	Address() Address
	Bytes() []byte
	VerifyBytes(msg []byte, sig []byte) bool
	Equals(PubKey) bool
	String() string
}

// ----------------------------------------
// PrivKey

// All operations must be deterministic.
type PrivKey interface {
	// Stable
	Bytes() []byte
	Sign(msg []byte) ([]byte, error)
	PubKey() PubKey
	Equals(PrivKey) bool
}

// ----------------------------------------
// Symmetric

type Symmetric interface {
	Keygen() []byte
	Encrypt(plaintext []byte, secret []byte) (ciphertext []byte)
	Decrypt(ciphertext []byte, secret []byte) (plaintext []byte, err error)
}
