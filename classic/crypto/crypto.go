package crypto

import (
	"bytes"
	"fmt"

	"github.com/tendermint/classic/crypto/tmhash"
	"github.com/tendermint/classic/libs/bech32"
)

//----------------------------------------
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
		panic(fmt.Errorf("invalid address string representation: %v, error: %v", str, err))
	}
	return
}

func AddressFromPreimage(bz []byte) Address {
	return AddressFromBytes(tmhash.SumTruncated(bz))
}

func AddressFromBytes(bz []byte) (ret Address) {
	if len(bz) != AddressSize {
		panic(fmt.Errorf("unexpected address byte length. expected %v, got %v", AddressSize, len(bz)))
	}
	copy(ret[:], bz)
	return
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
	// The "c" bech32 is intended to be constant,
	// and enforced upon all users of the tendermint/classic repo
	// and derivations of tendermint/classic.
	bech32Addr, err := bech32.Encode("c", addr[:])
	if err != nil {
		panic(err)
	}
	return bech32Addr
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
	if pre != "c" {
		return fmt.Errorf("unexpected bech32 prefix for address. expected \"c\", got %v", pre)
	}
	if len(bz) != AddressSize {
		return fmt.Errorf("unexpected address byte length. expected %v, got %v", AddressSize, len(bz))
	}
	copy((*addr)[:], bz)
	return nil
}

//----------------------------------------
// ID

// The bech32 representation w/ prefix "c".
type ID string

func (id ID) IsZero() bool {
	return id == ""
}

func (id ID) String() string {
	return string(id)
}

func (id ID) Validate() error {
	if id.IsZero() {
		return fmt.Errorf("zero ID is invalid")
	}
	var addr Address
	err := addr.DecodeID(id)
	return err
}

func AddressFromID(id ID) (addr Address, err error) {
	err = addr.DecodeString(string(id))
	return
}

func (addr Address) ID() ID {
	return ID(addr.String())
}

func (addr *Address) DecodeID(id ID) error {
	return addr.DecodeString(string(id))
}

//----------------------------------------
// PubKey

// All operations must be deterministic.
type PubKey interface {
	// Stable
	Address() Address
	Bytes() []byte
	VerifyBytes(msg []byte, sig []byte) bool
	Equals(PubKey) bool

	// Unstable
	String() string
}

//----------------------------------------
// PrivKey

// All operations must be deterministic.
type PrivKey interface {
	// Stable
	Bytes() []byte
	Sign(msg []byte) ([]byte, error)
	PubKey() PubKey
	Equals(PrivKey) bool
}

//----------------------------------------
// Symmetric

type Symmetric interface {
	Keygen() []byte
	Encrypt(plaintext []byte, secret []byte) (ciphertext []byte)
	Decrypt(ciphertext []byte, secret []byte) (plaintext []byte, err error)
}
