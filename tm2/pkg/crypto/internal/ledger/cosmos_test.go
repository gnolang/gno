package ledger

import (
	"encoding/binary"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// fakeCosmosDevice answers like the Ledger Cosmos app (as ledger-cosmos-go
// drives it) for a single secp256k1 key, and as the transport reports errors.
type fakeCosmosDevice struct {
	version [3]byte
	key     secp256k1.PrivKeySecp256k1
	reject  bool
	apdus   [][]byte
	signing []byte // path and payload of the signature in progress
	open    bool
}

func newFakeCosmosDevice() *fakeCosmosDevice {
	return &fakeCosmosDevice{version: [3]byte{2, 39, 1}, key: secp256k1.GenPrivKey()}
}

func (d *fakeCosmosDevice) address(hrp string) string {
	addr, err := bech32.ConvertAndEncode(hrp, d.key.PubKey().Address().Bytes())
	if err != nil {
		panic(err)
	}
	return addr
}

// checkPath accepts m/44'/118'/a'/c/i in the app's little-endian layout.
func checkCosmosPath(path []byte) error {
	if len(path) != 20 {
		return swError(0x6A80)
	}
	c := func(i int) uint32 { return binary.LittleEndian.Uint32(path[4*i:]) }
	if c(0) != 44|0x8000_0000 || c(1) != 118|0x8000_0000 || c(2) < 0x8000_0000 || c(3) >= 0x8000_0000 || c(4) >= 0x8000_0000 {
		return swError(0x6A80)
	}
	return nil
}

func (d *fakeCosmosDevice) Exchange(apdu []byte) ([]byte, error) {
	d.apdus = append(d.apdus, apdu)
	if apdu[0] != cosmosCLA {
		return nil, swError(0x6E00)
	}
	ins, p1, p2, data := apdu[1], apdu[2], apdu[3], apdu[5:]
	switch ins {
	case cosmosInsGetVersion:
		return []byte{0, d.version[0], d.version[1], d.version[2]}, nil
	case cosmosInsGetAddress:
		hrp := string(data[1 : 1+data[0]])
		if err := checkCosmosPath(data[1+data[0]:]); err != nil {
			return nil, err
		}
		if p1 == 1 && d.reject {
			return nil, swError(0x6986)
		}
		pub := d.key.PubKey().(secp256k1.PubKeySecp256k1)
		return append(pub[:], d.address(hrp)...), nil
	case cosmosInsSign:
		if p2 > cosmosSignModeTextual {
			return nil, swError(0x6B00)
		}
		switch p1 {
		case cosmosP1SignInit:
			if err := checkCosmosPath(data); err != nil {
				return nil, err
			}
			d.signing, d.open = nil, true
			return nil, nil
		case cosmosP1SignAdd, cosmosP1SignLast:
			if !d.open || len(data) > cosmosChunkSize {
				return nil, swError(0x6984)
			}
			d.signing = append(d.signing, data...)
			if p1 == cosmosP1SignAdd {
				return nil, nil
			}
		}
		d.open = false
		if d.reject {
			return nil, swError(0x6986)
		}
		if string(d.signing) == "not json" {
			return nil, &statusError{sw: 0x6A80, data: []byte("PARSER ERROR: JSMN_ERROR_INVAL")}
		}
		priv, _ := btcec.PrivKeyFromBytes(d.key[:])
		return ecdsa.Sign(priv, crypto.Sha256(d.signing)).Serialize(), nil
	}
	return nil, swError(0x6D00)
}

func (*fakeCosmosDevice) Close() error { return nil }

func TestCosmosPath(t *testing.T) {
	t.Parallel()

	got, err := cosmosPath(path)
	require.NoError(t, err)
	want := []byte{}
	for _, c := range []uint32{44 | 0x8000_0000, 118 | 0x8000_0000, 3 | 0x8000_0000, 0, 7} {
		want = binary.LittleEndian.AppendUint32(want, c)
	}
	assert.Equal(t, want, got)

	_, err = cosmosPath(path[:4])
	assert.Error(t, err)
	_, err = cosmosPath([]uint32{44, 118, 3 | 0x8000_0000, 0, 7})
	assert.Error(t, err)
}

func TestOpenCosmosApp(t *testing.T) {
	t.Parallel()

	_, err := openCosmosApp(newFakeCosmosDevice())
	require.NoError(t, err)

	old := newFakeCosmosDevice()
	old.version = [3]byte{2, 0, 9}
	_, err = openCosmosApp(old)
	assert.ErrorContains(t, err, "too old")

	exact := newFakeCosmosDevice()
	exact.version = [3]byte{2, 1, 0}
	_, err = openCosmosApp(exact)
	require.NoError(t, err)

	// The Gno.land app and the dashboard do not know CLA 0x55.
	_, err = openCosmosApp(newFakeGnoDevice())
	assert.ErrorIs(t, err, errNotCosmosApp)

	_, err = openCosmosApp(errorDevice{sw: swDeviceLocked})
	assert.ErrorContains(t, err, "unlock")
}

func TestCosmosAppPublicKey(t *testing.T) {
	t.Parallel()

	d := newFakeCosmosDevice()
	app, err := openCosmosApp(d)
	require.NoError(t, err)
	want := d.key.PubKey().(secp256k1.PubKeySecp256k1)

	pub, err := app.GetPublicKeySECP256K1(path)
	require.NoError(t, err)
	assert.Equal(t, want[:], pub)
	assert.Equal(t, byte(0), d.apdus[len(d.apdus)-1][2])

	pub, addr, err := app.GetAddressPubKeySECP256K1(path, crypto.Bech32AddrPrefix())
	require.NoError(t, err)
	assert.Equal(t, want[:], pub)
	assert.Equal(t, d.address(crypto.Bech32AddrPrefix()), addr)
	assert.Equal(t, byte(1), d.apdus[len(d.apdus)-1][2])

	for _, hrp := range []string{"", "G", "g\x7f"} {
		_, _, err = app.GetAddressPubKeySECP256K1(path, hrp)
		assert.ErrorContains(t, err, "invalid address prefix", "%q", hrp)
	}

	d.reject = true
	_, _, err = app.GetAddressPubKeySECP256K1(path, crypto.Bech32AddrPrefix())
	assert.ErrorContains(t, err, "rejected on the Ledger device")
}

func TestCosmosAppSign(t *testing.T) {
	t.Parallel()

	d := newFakeCosmosDevice()
	app, err := openCosmosApp(d)
	require.NoError(t, err)
	msg := make([]byte, 2*cosmosChunkSize+5)
	for i := range msg {
		msg[i] = byte(i)
	}

	d.apdus = nil
	der, err := app.SignSECP256K1(path, msg, 0)
	require.NoError(t, err)

	// The path, then 48-byte chunks: every APDU fits one HID report.
	require.Len(t, d.apdus, 4)
	for i, want := range []struct{ p1, n byte }{
		{cosmosP1SignInit, 20}, {cosmosP1SignAdd, cosmosChunkSize}, {cosmosP1SignAdd, cosmosChunkSize}, {cosmosP1SignLast, 5},
	} {
		assert.Equal(t, []byte{cosmosCLA, cosmosInsSign, want.p1, 0, want.n}, d.apdus[i][:5], "APDU %d", i)
		assert.Len(t, frameAPDU(d.apdus[i]), 1, "APDU %d", i)
	}

	sig, err := ecdsa.ParseDERSignature(der)
	require.NoError(t, err)
	r, s := sig.R(), sig.S()
	if s.IsOverHalfOrder() {
		s.Negate()
	}
	rb, sb := r.Bytes(), s.Bytes()
	assert.True(t, d.key.PubKey().VerifyBytes(msg, append(rb[:], sb[:]...)))

	_, err = app.SignSECP256K1(path, msg, 2)
	assert.Error(t, err)

	_, err = app.SignSECP256K1(path, []byte("not json"), 0)
	assert.ErrorContains(t, err, "JSMN_ERROR_INVAL")

	d.reject = true
	_, err = app.SignSECP256K1(path, msg, 0)
	assert.ErrorContains(t, err, "rejected on the Ledger device")
}

func TestOpenApp(t *testing.T) {
	t.Parallel()

	app, err := openApp(newFakeGnoDevice())
	require.NoError(t, err)
	assert.IsType(t, &GnoApp{}, app)

	app, err = openApp(newFakeCosmosDevice())
	require.NoError(t, err)
	assert.IsType(t, &CosmosApp{}, app)

	// The dashboard, or another app.
	_, err = openApp(errorDevice{sw: 0x6E00})
	assert.ErrorContains(t, err, "open the Gno.land or the Cosmos app")

	_, err = openApp(errorDevice{sw: swDeviceLocked})
	assert.ErrorContains(t, err, "unlock")

	_, err = openApp(failingDevice{})
	assert.ErrorContains(t, err, "hidapi")
}
