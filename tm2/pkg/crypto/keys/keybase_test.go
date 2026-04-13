package keys

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
)

const testMnemonic = `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`

func TestRename(t *testing.T) {
	t.Parallel()

	t.Run("successful rename", func(t *testing.T) {
		t.Parallel()

		cstore := NewInMemory()

		info, err := cstore.CreateAccount("original", testMnemonic, "", "password", 0, 0)
		require.NoError(t, err)

		err = cstore.Rename("original", "renamed")
		require.NoError(t, err)

		// New name should exist with same address
		got, err := cstore.GetByName("renamed")
		require.NoError(t, err)
		assert.Equal(t, info.GetAddress(), got.GetAddress())

		// Old name should not exist
		_, err = cstore.GetByName("original")
		require.Error(t, err)
	})

	t.Run("rename non-existent key", func(t *testing.T) {
		t.Parallel()

		cstore := NewInMemory()

		err := cstore.Rename("non-existent", "new-name")
		require.Error(t, err)
	})

	t.Run("rename to existing name", func(t *testing.T) {
		t.Parallel()

		cstore := NewInMemory()

		_, err := cstore.CreateAccount("key1", testMnemonic, "", "password", 0, 0)
		require.NoError(t, err)

		_, err = cstore.CreateAccount("key2", testMnemonic, "", "password", 0, 1)
		require.NoError(t, err)

		err = cstore.Rename("key1", "key2")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("rename offline key", func(t *testing.T) {
		t.Parallel()

		cstore := NewInMemory()

		key := ed25519.GenPrivKey()
		info, err := cstore.CreateOffline("offline-key", key.PubKey())
		require.NoError(t, err)

		err = cstore.Rename("offline-key", "renamed-offline")
		require.NoError(t, err)

		got, err := cstore.GetByName("renamed-offline")
		require.NoError(t, err)
		assert.Equal(t, info.GetAddress(), got.GetAddress())

		_, err = cstore.GetByName("offline-key")
		require.Error(t, err)
	})
}

func TestCreateAccountInvalidMnemonic(t *testing.T) {
	t.Parallel()

	kb := NewInMemory()
	_, err := kb.CreateAccount(
		"some_account",
		"malarkey pair crucial catch public canyon evil outer stage ten gym tornado",
		"", "", 0, 1)
	assert.Error(t, err)
	assert.Equal(t, "invalid mnemonic", err.Error())
}

// TestKeyManagement makes sure we can manipulate these keys well
func TestKeyManagement(t *testing.T) {
	t.Parallel()

	// make the storage with reasonable defaults
	cstore := NewInMemory()

	n1, n2, n3 := "personal", "business", "other"
	p1, p2 := "1234", "really-secure!@#$"
	mn1 := `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`
	mn2 := `lecture salt about avocado smooth height escape general arch head barrel clutch dismiss supply doctor project cat truck fruit abuse gorilla symbol portion glare`
	bip39Passphrase := ""

	// Check empty state
	l, err := cstore.List()
	require.Nil(t, err)
	assert.Empty(t, l)

	// create some keys
	has, err := cstore.HasByName(n1)
	require.NoError(t, err)
	require.False(t, has)
	i, err := cstore.CreateAccount(n1, mn1, bip39Passphrase, p1, 0, 0)
	require.NoError(t, err)
	require.Equal(t, n1, i.GetName())
	_, err = cstore.CreateAccount(n2, mn2, bip39Passphrase, p2, 0, 0)
	require.NoError(t, err)

	// we can get these keys
	i2, err := cstore.GetByName(n2)
	require.NoError(t, err)
	has, err = cstore.HasByName(n3)
	require.NoError(t, err)
	require.False(t, has)
	has, err = cstore.HasByAddress(i2.GetPubKey().Address())
	require.NoError(t, err)
	require.True(t, has)
	// Also check with HasByNameOrAddress
	has, err = cstore.HasByNameOrAddress(crypto.AddressToBech32(i2.GetPubKey().Address()))
	require.NoError(t, err)
	require.True(t, has)
	addr, err := crypto.AddressFromBech32("g1frtkxv37nq7arvyz5p0mtjqq7hwuvd4dnt892p")
	require.NoError(t, err)
	_, err = cstore.GetByAddress(addr)
	require.NotNil(t, err)
	require.True(t, keyerror.IsErrKeyNotFound(err))

	// list shows them in order
	keyS, err := cstore.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))
	// note these are in alphabetical order
	require.Equal(t, n2, keyS[0].GetName())
	require.Equal(t, n1, keyS[1].GetName())
	require.Equal(t, i2.GetPubKey(), keyS[0].GetPubKey())

	// deleting a key removes it
	err = cstore.Delete("bad name", "foo", false)
	require.NotNil(t, err)
	err = cstore.Delete(n1, p1, false)
	require.NoError(t, err)
	keyS, err = cstore.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))
	has, err = cstore.HasByName(n1)
	require.NoError(t, err)
	require.False(t, has)

	// create an offline key
	o1 := "offline"
	priv1 := ed25519.GenPrivKey()
	pub1 := priv1.PubKey()
	i, err = cstore.CreateOffline(o1, pub1)
	require.Nil(t, err)
	require.Equal(t, pub1, i.GetPubKey())
	require.Equal(t, o1, i.GetName())
	keyS, err = cstore.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))

	// delete the offline key
	err = cstore.Delete(o1, "", false)
	require.NoError(t, err)
	keyS, err = cstore.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))

	// Lookup by original i2 address
	infoByAddress, err := cstore.GetByAddress(i2.GetAddress())
	require.NoError(t, err)
	// GetByAddress should return Info with the corresponding public key
	require.Equal(t, infoByAddress.GetPubKey(), i2.GetPubKey())
	// Replace n2 with a new address
	mn2New := `fancy assault crane note start invite ladder ordinary gold amateur check cousin text mercy speak chuckle wine raw chief isolate swallow cushion wrist piece`
	_, err = cstore.CreateAccount(n2, mn2New, bip39Passphrase, p2, 0, 0)
	require.NoError(t, err)
	// Check that CreateAccount removes the entry for the original address (public key)
	_, err = cstore.GetByAddress(i2.GetAddress())
	require.NotNil(t, err)

	// addr cache gets nuked - and test skip flag
	err = cstore.Delete(n2, "", true)
	require.NoError(t, err)
}

// TestSignVerify does some detailed checks on how we sign and validate
// signatures
func TestSignVerify(t *testing.T) {
	t.Parallel()

	cstore := NewInMemory()

	n1, n2, n3 := "some dude", "a dudette", "dude-ish"
	p1, p2, p3 := "1234", "foobar", "foobar"
	mn1 := `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`
	mn2 := `lecture salt about avocado smooth height escape general arch head barrel clutch dismiss supply doctor project cat truck fruit abuse gorilla symbol portion glare`
	bip39Passphrase := ""

	// create two users and get their info
	i1, err := cstore.CreateAccount(n1, mn1, bip39Passphrase, p1, 0, 0)
	require.Nil(t, err)

	i2, err := cstore.CreateAccount(n2, mn2, bip39Passphrase, p2, 0, 0)
	require.Nil(t, err)

	i3Key := ed25519.GenPrivKey()

	// Import a public key
	_, err = cstore.CreateOffline(n3, i3Key.PubKey())
	require.NoError(t, err)
	i3, err := cstore.GetByName(n3)
	require.NoError(t, err)
	require.Equal(t, i3.GetName(), n3)

	// let's try to sign some messages
	d1 := []byte("my first message")
	d2 := []byte("some other important info!")
	d3 := []byte("feels like I forgot something...")

	// try signing both data with both ..
	s11, pub1, err := cstore.Sign(n1, p1, d1)
	require.Nil(t, err)
	require.Equal(t, i1.GetPubKey(), pub1)

	s12, pub1, err := cstore.Sign(n1, p1, d2)
	require.Nil(t, err)
	require.Equal(t, i1.GetPubKey(), pub1)

	s21, pub2, err := cstore.Sign(n2, p2, d1)
	require.Nil(t, err)
	require.Equal(t, i2.GetPubKey(), pub2)
	require.Equal(t, i3.GetPubKey(), i3Key.PubKey())

	s22, pub2, err := cstore.Sign(n2, p2, d2)
	require.Nil(t, err)
	require.Equal(t, i2.GetPubKey(), pub2)

	// let's try to validate and make sure it only works when everything is proper
	cases := []struct {
		key   crypto.PubKey
		data  []byte
		sig   []byte
		valid bool
	}{
		// proper matches
		{i1.GetPubKey(), d1, s11, true},
		// change data, pubkey, or signature leads to fail
		{i1.GetPubKey(), d2, s11, false},
		{i2.GetPubKey(), d1, s11, false},
		{i1.GetPubKey(), d1, s21, false},
		// make sure other successes
		{i1.GetPubKey(), d2, s12, true},
		{i2.GetPubKey(), d1, s21, true},
		{i2.GetPubKey(), d2, s22, true},
	}

	for i, tc := range cases {
		valid := tc.key.VerifyBytes(tc.data, tc.sig)
		require.Equal(t, tc.valid, valid, "%d", i)
	}

	// Now try to sign data with a secret-less key
	_, _, err = cstore.Sign(n3, p3, d3)
	require.NotNil(t, err)
}

func assertPassword(t *testing.T, cstore Keybase, name, pass, badpass string) {
	t.Helper()

	getNewpass := func() (string, error) { return pass, nil }
	err := cstore.Rotate(name, badpass, getNewpass)
	require.NotNil(t, err)
	err = cstore.Rotate(name, pass, getNewpass)
	require.Nil(t, err, "%+v", err)
}

func TestExportImportPubKey(t *testing.T) {
	t.Parallel()

	// make the storage with reasonable defaults
	cstore := NewInMemory()

	// CreateAccount a private-public key pair and ensure consistency
	mn1 := `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`
	bip39Passphrase := ""
	notPasswd := "n9y25ah7"
	info, err := cstore.CreateAccount("john", mn1, bip39Passphrase, notPasswd, 0, 0)
	require.Nil(t, err)
	require.NotEqual(t, info, "")
	require.Equal(t, info.GetName(), "john")
	addr := info.GetPubKey().Address()
	john, err := cstore.GetByName("john")
	require.NoError(t, err)
	require.Equal(t, john.GetName(), "john")
	require.Equal(t, john.GetPubKey().Address(), addr)

	// Import it under a different name
	_, err = cstore.CreateOffline("john-pubkey-only", john.GetPubKey())
	require.NoError(t, err)
	// Ensure consistency
	john2, err := cstore.GetByName("john-pubkey-only")
	require.NoError(t, err)
	// Compare the public keys
	require.True(t, john.GetPubKey().Equals(john2.GetPubKey()))
	// Ensure that storing with the address of "john-pubkey-only" removed the entry for "john"
	has, err := cstore.HasByName("john")
	require.NoError(t, err)
	require.False(t, has)
}

// TestAdvancedKeyManagement verifies rotate functionality
func TestAdvancedKeyManagement(t *testing.T) {
	t.Parallel()

	// make the storage with reasonable defaults
	cstore := NewInMemory()

	n1 := "old-name"
	p1, p2 := "1234", "foobar"
	mn1 := `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`
	bip39Passphrase := ""

	// make sure key works with initial password
	_, err := cstore.CreateAccount(n1, mn1, bip39Passphrase, p1, 0, 0)
	require.Nil(t, err, "%+v", err)
	assertPassword(t, cstore, n1, p1, p2)

	// rotate password requires the existing password
	getNewpass := func() (string, error) { return p2, nil }
	err = cstore.Rotate(n1, "jkkgkg", getNewpass)
	require.NotNil(t, err)
	assertPassword(t, cstore, n1, p1, p2)

	// then it changes the password when correct
	err = cstore.Rotate(n1, p1, getNewpass)
	require.NoError(t, err)
	// p2 is now the proper one!
	assertPassword(t, cstore, n1, p2, p1)
}

// TestSeedPhrase verifies restoring from a seed phrase
func TestSeedPhrase(t *testing.T) {
	t.Parallel()

	// make the storage with reasonable defaults
	cstore := NewInMemory()

	n1 := "lost-key"
	p1 := "1234"
	mn1 := `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`
	bip39Passphrase := ""

	// make sure key works with initial password
	info, err := cstore.CreateAccount(n1, mn1, bip39Passphrase, p1, 0, 0)
	require.Nil(t, err, "%+v", err)
	require.Equal(t, n1, info.GetName())

	// now, let us delete this key
	err = cstore.Delete(n1, p1, false)
	require.Nil(t, err, "%+v", err)
	has, err := cstore.HasByName(n1)
	require.NoError(t, err)
	require.False(t, has)
}

func ExampleNew() {
	// Select the encryption and storage for your cryptostore
	cstore := NewInMemory()

	mn1 := `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`
	mn2 := `lecture salt about avocado smooth height escape general arch head barrel clutch dismiss supply doctor project cat truck fruit abuse gorilla symbol portion glare`
	mn3 := `jar nest rug lion shallow spring abuse west gravity skin project comic again dirt pelican better galaxy click hold lottery swap solution census own`
	bip39Passphrase := ""

	// Add keys and see they return in alphabetical order
	bob, err := cstore.CreateAccount("Bob", mn1, bip39Passphrase, "friend", 0, 0)
	if err != nil {
		// this should never happen
		fmt.Println(err)
	} else {
		// return info here just like in List
		fmt.Println(bob.GetName())
	}
	_, _ = cstore.CreateAccount("Alice", mn2, bip39Passphrase, "secret", 0, 0)
	_, _ = cstore.CreateAccount("Carl", mn3, bip39Passphrase, "mitm", 0, 0)
	info, _ := cstore.List()
	for _, i := range info {
		fmt.Println(i.GetName())
	}

	// We need to use passphrase to generate a signature
	tx := []byte("deadbeef")
	sig, pub, err := cstore.Sign("Bob", "friend", tx)
	if err != nil {
		fmt.Println("don't accept real passphrase")
	}

	// and we can validate the signature with publicly available info
	binfo, _ := cstore.GetByName("Bob")
	if !binfo.GetPubKey().Equals(bob.GetPubKey()) {
		fmt.Println("Get and Create return different keys")
	}

	if pub.Equals(binfo.GetPubKey()) {
		fmt.Println("signed by Bob")
	}
	if !pub.VerifyBytes(tx, sig) {
		fmt.Println("invalid signature")
	}

	// Output:
	// Bob
	// Alice
	// Bob
	// Carl
	// signed by Bob
}

func TestKeybase_ImportPrivKey(t *testing.T) {
	t.Parallel()

	t.Run("unable to overwrite key", func(t *testing.T) {
		t.Parallel()

		var (
			cstore      = NewInMemory()
			privKey     = ed25519.GenPrivKey()
			name        = "key-name"
			encryptPass = "password"
		)

		// Import the private key
		require.NoError(t, cstore.ImportPrivKey(name, privKey, encryptPass))

		// Attempt to import a key with the same name
		assert.ErrorIs(
			t,
			cstore.ImportPrivKey(name, ed25519.GenPrivKey(), encryptPass),
			errCannotOverwrite,
		)
	})

	t.Run("valid key import", func(t *testing.T) {
		t.Parallel()

		var (
			cstore      = NewInMemory()
			privKey     = ed25519.GenPrivKey()
			name        = "key-name"
			encryptPass = "password"
		)

		// Import the private key
		require.NoError(t, cstore.ImportPrivKey(name, privKey, encryptPass))

		// Make sure the key is present
		info, err := cstore.GetByName(name)
		require.NoError(t, err)

		assert.Equal(t, name, info.GetName())
		assert.True(t, privKey.PubKey().Equals(info.GetPubKey()))
	})
}

func TestKeybase_ExportPrivKey(t *testing.T) {
	t.Parallel()

	t.Run("missing key", func(t *testing.T) {
		t.Parallel()

		var (
			cstore      = NewInMemory()
			name        = "key-name"
			decryptPass = "password"
		)

		keys, err := cstore.List()
		require.NoError(t, err)

		// Make sure the keybase is empty
		require.Empty(t, keys)

		// Attempt to export a missing key
		_, err = cstore.ExportPrivKey(name, decryptPass)
		assert.True(t, keyerror.IsErrKeyNotFound(err))
	})

	t.Run("valid key export", func(t *testing.T) {
		t.Parallel()

		var (
			cstore      = NewInMemory()
			name        = "key-name"
			key         = ed25519.GenPrivKey()
			encryptPass = "password"
		)

		// Add the key
		require.NoError(t, cstore.ImportPrivKey(name, key, encryptPass))

		// Export the key
		exportedKey, err := cstore.ExportPrivKey(name, encryptPass)
		require.NoError(t, err)

		assert.True(t, key.Equals(exportedKey))
	})
}
