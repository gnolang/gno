package keys

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/crypto/hd"
	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/os"
)

var _ Keybase = lazyKeybase{}

type lazyKeybase struct {
	name string
	dir  string
}

// New creates a new instance of a lazy keybase.
func NewLazyDBKeybase(name, dir string) Keybase {
	if err := os.EnsureDir(dir, 0o700); err != nil {
		panic(fmt.Sprintf("failed to create Keybase directory: %s", err))
	}

	return lazyKeybase{name: name, dir: dir}
}

func (lkb lazyKeybase) List() ([]Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).List()
}

func (lkb lazyKeybase) GetByNameOrAddress(nameOrBech32 string) (Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).GetByNameOrAddress(nameOrBech32)
}

func (lkb lazyKeybase) GetByName(name string) (Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).GetByName(name)
}

func (lkb lazyKeybase) GetByAddress(address crypto.Address) (Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).GetByAddress(address)
}

func (lkb lazyKeybase) Delete(name, passphrase string, skipPass bool) error {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return NewDBKeybase(db).Delete(name, passphrase, skipPass)
}

func (lkb lazyKeybase) Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).Sign(name, passphrase, msg)
}

func (lkb lazyKeybase) Verify(name string, msg, sig []byte) error {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return NewDBKeybase(db).Verify(name, msg, sig)
}

func (lkb lazyKeybase) CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd, account, index)
}

func (lkb lazyKeybase) CreateAccountBip44(name, mnemonic, bip39Passwd, encryptPasswd string, params hd.BIP44Params) (Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).CreateAccountBip44(name, mnemonic, bip39Passwd, encryptPasswd, params)
}

func (lkb lazyKeybase) CreateLedger(name string, algo SigningAlgo, hrp string, account, index uint32) (info Info, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).CreateLedger(name, algo, hrp, account, index)
}

func (lkb lazyKeybase) CreateOffline(name string, pubkey crypto.PubKey) (info Info, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).CreateOffline(name, pubkey)
}

func (lkb lazyKeybase) CreateMulti(name string, pubkey crypto.PubKey) (info Info, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).CreateMulti(name, pubkey)
}

func (lkb lazyKeybase) Update(name, oldpass string, getNewpass func() (string, error)) error {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return NewDBKeybase(db).Update(name, oldpass, getNewpass)
}

func (lkb lazyKeybase) Import(name string, armor string) (err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return NewDBKeybase(db).Import(name, armor)
}

func (lkb lazyKeybase) ImportPrivKey(name string, armor string, passphrase string) error {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return NewDBKeybase(db).ImportPrivKey(name, armor, passphrase)
}

func (lkb lazyKeybase) ImportPubKey(name string, armor string) (err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return NewDBKeybase(db).ImportPubKey(name, armor)
}

func (lkb lazyKeybase) Export(name string) (armor string, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return "", err
	}
	defer db.Close()

	return NewDBKeybase(db).Export(name)
}

func (lkb lazyKeybase) ExportPubKey(name string) (armor string, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return "", err
	}
	defer db.Close()

	return NewDBKeybase(db).ExportPubKey(name)
}

func (lkb lazyKeybase) ExportPrivateKeyObject(name string, passphrase string) (crypto.PrivKey, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return NewDBKeybase(db).ExportPrivateKeyObject(name, passphrase)
}

func (lkb lazyKeybase) ExportPrivKey(name string, decryptPassphrase string,
	encryptPassphrase string,
) (armor string, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return "", err
	}
	defer db.Close()

	return NewDBKeybase(db).ExportPrivKey(name, decryptPassphrase, encryptPassphrase)
}

func (lkb lazyKeybase) CloseDB() {}
