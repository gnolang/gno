package keys

import (

	"testing"

	"github.com/gnolang/gno/pkgs/crypto/keys/armor"
	"github.com/stretchr/testify/assert"
)

const key_name = "test1"
const test1_passcode = "test1rocks"

const test1_mnemonic = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
const primary_pubkey = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj"

const test1bk_mnemonic = "curious syrup memory cabbage razor emotion ketchup best alley cotton enjoy nature furnace shallow donor oval tornado razor clock roof pave enroll solar wrist"
const backup_pubkey = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp7xtkykvttvcxnz9n74hfd8t4tav3t7l33p5trvyeuxd3ea8d95vhp767p"

func TestBackupAccount(t *testing.T) {

	kb := NewInMemory()
	kbBk := NewInMemory()

	primaryInfo, err := kb.CreateAccount(
		key_name,
		test1_mnemonic,
		"", test1_passcode, 0, 1)

	assert.NoError(t, err, "create primary account failed")

	p, ok := primaryInfo.(*localInfo)

	assert.True(t, ok, "primaryInfo should be localInfo")
  assert.NotNil(t, p, "primaryInfo should not be nil")

	primaryPrivKey, err := armor.UnarmorDecryptPrivKey(p.PrivKeyArmor, test1_passcode)


	assert.NoError(t, err, "read primary localInfo.PrivKeyArmor failed")

	assert.NotNil(t, primaryPrivKey, "primaryPrivKey should not be nil")


	info, err := BackupAccount(primaryPrivKey, kbBk, key_name, test1_mnemonic, "", test1_passcode, 0, 1)
	assert.NoError(t, err, "creating backup info failed")

	assert.Equal(t, primaryInfo.GetName(), info.GetName(), "Names are equal")

	infoBackup, ok := info.(*infoBk)
	assert.True(t, ok, "info should be infoBk")
  assert.NotNil(t, p, "info should not be nil")

	err = verifyBkInfo(*infoBackup, primaryPrivKey)

	assert.NotNil(t, primaryPrivKey, "BkInfo is not corrected created")

}
