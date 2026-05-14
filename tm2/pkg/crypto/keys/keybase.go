package keys

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/armor"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	"github.com/gnolang/gno/tm2/pkg/crypto/ledger"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

var (
	errCannotOverwrite = errors.New("cannot overwrite existing key")
	errKeyNotAvailable = errors.New("private key not available")
)

var _ Keybase = dbKeybase{}

// Language is a language to create the BIP 39 mnemonic in.
// Currently, only english is supported though.
// Find a list of all supported languages in the BIP 39 spec (word lists).
type Language int

// noinspection ALL
const (
	// English is the default language to create a mnemonic.
	// It is the only supported language by this package.
	English Language = iota + 1
	// Japanese is currently not supported.
	Japanese
	// Korean is currently not supported.
	Korean
	// Spanish is currently not supported.
	Spanish
	// ChineseSimplified is currently not supported.
	ChineseSimplified
	// ChineseTraditional is currently not supported.
	ChineseTraditional
	// French is currently not supported.
	French
	// Italian is currently not supported.
	Italian
)

const (
	addressSuffix = "address"
	infoSuffix    = "info"
)

var (
	// ErrUnsupportedSigningAlgo is raised when the caller tries to use a
	// different signing scheme than secp256k1.
	ErrUnsupportedSigningAlgo = errors.New("unsupported signing algo: only secp256k1 is supported")

	// ErrUnsupportedLanguage is raised when the caller tries to use a
	// different language than english for creating a mnemonic sentence.
	ErrUnsupportedLanguage = errors.New("unsupported language: only english is supported")
)

// dbKeybase combines encryption and storage implementation to provide
// a full-featured key manager
type dbKeybase struct {
	db dbm.DB
}

// NewDBKeybase creates a new keybase instance using the passed DB for reading and writing keys.
func NewDBKeybase(db dbm.DB) Keybase {
	return dbKeybase{
		db: db,
	}
}

// NewInMemory creates a transient keybase on top of in-memory storage
// instance useful for testing purposes and on-the-fly key generation.
func NewInMemory() Keybase { return dbKeybase{memdb.NewMemDB()} }

// CreateAccount converts a mnemonic to a private key and persists it, encrypted with the given password.
// XXX Info could include the separately derived ed25519 key,
// XXX and a signature from the sec2561key as certificate.
// XXX NOTE: we are not saving the derivation path.
// XXX but this doesn't help encrypted communication.
// XXX also there is no document structure.
func (kb dbKeybase) CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error) {
	coinType := crypto.CoinType
	hdPath := hd.NewFundraiserParams(account, coinType, index)
	return kb.CreateAccountBip44(name, mnemonic, bip39Passwd, encryptPasswd, *hdPath)
}

func (kb dbKeybase) CreateAccountBip44(name, mnemonic, bip39Passphrase, encryptPasswd string, params hd.BIP44Params) (info Info, err error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return
	}

	info, err = kb.persistDerivedKey(seed, encryptPasswd, name, params.String())
	return
}

// CreateLedger creates a new locally-stored reference to a Ledger keypair
// It returns the created key info and an error if the Ledger could not be queried
func (kb dbKeybase) CreateLedger(name string, algo SigningAlgo, hrp string, account, index uint32) (Info, error) {
	if algo != Secp256k1 {
		return nil, ErrUnsupportedSigningAlgo
	}

	coinType := crypto.CoinType
	hdPath := hd.NewFundraiserParams(account, coinType, index)
	priv, _, err := ledger.NewPrivKeyLedgerSecp256k1(*hdPath, hrp)
	if err != nil {
		return nil, err
	}
	pub := priv.PubKey()

	// Note: Once Cosmos App v1.3.1 is compulsory, it could be possible to check that pubkey and addr match
	return kb.writeLedgerKey(name, pub, *hdPath)
}

// CreateOffline creates a new reference to an offline keypair. It returns the
// created key info.
func (kb dbKeybase) CreateOffline(name string, pub crypto.PubKey) (Info, error) {
	return kb.writeOfflineKey(name, pub)
}

// CreateMulti creates a new reference to a multisig (offline) keypair. It
// returns the created key info.
func (kb dbKeybase) CreateMulti(name string, pub crypto.PubKey) (Info, error) {
	return kb.writeMultisigKey(name, pub)
}

func (kb *dbKeybase) persistDerivedKey(seed []byte, passwd, name, fullHdPath string) (Info, error) {
	// create master key and derive first key:
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, fullHdPath)
	if err != nil {
		return nil, err
	}

	// use possibly blank password to encrypt the private
	// key and store it. User must enforce good passwords.
	return kb.writeLocalKey(name, secp256k1.PrivKeySecp256k1(derivedPriv), passwd)
}

// List returns the keys from storage in alphabetical order.
func (kb dbKeybase) List() ([]Info, error) {
	var res []Info
	iter, err := kb.db.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		key := string(iter.Key())

		// need to include only keys in storage that have an info suffix
		if strings.HasSuffix(key, infoSuffix) {
			info, err := readInfo(iter.Value())
			if err != nil {
				return nil, err
			}
			res = append(res, info)
		}
	}
	return res, nil
}

// HasByNameOrAddress checks if a key with the name or bech32 string address is in the keybase.
func (kb dbKeybase) HasByNameOrAddress(nameOrBech32 string) (bool, error) {
	address, err := crypto.AddressFromBech32(nameOrBech32)
	if err != nil {
		return kb.HasByName(nameOrBech32)
	}
	return kb.HasByAddress(address)
}

// HasByName checks if a key with the name is in the keybase.
func (kb dbKeybase) HasByName(name string) (bool, error) {
	return kb.db.Has(infoKey(name))
}

// HasByAddress checks if a key with the address is in the keybase.
func (kb dbKeybase) HasByAddress(address crypto.Address) (bool, error) {
	return kb.db.Has(addrKey(address))
}

// Get returns the public information about one key.
func (kb dbKeybase) GetByNameOrAddress(nameOrBech32 string) (Info, error) {
	addr, err := crypto.AddressFromBech32(nameOrBech32)
	if err != nil {
		return kb.GetByName(nameOrBech32)
	}
	return kb.GetByAddress(addr)
}

func (kb dbKeybase) GetByName(name string) (Info, error) {
	bs, err := kb.db.Get(infoKey(name))
	if err != nil {
		return nil, fmt.Errorf("error while getting key %s from db: %w", name, err)
	}
	if len(bs) == 0 {
		return nil, keyerror.NewErrKeyNotFound(name)
	}
	return readInfo(bs)
}

func (kb dbKeybase) GetByAddress(address crypto.Address) (Info, error) {
	ik, err := kb.db.Get(addrKey(address))
	if err != nil {
		return nil, fmt.Errorf("error while getting key with address %s from db: %w", address, err)
	}
	if len(ik) == 0 {
		return nil, keyerror.NewErrKeyNotFound(fmt.Sprintf("key with address %s not found", address))
	}
	bs, err := kb.db.Get(ik)
	if err != nil {
		return nil, fmt.Errorf("error while getting info for address %s from db: %w", address, err)
	}
	return readInfo(bs)
}

// Sign signs the msg with the named key.
// It returns an error if the key doesn't exist or the decryption fails.
func (kb dbKeybase) Sign(nameOrBech32, passphrase string, msg []byte) (sig []byte, pub crypto.PubKey, err error) {
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return
	}

	var priv crypto.PrivKey

	switch info := info.(type) {
	case localInfo:
		if info.PrivKeyArmor == "" {
			err = fmt.Errorf("%w: %s", errKeyNotAvailable, nameOrBech32)
			return
		}

		priv, err = armor.UnarmorDecryptPrivKey(info.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, nil, err
		}

	case ledgerInfo:
		priv, err = ledger.NewPrivKeyLedgerSecp256k1Unsafe(info.Path)
		if err != nil {
			return
		}

	case offlineInfo, multiInfo:
		err = fmt.Errorf("cannot sign with key or addr %s", nameOrBech32)
		return
	}

	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	pub = priv.PubKey()
	return sig, pub, nil
}

// Verify verifies the sig+msg with the named key.
// It returns an error if the key doesn't exist or verification fails.
func (kb dbKeybase) Verify(nameOrBech32 string, msg []byte, sig []byte) (err error) {
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return
	}

	pub := info.GetPubKey()
	if !pub.VerifyBytes(msg, sig) {
		return errors.New("invalid signature")
	}
	return nil
}

func (kb dbKeybase) ImportPrivKey(name string, key crypto.PrivKey, encryptPass string) error {
	if _, err := kb.GetByNameOrAddress(name); err == nil {
		return fmt.Errorf("%w: %s", errCannotOverwrite, name)
	}

	_, err := kb.writeLocalKey(name, key, encryptPass)

	return err
}

func (kb dbKeybase) ExportPrivKey(nameOrBech32 string, passphrase string) (crypto.PrivKey, error) {
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return nil, err
	}

	var priv crypto.PrivKey

	switch info := info.(type) {
	case localInfo:
		if info.PrivKeyArmor == "" {
			return nil, fmt.Errorf("%w: %s", errKeyNotAvailable, nameOrBech32)
		}

		priv, err = armor.UnarmorDecryptPrivKey(info.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, err
		}
	case ledgerInfo, offlineInfo, multiInfo:
		return nil, errors.New("only works on local private keys")
	}

	return priv, nil
}

// Delete removes key forever, but we must present the
// proper passphrase before deleting it (for security).
// It returns an error if the key doesn't exist or
// passphrases don't match.
// Passphrase is ignored when deleting references to
// offline and Ledger / HW wallet keys.
func (kb dbKeybase) Delete(nameOrBech32, passphrase string, skipPass bool) error {
	// verify we have the proper password before deleting
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	if linfo, ok := info.(localInfo); ok && !skipPass {
		if _, err = armor.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase); err != nil {
			return err
		}
	}
	kb.db.DeleteSync(addrKey(info.GetAddress()))
	kb.db.DeleteSync(infoKey(info.GetName()))
	return nil
}

// Rename renames an existing key from oldName to newName.
// It returns an error if oldName doesn't exist or newName already exists.
func (kb dbKeybase) Rename(oldName, newName string) error {
	info, err := kb.GetByName(oldName)
	if err != nil {
		return err
	}

	exists, err := kb.HasByName(newName)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("key with name %q already exists", newName)
	}

	var newInfo Info

	switch i := info.(type) {
	case localInfo:
		i.Name = newName
		newInfo = &i
	case ledgerInfo:
		i.Name = newName
		newInfo = &i
	case offlineInfo:
		i.Name = newName
		newInfo = &i
	case multiInfo:
		i.Name = newName
		newInfo = &i
	default:
		return fmt.Errorf("unsupported key type for rename")
	}

	// Explicitly delete the old name entry before writing the new one.
	// writeInfo would clean it up via address dedup, but being explicit
	// avoids relying on that side effect.
	kb.db.DeleteSync(infoKey(oldName))

	return kb.writeInfo(newName, newInfo)
}

// Rotate changes the passphrase with which an already stored key is
// encrypted.
//
// oldpass must be the current passphrase used for encryption,
// getNewpass is a function to get the passphrase to permanently replace
// the current passphrase
func (kb dbKeybase) Rotate(nameOrBech32, oldpass string, getNewpass func() (string, error)) error {
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	switch info := info.(type) {
	case localInfo:
		key, err := armor.UnarmorDecryptPrivKey(info.PrivKeyArmor, oldpass)
		if err != nil {
			return err
		}
		newpass, err := getNewpass()
		if err != nil {
			return err
		}
		kb.writeLocalKey(info.GetName(), key, newpass)
		return nil
	default:
		return fmt.Errorf("locally stored key required. Received: %v", reflect.TypeOf(info).String())
	}
}

// CloseDB releases the lock and closes the storage backend.
func (kb dbKeybase) CloseDB() {
	kb.db.Close()
}

func (kb dbKeybase) writeLocalKey(name string, priv crypto.PrivKey, passphrase string) (Info, error) {
	// encrypt private key using passphrase
	privArmor := armor.EncryptArmorPrivKey(priv, passphrase)
	// make Info
	pub := priv.PubKey()
	info := newLocalInfo(name, pub, privArmor)
	if err := kb.writeInfo(name, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (kb dbKeybase) writeLedgerKey(name string, pub crypto.PubKey, path hd.BIP44Params) (Info, error) {
	info := newLedgerInfo(name, pub, path)
	if err := kb.writeInfo(name, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (kb dbKeybase) writeOfflineKey(name string, pub crypto.PubKey) (Info, error) {
	info := newOfflineInfo(name, pub)
	if err := kb.writeInfo(name, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (kb dbKeybase) writeMultisigKey(name string, pub crypto.PubKey) (Info, error) {
	info := NewMultiInfo(name, pub)
	if err := kb.writeInfo(name, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (kb dbKeybase) writeInfo(name string, info Info) error {
	// write the info by key
	key := infoKey(name)
	oldInfob, err := kb.db.Get(key)
	if err != nil {
		return fmt.Errorf("error while getting info by key %v: %w", key, err)
	}
	if len(oldInfob) > 0 {
		// Enforce 1-to-1 name to address. Remove the lookup by the old address
		oldInfo, err := readInfo(oldInfob)
		if err != nil {
			return err
		}
		kb.db.DeleteSync(addrKey(oldInfo.GetAddress()))
	}

	addressKey := addrKey(info.GetAddress())
	nameKeyForAddress, err := kb.db.Get(addressKey)
	if err != nil {
		return fmt.Errorf("error while getting key for address %v: %w", info.GetAddress().String(), err)
	}
	if len(nameKeyForAddress) > 0 {
		// Enforce 1-to-1 name to address. Remove the info by the old name with the same address
		kb.db.DeleteSync(nameKeyForAddress)
	}

	serializedInfo := writeInfo(info)
	kb.db.SetSync(key, serializedInfo)
	// store a pointer to the infokey by address for fast lookup
	kb.db.SetSync(addressKey, key)
	return nil
}

func addrKey(address crypto.Address) []byte {
	return fmt.Appendf(nil, "%s.%s", address.String(), addressSuffix)
}

func infoKey(name string) []byte {
	return fmt.Appendf(nil, "%s.%s", name, infoSuffix)
}
