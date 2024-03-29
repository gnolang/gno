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
	return kb.writeLedgerKey(name, pub, *hdPath), nil
}

// CreateOffline creates a new reference to an offline keypair. It returns the
// created key info.
func (kb dbKeybase) CreateOffline(name string, pub crypto.PubKey) (Info, error) {
	return kb.writeOfflineKey(name, pub), nil
}

// CreateMulti creates a new reference to a multisig (offline) keypair. It
// returns the created key info.
func (kb dbKeybase) CreateMulti(name string, pub crypto.PubKey) (Info, error) {
	return kb.writeMultisigKey(name, pub), nil
}

func (kb *dbKeybase) persistDerivedKey(seed []byte, passwd, name, fullHdPath string) (info Info, err error) {
	// create master key and derive first key:
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, fullHdPath)
	if err != nil {
		return
	}

	// use possibly blank password to encrypt the private
	// key and store it. User must enforce good passwords.
	info = kb.writeLocalKey(name, secp256k1.PrivKeySecp256k1(derivedPriv), passwd)
	return
}

// List returns the keys from storage in alphabetical order.
func (kb dbKeybase) List() ([]Info, error) {
	var res []Info
	iter := kb.db.Iterator(nil, nil)
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
	return kb.db.Has(infoKey(name)), nil
}

// HasByAddress checks if a key with the address is in the keybase.
func (kb dbKeybase) HasByAddress(address crypto.Address) (bool, error) {
	return kb.db.Has(addrKey(address)), nil
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
	bs := kb.db.Get(infoKey(name))
	if len(bs) == 0 {
		return nil, keyerror.NewErrKeyNotFound(name)
	}
	return readInfo(bs)
}

func (kb dbKeybase) GetByAddress(address crypto.Address) (Info, error) {
	ik := kb.db.Get(addrKey(address))
	if len(ik) == 0 {
		return nil, keyerror.NewErrKeyNotFound(fmt.Sprintf("key with address %s not found", address))
	}
	bs := kb.db.Get(ik)
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

	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return
		}

		priv, err = armor.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, nil, err
		}

	case ledgerInfo:
		linfo := info.(ledgerInfo)
		priv, err = ledger.NewPrivKeyLedgerSecp256k1Unsafe(linfo.Path)
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

	var pub crypto.PubKey
	pub = info.GetPubKey()
	if !pub.VerifyBytes(msg, sig) {
		return errors.New("invalid signature")
	}
	return nil
}

func (kb dbKeybase) ExportPrivateKeyObject(nameOrBech32 string, passphrase string) (crypto.PrivKey, error) {
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return nil, err
	}

	var priv crypto.PrivKey

	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return nil, err
		}
		priv, err = armor.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, err
		}

	case ledgerInfo, offlineInfo, multiInfo:
		return nil, errors.New("only works on local private keys")
	}

	return priv, nil
}

func (kb dbKeybase) Export(nameOrBech32 string) (astr string, err error) {
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return "", errors.Wrap(err, "getting info for name %s", nameOrBech32)
	}
	bz := kb.db.Get(infoKey(info.GetName()))
	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", nameOrBech32)
	}
	return armor.ArmorInfoBytes(bz), nil
}

// ExportPubKey returns public keys in ASCII armored format.
// Retrieve a Info object by its name and return the public key in
// a portable format.
func (kb dbKeybase) ExportPubKey(nameOrBech32 string) (astr string, err error) {
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return "", errors.Wrap(err, "getting info for name %s", nameOrBech32)
	}
	return armor.ArmorPubKeyBytes(info.GetPubKey().Bytes()), nil
}

// ExportPrivKey returns a private key in ASCII armored format.
// It returns an error if the key does not exist or a wrong encryption passphrase is supplied.
func (kb dbKeybase) ExportPrivKey(
	name,
	decryptPassphrase,
	encryptPassphrase string,
) (astr string, err error) {
	priv, err := kb.ExportPrivateKeyObject(name, decryptPassphrase)
	if err != nil {
		return "", err
	}

	return armor.EncryptArmorPrivKey(priv, encryptPassphrase), nil
}

// ExportPrivKeyUnsafe returns a private key in ASCII armored format.
// The returned armor is unencrypted.
// It returns an error if the key does not exist
func (kb dbKeybase) ExportPrivKeyUnsafe(
	name,
	decryptPassphrase string,
) (string, error) {
	priv, err := kb.ExportPrivateKeyObject(name, decryptPassphrase)
	if err != nil {
		return "", err
	}

	return armor.ArmorPrivateKey(priv), nil
}

// ImportPrivKey imports a private key in ASCII armor format.
// It returns an error if a key with the same name exists or a wrong encryption passphrase is
// supplied.
func (kb dbKeybase) ImportPrivKey(
	name,
	astr,
	decryptPassphrase,
	encryptPassphrase string,
) error {
	if _, err := kb.GetByNameOrAddress(name); err == nil {
		return errors.New("Cannot overwrite key " + name)
	}
	privKey, err := armor.UnarmorDecryptPrivKey(astr, decryptPassphrase)
	if err != nil {
		return errors.Wrap(err, "couldn't import private key")
	}

	kb.writeLocalKey(name, privKey, encryptPassphrase)
	return nil
}

func (kb dbKeybase) ImportPrivKeyUnsafe(
	name,
	armorStr,
	encryptPassphrase string,
) error {
	if _, err := kb.GetByNameOrAddress(name); err == nil {
		return fmt.Errorf("cannot overwrite key %s", name)
	}

	privKey, err := armor.UnarmorPrivateKey(armorStr)
	if err != nil {
		return errors.Wrap(err, "couldn't import private key")
	}

	kb.writeLocalKey(name, privKey, encryptPassphrase)
	return nil
}

func (kb dbKeybase) Import(name, astr string) (err error) {
	if _, err := kb.GetByNameOrAddress(name); err == nil {
		return errors.New("Cannot overwrite key " + name)
	}
	infoBytes, err := armor.UnarmorInfoBytes(astr)
	if err != nil {
		return
	}
	kb.db.Set(infoKey(name), infoBytes)
	return nil
}

// ImportPubKey imports ASCII-armored public keys.
// Store a new Info object holding a public key only, i.e. it will
// not be possible to sign with it as it lacks the secret key.
func (kb dbKeybase) ImportPubKey(name, astr string) (err error) {
	if _, err := kb.GetByNameOrAddress(name); err == nil {
		return errors.New("Cannot overwrite data for name " + name)
	}
	pubBytes, err := armor.UnarmorPubKeyBytes(astr)
	if err != nil {
		return
	}
	pubKey, err := crypto.PubKeyFromBytes(pubBytes)
	if err != nil {
		return
	}
	kb.writeOfflineKey(name, pubKey)
	return
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

// Update changes the passphrase with which an already stored key is
// encrypted.
//
// oldpass must be the current passphrase used for encryption,
// getNewpass is a function to get the passphrase to permanently replace
// the current passphrase
func (kb dbKeybase) Update(nameOrBech32, oldpass string, getNewpass func() (string, error)) error {
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		key, err := armor.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, oldpass)
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

func (kb dbKeybase) writeLocalKey(name string, priv crypto.PrivKey, passphrase string) Info {
	// encrypt private key using passphrase
	privArmor := armor.EncryptArmorPrivKey(priv, passphrase)
	// make Info
	pub := priv.PubKey()
	info := newLocalInfo(name, pub, privArmor)
	kb.writeInfo(name, info)
	return info
}

func (kb dbKeybase) writeLedgerKey(name string, pub crypto.PubKey, path hd.BIP44Params) Info {
	info := newLedgerInfo(name, pub, path)
	kb.writeInfo(name, info)
	return info
}

func (kb dbKeybase) writeOfflineKey(name string, pub crypto.PubKey) Info {
	info := newOfflineInfo(name, pub)
	kb.writeInfo(name, info)
	return info
}

func (kb dbKeybase) writeMultisigKey(name string, pub crypto.PubKey) Info {
	info := NewMultiInfo(name, pub)
	kb.writeInfo(name, info)
	return info
}

func (kb dbKeybase) writeInfo(name string, info Info) {
	// write the info by key
	key := infoKey(name)
	serializedInfo := writeInfo(info)
	kb.db.SetSync(key, serializedInfo)
	// store a pointer to the infokey by address for fast lookup
	kb.db.SetSync(addrKey(info.GetAddress()), key)
}

func addrKey(address crypto.Address) []byte {
	return []byte(fmt.Sprintf("%s.%s", address.String(), addressSuffix))
}

func infoKey(name string) []byte {
	return []byte(fmt.Sprintf("%s.%s", name, infoSuffix))
}
