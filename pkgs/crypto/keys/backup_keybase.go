package keys

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/gnolang/gno/pkgs/amino"
	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/std"

	"github.com/gnolang/gno/pkgs/crypto"

	"github.com/gnolang/gno/pkgs/crypto/bip39"
	gnoEd25519 "github.com/gnolang/gno/pkgs/crypto/ed25519"
	"github.com/gnolang/gno/pkgs/crypto/hd"
	"github.com/gnolang/gno/pkgs/crypto/keys/armor"
	"github.com/gnolang/gno/pkgs/sdk/vm"

	"github.com/gnolang/gno/pkgs/crypto/multisig"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/hkdf"
)

// InfoBK store's back up key in a backup local storage
// TODO: move gno/pkgs/crypto/keys/type.go

// need review for  this infoBk structure
/*
type Info = keys.Info
type Keybase = keys.Keybase
type KeyType = keys.KeyType

const TypeLocal = keys.TypeLocal
*/

// infoBk contains multisig info that is the main property to callers.

type infoBk struct {
	// backup local key information
	Name string `json:"name"` // same as primary key
	// no pubkey field. the pubkey or infoBk is in MultisignInfo

	PrivKeyArmor string `json:"privkey.armor"` // private back key in armored ASCII format
	MultisigInfo Info   `json:"multisig_info"` // Multisig  holds the primary pubkey and back pubkey as a 2/2 multisig
	//  A Secp256k1 signature.
	//  Use the primary priv key sign  the  ecoded JSON string back up info Name + Pubkey(backup)+PrivKeyArmo(backup)
	//  The signature  is to show that infoBk is created by the primary key holder.
	//  It is also verifable if someone change the infoBk record.

	Signature []byte `json:"signature"`

	// this is used to verify the signature signed using primary key secp256k1
	PrimaryPubKey crypto.PubKey `json:"primary_pubkey"`
}

//ask the compiler to check infoBk type implements Info interface

var _ Info = &infoBk{}

func newInfoBk(name string, privArmor string) Info {
	return &infoBk{
		Name: name,

		PrivKeyArmor: privArmor,
	}

}

// GetType implements Info interface
func (i infoBk) GetType() KeyType {
	return TypeLocal
}

// GetType implements Info interface
func (i infoBk) GetName() string {
	return i.Name
}

// GetType implements Info interface
func (i infoBk) GetPubKey() crypto.PubKey {
	return i.MultisigInfo.GetPubKey()
}

// GetType implements Info interface
func (i infoBk) GetAddress() crypto.Address {
	return i.MultisigInfo.GetAddress()
}

// GetType implements Info interface
func (i infoBk) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

//TODO: once reviewed passed, merge this methods to  /pkgs/crypto/keys/keybase.go
func BackupAccount(primaryPrivKey crypto.PrivKey, kbBk Keybase, name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error) {

	coinType := crypto.CoinType
	hdPath := hd.NewFundraiserParams(account, coinType, index)
	//create  a backup info
	info, err := CreateBackupAccountBip44(primaryPrivKey, kbBk, name, mnemonic, bip39Passwd, encryptPasswd, *hdPath)
	return info, err

}

func CreateBackupAccountBip44(primaryPrivKey crypto.PrivKey, kbBk Keybase, name, mnemonic, bip39Passphrase, encryptPasswd string, params hd.BIP44Params) (Info, error) {

	//bip39 uses  PBKDF2 to hash the mnemonic. PBKDF2 is a pass word hash function and not a
	// KDF which provides key extraction and extension
	// at this point the seed is still a seed not a private key yet.
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return infoBk{}, err
	}

	info, err := persistBkKey(primaryPrivKey, kbBk, seed, name, encryptPasswd, params.String())
	return info, err
}

//
// persistBkKey uses primary key to sign the backup key info to key the backup key's integrity

func persistBkKey(primaryPrivKey crypto.PrivKey, kbBk Keybase, seed []byte, name, passwd, fullHdPath string) (Info, error) {

	//The back up key need a simple Sha256 based KDF  which is different from the primary key

	//We can use HKDF for KDF
	//https://rfc-editor.org/rfc/rfc5869.html
	// hkdf does not expect the salt to be a secret and is optional

	/*
		hash := sha256.New
		salt := make([]byte, hash().Size())

		if _, err := rand.Read(salt); err != nil {
			panic(err)
		}
	*/
	//TODO: should we use a hard coded salt?

	var salt []byte
	info := []byte("gnokey hkdf")
	// the size of the privkey is 32 byte
	expanedKeyReader := hkdf.New(sha256.New, seed, salt, info)
	privBkKey := make([]byte, 32)
	if _, err := io.ReadFull(expanedKeyReader, privBkKey); err != nil {

		panic(err)
	}

	// in go/x/crypto/ed25519/ed25519.go
	//This package refers to the RFC8032 private key as the “seed”.

	// ed25519 is used to generate public key from private key
	// the returned a key (64 byte)  =  priveky(32 byte) + pubkey(32 byte)
	// the first 32 byte is private key (see) and  the reset is public key
	bkKey := ed25519.NewKeyFromSeed(privBkKey)

	// cover to  PriveKeyEd25519 type used in gno
	var privKeyEd gnoEd25519.PrivKeyEd25519
	copy(privKeyEd[:], bkKey)

	bkInfo, err := writeLocalBkKey(kbBk, name, privKeyEd, primaryPrivKey, passwd)

	return bkInfo, err
}

//primary key + backup key is a 2/2 multisig threshold pubkey

func writeLocalBkKey(kbBk Keybase, name string, bkKey crypto.PrivKey, primaryKey crypto.PrivKey, passphrase string) (Info, error) {

	//TODO: updated the armored privKey file with correct passwd encryption notaion
	// bcrypt is not KDF. It is a secure hash to protect the password
	privArmor := armor.EncryptArmorPrivKey(bkKey, passphrase)
	pub := bkKey.PubKey() // signle backup Key
	info := newInfoBk(name, privArmor)
	//fmt.Println("back up PubKey", pub)
	//fmt.Println("privArmor", privArmor)

	// sign  name+pubkey+privArmor + multisiginfo

	infobk := info.(*infoBk)
	pubkeys := []crypto.PubKey{
		primaryKey.PubKey(), //primary pubkey
		pub,                 //backup pubkey
	}

	multisig := multisig.NewPubKeyMultisigThreshold(2, pubkeys)

	infobk.MultisigInfo = NewMultiInfo("backup", multisig)

	//TODO: disussion,  could use a document structure. json is simple and good enough  for now.
	msg, err := json.Marshal(infobk)
	//fmt.Println("msg", string(msg))

	//  sign  name + PubKey + PrivKeyArmor + MultisgInfo
	//  To show that the multisig is created by the primary key holder
	infobk.Signature, err = primaryKey.Sign(msg)
	//fmt.Println("Signature", infobk.Signature)

	// attach the primary pubkey in the end. it is used to verify the signature and pubkey
	infobk.PrimaryPubKey = primaryKey.PubKey()
	//fmt.Println("PrimaryPubkey", infobk.PrimaryPubKey.String())

	k := kbBk.(dbKeybase)

	k.writeInfo(name, infobk)
	return info.(*infoBk), err
}

// Sign uses primary key and backup key to sign the message with the multisig
// The primary keybase and backup keybase must be accessible at the same time, which
// is more secure.
// the other option is to sign the the message with priamaryKey and back up KEY seperately.
// since the primary private key is not available at time of siging, the verificatin need to
// only relies on the signature, primary pubkey and information in backup keybase.
// it will introduce attacking oppertunity at the time the messages are combined.
// TODO: A ADR This is also a trade off between usability and security and implementaion complexity.

func signBackup(primaryPriv crypto.PrivKey, backupInfo infoBk, name, passPhrase string, msg []byte) (sig []byte, pub crypto.PubKey, err error) {

	var backupPriv crypto.PrivKey

	// validate

	err = verifyBkInfo(backupInfo, primaryPriv)

	if err != nil {

		return nil, nil, err

	}
	backupPriv, err = armor.UnarmorDecryptPrivKey(backupInfo.PrivKeyArmor, passPhrase)
	if err != nil {
		return nil, nil, err
	}

	//sign the message

	// the signer property of the message is primaryKey

	backupSig, err := backupPriv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	backupPub := backupPriv.PubKey()

	return backupSig, backupPub, nil

}

// verifyBkInfo verify if the info entry in bkKeybase is modify by attackers.
// It checks Pubkey, Signature of infoBk
// TODO: you don't need private key to validate signature with signer's PubKey
//Here is used a shortcut solution since this function is only called by Sign() which has
//privkey at the time calling verifyBkInfo already.

func verifyBkInfo(binfo infoBk, primaryPrivKey crypto.PrivKey) (err error) {

	var mPub multisig.PubKeyMultisigThreshold
	var ok bool

	primaryPubKey := primaryPrivKey.PubKey()
	backupMultiKeys := binfo.GetPubKey()

	if mPub, ok = backupMultiKeys.(multisig.PubKeyMultisigThreshold); ok {

		//check pubkey

		if primaryPubKey.Equals(mPub.PubKeys[0]) == false {

			return fmt.Errorf("pubkey in back up info %v does not match with primary pubkey %v", mPub.PubKeys[0], primaryPubKey)

		}

	} else {

		return fmt.Errorf("backup keybase is compromised: can assert the type %T", backupMultiKeys)
	}

	// check signature

	backupPubKey := mPub.PubKeys[1]

	infoBkSig, err := createInfoBkSignature(primaryPrivKey, primaryPubKey, backupPubKey, binfo.GetName(), binfo.PrivKeyArmor)

	if err != nil {

		return err

	}

	if bytes.Equal(infoBkSig, binfo.Signature) == false {

		return errors.New("infoBk's signature does not match with orignal")
	}

	return nil
}

func createInfoBkSignature(primaryPrivKey crypto.PrivKey, primaryPubKey, backupPubKey crypto.PubKey, name string, privkeyArmor string) (infoBkSig []byte, err error) {

	// create infoBk
	info := newInfoBk(name, privkeyArmor)
	infobk := info.(*infoBk)
	pubkeys := []crypto.PubKey{
		primaryPubKey, //primary pubkey
		backupPubKey,  //backup pubkey
	}

	multisig := multisig.NewPubKeyMultisigThreshold(2, pubkeys)
	infobk.MultisigInfo = NewMultiInfo("backup", multisig)

	//sign  name+pubkey+privArmor + multisiginfo
	msg, err := json.Marshal(infobk)

	infoBkSig, err = primaryPrivKey.Sign(msg)

	return

}

//Todo merge it to keys/uitls.go
const defaultBkKeyDBName = "keys_backup"

func NewBkKeyBaseFromDir(rootDir string) (Keybase, error) {
	//TODO: Remove this after BackupAccount() method are implemented in lazyDBKeybase
	// create data directory and make sure the program has the rwx ownership of data directory
	_ = NewLazyDBKeybase(defaultBkKeyDBName, filepath.Join(rootDir, "data"))

	db, err := dbm.NewGoLevelDB(defaultBkKeyDBName, filepath.Join(rootDir, "data"))
	if err != nil {

		return nil, err
	}

	//defer db.Close()

	return NewDBKeybase(db), nil
}

type SignerInfo struct {
	ChainId       string
	AccountNumber uint64
	Sequence      uint64
}

func SignTx(kbPrimary Keybase, kbBackup Keybase, name, passPhrase string, unsignedTx std.Tx, signerInfo SignerInfo) (signedTx std.Tx, err error) {

	// get primary
	primaryInfo, err := kbPrimary.Get(name)

	if err != nil {

		err = fmt.Errorf("%s not found in primary keybase\n", name)
		return signedTx, err
	}
	var primaryPriv crypto.PrivKey

	switch primaryInfo.(type) {

	case localInfo:
		p := primaryInfo.(localInfo)
		if p.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return signedTx, err
		}

		primaryPriv, err = armor.UnarmorDecryptPrivKey(p.PrivKeyArmor, passPhrase)
		if err != nil {
			return signedTx, err
		}

	case ledgerInfo, offlineInfo, multiInfo:
		err = fmt.Errorf("cannot sign with key %s, only a local key is supported", name)

		return signedTx, err
	}

	// if the backup database is presented in the directory.
	// sign the transaction with back keys
	// TODO: add indicator in primary keybase that a back up key is generated and prompt user to provide
	// bkKeybase if it is presented in the directory

	primaryPub := primaryPriv.PubKey()
	backupInfo, err := kbBackup.Get(name)

	if err != nil {
		err = fmt.Errorf("%s not found in backup keybase. backup your %s first", name, name)
		return signedTx, err
	}
	b := backupInfo.(infoBk)

	multisigInfo := b.MultisigInfo

	// The signature needs to be multisig with sequence. The first is the primary key and second is the back up keys
	// However, account # and sequences # of primary account maybe different from those of backup account.

	multisigPub := multisigInfo.GetPubKey().(multisig.PubKeyMultisigThreshold)
	multisigAddress := multisigInfo.GetAddress()
	multisigSig := multisig.NewMultisig(len(multisigPub.PubKeys))

	var msg std.Msg
	// replace creator to backup key address
	for i := 0; i < len(unsignedTx.Msgs); i++ {
		//TODO: we need to refactor this. we should not check caller and creator for every messages.
		// Is caller's address of a MsgCall also signers or should be address of another smart contract?

		msg = unsignedTx.Msgs[i]

		switch msg.(type) {

		case vm.MsgAddPackage:

			m, ok := msg.(vm.MsgAddPackage)
			if !ok {

				return signedTx, err

			}

			m.Creator = multisigAddress
			msg = m

		case vm.MsgCall:

			m, ok := msg.(vm.MsgCall)
			if !ok {

				return signedTx, err

			}

			m.Caller = multisigAddress
			msg = m

		default:

			return signedTx, fmt.Errorf("Msg type T% is not supported", msg)

		}

		unsignedTx.Msgs[i] = msg

	}

	signbz := unsignedTx.GetSignBytes(signerInfo.ChainId, signerInfo.AccountNumber, signerInfo.Sequence)

	primarySig, err := primaryPriv.Sign(signbz)

	if err != nil {
		return
	}

	backupSig, backupPub, err := signBackup(primaryPriv, b, name, passPhrase, signbz)
	if err != nil {

		return
	}

	err = multisigSig.AddSignatureFromPubKey(primarySig, primaryPub, multisigPub.PubKeys)
	if err != nil {

		return
	}
	err = multisigSig.AddSignatureFromPubKey(backupSig, backupPub, multisigPub.PubKeys)
	if err != nil {

		return
	}

	newStdSig := std.Signature{Signature: amino.MustMarshal(multisigSig), PubKey: multisigPub}

	signedTx = std.Tx{
		Msgs:       unsignedTx.GetMsgs(),
		Fee:        unsignedTx.Fee,
		Signatures: []std.Signature{newStdSig},
		Memo:       unsignedTx.GetMemo(),
	}

	return

}

/*
// Verify verifies the msg signed by primaryKey and backupKey multisig
func Verify(kbBk Keybase, name string, msg []byte, sig []byte) (err error) {

}
*/
