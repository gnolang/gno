package keys

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
)

// Keybase exposes operations on a generic keystore
type Keybase interface {
	// CRUD on the keystore
	List() ([]Info, error)
	HasByNameOrAddress(nameOrBech32 string) (bool, error)
	HasByName(name string) (bool, error)
	HasByAddress(address crypto.Address) (bool, error)
	GetByNameOrAddress(nameOrBech32 string) (Info, error)
	GetByName(name string) (Info, error)
	GetByAddress(address crypto.Address) (Info, error)
	Delete(name, passphrase string, skipPass bool) error

	// Sign some bytes, looking up the private key to use
	Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error)
	Verify(name string, msg, sig []byte) error

	// CreateAccount creates an account based using the BIP44 path (44'/118'/{account}'/0/{index}
	// Encrypt the key to disk using encryptPasswd.
	// If an account exists with the same address but a different name, it is replaced by the new name.
	// See https://github.com/cosmos/cosmos-sdk/issues/2095
	CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error)

	// Like CreateAccount but from general bip44 params.
	// If an account exists with the same address but a different name, it is replaced by the new name.
	CreateAccountBip44(name, mnemonic, bip39Passwd, encryptPasswd string, params hd.BIP44Params) (Info, error)

	// CreateLedger creates, stores, and returns a new Ledger key reference
	CreateLedger(name string, algo SigningAlgo, hrp string, account, index uint32) (info Info, err error)

	// CreateOffline creates, stores, and returns a new offline key reference
	CreateOffline(name string, pubkey crypto.PubKey) (info Info, err error)

	// CreateMulti creates, stores, and returns a new multsig (offline) key reference
	CreateMulti(name string, pubkey crypto.PubKey) (info Info, err error)

	// Rename renames an existing key from oldName to newName
	Rename(oldName, newName string) error

	// Rotate replaces the encryption password for a given key
	Rotate(name, oldpass string, getNewpass func() (string, error)) error

	// ImportPrivKey imports the given private key into the keybase.
	// In all import operations, if an account exists with the same address but a different name, it is replaced by the new name
	ImportPrivKey(name string, key crypto.PrivKey, encryptPass string) error

	// ExportPrivKey exports the private key from the keybase. It *only* works on locally-stored keys
	ExportPrivKey(name string, decryptPass string) (crypto.PrivKey, error)

	// CloseDB closes the database.
	CloseDB()
}

// KeyType reflects a human-readable type for key listing.
type KeyType uint

// Info KeyTypes
const (
	TypeLocal   KeyType = 0
	TypeLedger  KeyType = 1
	TypeOffline KeyType = 2
	TypeMulti   KeyType = 3
)

var keyTypes = map[KeyType]string{
	TypeLocal:   "local",
	TypeLedger:  "ledger",
	TypeOffline: "offline",
	TypeMulti:   "multi",
}

// String implements the stringer interface for KeyType.
func (kt KeyType) String() string {
	return keyTypes[kt]
}

// Info is the publicly exposed information about a keypair
type Info interface {
	// Human-readable type for key listing
	GetType() KeyType
	// Name of the key
	GetName() string
	// Public key
	GetPubKey() crypto.PubKey
	// Address
	GetAddress() crypto.Address
	// Bip44 Path
	GetPath() (*hd.BIP44Params, error)
}

var (
	_ Info = &localInfo{}
	_ Info = &ledgerInfo{}
	_ Info = &offlineInfo{}
	_ Info = &multiInfo{}
)

// localInfo is the public information about a locally stored key
type localInfo struct {
	Name         string        `json:"name"`
	PubKey       crypto.PubKey `json:"pubkey"`
	PrivKeyArmor string        `json:"privkey.armor"`
}

func newLocalInfo(name string, pub crypto.PubKey, privArmor string) Info {
	return &localInfo{
		Name:         name,
		PubKey:       pub,
		PrivKeyArmor: privArmor,
	}
}

// GetType implements Info interface
func (i localInfo) GetType() KeyType {
	return TypeLocal
}

// GetType implements Info interface
func (i localInfo) GetName() string {
	return i.Name
}

// GetType implements Info interface
func (i localInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// GetType implements Info interface
func (i localInfo) GetAddress() crypto.Address {
	return i.PubKey.Address()
}

// GetType implements Info interface
func (i localInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// ledgerInfo is the public information about a Ledger key
type ledgerInfo struct {
	Name   string         `json:"name"`
	PubKey crypto.PubKey  `json:"pubkey"`
	Path   hd.BIP44Params `json:"path"`
}

func newLedgerInfo(name string, pub crypto.PubKey, path hd.BIP44Params) Info {
	return &ledgerInfo{
		Name:   name,
		PubKey: pub,
		Path:   path,
	}
}

// GetType implements Info interface
func (i ledgerInfo) GetType() KeyType {
	return TypeLedger
}

// GetName implements Info interface
func (i ledgerInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i ledgerInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// GetAddress implements Info interface
func (i ledgerInfo) GetAddress() crypto.Address {
	return i.PubKey.Address()
}

// GetPath implements Info interface
func (i ledgerInfo) GetPath() (*hd.BIP44Params, error) {
	tmp := i.Path
	return &tmp, nil
}

// offlineInfo is the public information about an offline key
type offlineInfo struct {
	Name   string        `json:"name"`
	PubKey crypto.PubKey `json:"pubkey"`
}

func newOfflineInfo(name string, pub crypto.PubKey) Info {
	return &offlineInfo{
		Name:   name,
		PubKey: pub,
	}
}

// GetType implements Info interface
func (i offlineInfo) GetType() KeyType {
	return TypeOffline
}

// GetName implements Info interface
func (i offlineInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i offlineInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// GetAddress implements Info interface
func (i offlineInfo) GetAddress() crypto.Address {
	return i.PubKey.Address()
}

// GetPath implements Info interface
func (i offlineInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

type multisigPubKeyInfo struct {
	PubKey crypto.PubKey `json:"pubkey"`
	Weight uint          `json:"weight"`
}

// multiInfo is the public information about a multisig key
type multiInfo struct {
	Name      string               `json:"name"`
	PubKey    crypto.PubKey        `json:"pubkey"`
	Threshold uint                 `json:"threshold"`
	PubKeys   []multisigPubKeyInfo `json:"pubkeys"`
}

// NewMultiInfo creates a new multiInfo instance
func NewMultiInfo(name string, pub crypto.PubKey) Info {
	multiPK := pub.(multisig.PubKeyMultisigThreshold)

	pubKeys := make([]multisigPubKeyInfo, len(multiPK.PubKeys))
	for i, pk := range multiPK.PubKeys {
		// TODO: Recursively check pk for total weight?
		pubKeys[i] = multisigPubKeyInfo{pk, 1}
	}

	return &multiInfo{
		Name:      name,
		PubKey:    pub,
		Threshold: multiPK.K,
		PubKeys:   pubKeys,
	}
}

// GetType implements Info interface
func (i multiInfo) GetType() KeyType {
	return TypeMulti
}

// GetName implements Info interface
func (i multiInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i multiInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// GetAddress implements Info interface
func (i multiInfo) GetAddress() crypto.Address {
	return i.PubKey.Address()
}

// GetPath implements Info interface
func (i multiInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// encoding info
func writeInfo(i Info) []byte {
	return amino.MustMarshalAnySized(i)
}

// decoding info
func readInfo(bz []byte) (info Info, err error) {
	err = amino.UnmarshalAnySized(bz, &info)
	return
}
