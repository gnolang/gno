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
	has, err = cstore.HasByAddress(toAddr(i2))
	require.NoError(t, err)
	require.True(t, has)
	// Also check with HasByNameOrAddress
	has, err = cstore.HasByNameOrAddress(crypto.AddressToBech32(toAddr(i2)))
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

	// Import a public key
	armor, err := cstore.ExportPubKey(n2)
	require.Nil(t, err)
	cstore.ImportPubKey(n3, armor)
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
	err := cstore.Update(name, badpass, getNewpass)
	require.NotNil(t, err)
	err = cstore.Update(name, pass, getNewpass)
	require.Nil(t, err, "%+v", err)
}

// TestExportImport tests exporting and importing
func TestExportImport(t *testing.T) {
	t.Parallel()

	// make the storage with reasonable defaults
	cstore := NewInMemory()

	mn1 := `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`
	bip39Passphrase := ""

	info, err := cstore.CreateAccount("john", mn1, bip39Passphrase, "secretcpw", 0, 0)
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")

	john, err := cstore.GetByName("john")
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")
	johnAddr := info.GetPubKey().Address()

	armor, err := cstore.Export("john")
	require.NoError(t, err)

	err = cstore.Import("john2", armor)
	require.NoError(t, err)

	john2, err := cstore.GetByName("john2")
	require.NoError(t, err)

	require.Equal(t, john.GetPubKey().Address(), johnAddr)
	require.Equal(t, john.GetName(), "john")
	require.Equal(t, john, john2)
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

	// Export the public key only
	armor, err := cstore.ExportPubKey("john")
	require.NoError(t, err)
	// Import it under a different name
	err = cstore.ImportPubKey("john-pubkey-only", armor)
	require.NoError(t, err)
	// Ensure consistency
	john2, err := cstore.GetByName("john-pubkey-only")
	require.NoError(t, err)
	// Compare the public keys
	require.True(t, john.GetPubKey().Equals(john2.GetPubKey()))
	// Ensure the original key hasn't changed
	john, err = cstore.GetByName("john")
	require.NoError(t, err)
	require.Equal(t, john.GetPubKey().Address(), addr)
	require.Equal(t, john.GetName(), "john")

	// Ensure keys cannot be overwritten
	err = cstore.ImportPubKey("john-pubkey-only", armor)
	require.NotNil(t, err)
}

// TestAdvancedKeyManagement verifies update, import, export functionality
func TestAdvancedKeyManagement(t *testing.T) {
	t.Parallel()

	// make the storage with reasonable defaults
	cstore := NewInMemory()

	n1, n2 := "old-name", "new name"
	p1, p2 := "1234", "foobar"
	mn1 := `lounge napkin all odor tilt dove win inject sleep jazz uncover traffic hint require cargo arm rocket round scan bread report squirrel step lake`
	bip39Passphrase := ""

	// make sure key works with initial password
	_, err := cstore.CreateAccount(n1, mn1, bip39Passphrase, p1, 0, 0)
	require.Nil(t, err, "%+v", err)
	assertPassword(t, cstore, n1, p1, p2)

	// update password requires the existing password
	getNewpass := func() (string, error) { return p2, nil }
	err = cstore.Update(n1, "jkkgkg", getNewpass)
	require.NotNil(t, err)
	assertPassword(t, cstore, n1, p1, p2)

	// then it changes the password when correct
	err = cstore.Update(n1, p1, getNewpass)
	require.NoError(t, err)
	// p2 is now the proper one!
	assertPassword(t, cstore, n1, p2, p1)

	// exporting requires the proper name and passphrase
	_, err = cstore.Export(n1 + ".notreal")
	require.NotNil(t, err)
	_, err = cstore.Export(" " + n1)
	require.NotNil(t, err)
	_, err = cstore.Export(n1 + " ")
	require.NotNil(t, err)
	_, err = cstore.Export("")
	require.NotNil(t, err)
	exported, err := cstore.Export(n1)
	require.Nil(t, err, "%+v", err)

	// import succeeds
	err = cstore.Import(n2, exported)
	require.NoError(t, err)

	// second import fails
	err = cstore.Import(n2, exported)
	require.NotNil(t, err)
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

func toAddr(info Info) crypto.Address {
	return info.GetPubKey().Address()
}
