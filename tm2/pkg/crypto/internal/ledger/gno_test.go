package ledger

import (
	"encoding/binary"
	"errors"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// fakeGnoDevice answers like the Gno.land app (docs/APDU.md in ledger-app-gno) for
// a single secp256k1 key, and reports errors as the transport does.
type fakeGnoDevice struct {
	name    string
	key     secp256k1.PrivKeySecp256k1
	busy    int  // commands answered 0x6901 before the next one is processed
	reject  bool // the user rejects every review
	apdus   [][]byte
	session []byte // payload of the open signing session, nil if none
	open    bool
}

func newFakeGnoDevice() *fakeGnoDevice {
	return &fakeGnoDevice{name: GnoAppName, key: secp256k1.GenPrivKey()}
}

// swError is how the transport reports a status word other than 0x9000.
func swError(sw uint16) error {
	return &statusError{sw: sw}
}

func (d *fakeGnoDevice) address() string {
	return d.key.PubKey().Address().String()
}

func (d *fakeGnoDevice) checkPath(data []byte) error {
	if len(data) != 21 || data[0] != 5 {
		return swError(0x6E03)
	}
	// m/44'/118'/account'/0/index
	c := func(i int) uint32 { return binary.BigEndian.Uint32(data[1+4*i:]) }
	if c(0) != 44|0x8000_0000 || c(1) != 118|0x8000_0000 || c(2) < 0x8000_0000 ||
		c(3) != 0 || c(4) >= 0x8000_0000 {
		return swError(0xB00B)
	}
	return nil
}

func (d *fakeGnoDevice) Exchange(apdu []byte) ([]byte, error) {
	d.apdus = append(d.apdus, apdu)
	if d.busy > 0 {
		d.busy--
		return nil, swError(gnoSWCommandNotAccepted)
	}
	if apdu[0] != gnoCLA {
		return nil, swError(0x6E00)
	}
	ins, p1, p2, data := apdu[1], apdu[2], apdu[3], apdu[5:]
	switch ins {
	case gnoInsGetAppName:
		return []byte(d.name), nil
	case gnoInsGetPublicKey:
		if err := d.checkPath(data); err != nil {
			return nil, err
		}
		if p1 == gnoP1PubKeyDisplay && d.reject {
			return nil, swError(0x6985)
		}
		pub := d.key.PubKey().(secp256k1.PubKeySecp256k1)
		resp := append([]byte{gnoPubKeyLen}, pub[:]...)
		resp = append(resp, gnoAddressLen)
		return append(resp, d.address()...), nil
	case gnoInsSignTx:
		switch {
		case p1 == gnoP1SignFirst && p2 == gnoP2SignMore:
			if err := d.checkPath(data); err != nil {
				return nil, err
			}
			d.session, d.open = nil, true
			return nil, nil
		case !d.open:
			return nil, swError(0xB00C)
		case p2 == gnoP2SignMore:
			d.session = append(d.session, data...)
			return nil, nil
		}
		d.open = false
		if d.reject {
			return nil, swError(0x6985)
		}
		return d.key.Sign(append(d.session, data...))
	}
	return nil, swError(0x6D00)
}

func (*fakeGnoDevice) Close() error { return nil }

func openFake(t *testing.T, d *fakeGnoDevice) *GnoApp {
	t.Helper()
	app, err := openGnoApp(d)
	require.NoError(t, err)
	app.sleep = func(time.Duration) {}
	return app
}

// path is m/44'/118'/3'/0/7 as hd.BIP44Params.DerivationPath returns it.
var path = []uint32{44, 118, 3, 0, 7}

func TestGnoPath(t *testing.T) {
	t.Parallel()

	got, err := gnoPath(path)
	require.NoError(t, err)
	want := []byte{5}
	for _, c := range []uint32{44 | 0x8000_0000, 118 | 0x8000_0000, 3 | 0x8000_0000, 0, 7} {
		want = binary.BigEndian.AppendUint32(want, c)
	}
	assert.Equal(t, want, got)

	_, err = gnoPath(path[:4])
	assert.Error(t, err)
	_, err = gnoPath([]uint32{44, 118, 3 | 0x8000_0000, 0, 7})
	assert.Error(t, err)
}

func TestOpenGnoApp(t *testing.T) {
	t.Parallel()

	_, err := openGnoApp(newFakeGnoDevice())
	require.NoError(t, err)

	other := newFakeGnoDevice()
	other.name = "Cosmos"
	_, err = openGnoApp(other)
	assert.ErrorIs(t, err, errNotGnoApp)

	// The Cosmos app and the dashboard do not know CLA 0xE0; other apps answer
	// INS 0x04 with status words of their own.
	for _, sw := range []uint16{0x6E00, 0x6D00, 0x6511, 0x6A80} {
		_, err = openGnoApp(errorDevice{sw: sw})
		assert.ErrorIs(t, err, errNotGnoApp, "0x%04X", sw)
	}

	// A failed transfer is not an answer from another app.
	_, err = openGnoApp(failingDevice{})
	assert.ErrorContains(t, err, "hidapi")
	assert.NotErrorIs(t, err, errNotGnoApp)

	_, err = openGnoApp(errorDevice{sw: swDeviceLocked})
	var status *GnoStatusError
	require.ErrorAs(t, err, &status)
	assert.Equal(t, uint16(swDeviceLocked), status.SW)
	assert.ErrorContains(t, err, "unlock")
}

func TestGnoAppPublicKey(t *testing.T) {
	t.Parallel()

	d := newFakeGnoDevice()
	app := openFake(t, d)
	want := d.key.PubKey().(secp256k1.PubKeySecp256k1)

	pub, err := app.GetPublicKeySECP256K1(path)
	require.NoError(t, err)
	assert.Equal(t, want[:], pub)
	assert.Equal(t, byte(gnoP1PubKeySilent), d.apdus[len(d.apdus)-1][2])

	pub, addr, err := app.GetAddressPubKeySECP256K1(path, crypto.Bech32AddrPrefix())
	require.NoError(t, err)
	assert.Equal(t, want[:], pub)
	assert.Equal(t, d.address(), addr)
	assert.Equal(t, byte(gnoP1PubKeyDisplay), d.apdus[len(d.apdus)-1][2])

	_, _, err = app.GetAddressPubKeySECP256K1(path, "cosmos")
	assert.ErrorContains(t, err, "prefix")
}

func TestGnoAppSign(t *testing.T) {
	t.Parallel()

	d := newFakeGnoDevice()
	app := openFake(t, d)
	// Three chunks: 255, 255 and 90 bytes.
	msg := make([]byte, 2*gnoMaxChunk+90)
	for i := range msg {
		msg[i] = byte(i)
	}

	d.apdus = nil
	der, err := app.SignSECP256K1(path, msg, 0)
	require.NoError(t, err)

	require.Len(t, d.apdus, 4)
	for i, want := range []struct {
		p1, p2 byte
		n      int
	}{
		{gnoP1SignFirst, gnoP2SignMore, 21},
		{gnoP1SignNext, gnoP2SignMore, gnoMaxChunk},
		{gnoP1SignNext, gnoP2SignMore, gnoMaxChunk},
		{gnoP1SignNext, gnoP2SignLast, 90},
	} {
		a := d.apdus[i]
		assert.Equal(t, []byte{gnoCLA, gnoInsSignTx, want.p1, want.p2, byte(want.n)}, a[:5], "APDU %d", i)
		assert.Len(t, a[5:], want.n, "APDU %d", i)
	}

	// The DER signature carries the device's R || S, which tm2 verifies.
	sig, err := ecdsa.ParseDERSignature(der)
	require.NoError(t, err)
	r, s := sig.R(), sig.S()
	rb, sb := r.Bytes(), s.Bytes()
	assert.True(t, d.key.PubKey().VerifyBytes(msg, append(rb[:], sb[:]...)))

	_, err = app.SignSECP256K1(path, msg, 1)
	assert.Error(t, err)
	_, err = app.SignSECP256K1(path, nil, 0)
	assert.Error(t, err)
}

func TestGnoAppStatusWords(t *testing.T) {
	t.Parallel()

	d := newFakeGnoDevice()
	app := openFake(t, d)
	d.reject = true

	_, err := app.SignSECP256K1(path, []byte("payload"), 0)
	var status *GnoStatusError
	require.ErrorAs(t, err, &status)
	assert.Equal(t, uint16(0x6985), status.SW)
	assert.ErrorContains(t, err, "rejected on the device")

	_, _, err = app.GetAddressPubKeySECP256K1(path, crypto.Bech32AddrPrefix())
	require.ErrorAs(t, err, &status)
	assert.Equal(t, uint16(0x6985), status.SW)

	_, err = app.GetPublicKeySECP256K1([]uint32{44, 118, 3, 1, 7})
	require.ErrorAs(t, err, &status)
	assert.Equal(t, uint16(0xB00B), status.SW)
}

func TestGnoAppRetriesWhileBusy(t *testing.T) {
	t.Parallel()

	d := newFakeGnoDevice()
	app := openFake(t, d)

	d.busy = 3
	d.apdus = nil
	_, err := app.GetPublicKeySECP256K1(path)
	require.NoError(t, err)
	assert.Len(t, d.apdus, 4)

	// Past the timeout, the busy status is returned.
	d.busy = 1000
	_, err = app.GetPublicKeySECP256K1(path)
	var status *GnoStatusError
	require.ErrorAs(t, err, &status)
	assert.Equal(t, uint16(gnoSWCommandNotAccepted), status.SW)
	assert.Len(t, d.apdus, 4+1+int(app.busyTimeout/app.busyRetry))
}

func TestGnoAppMalformedResponses(t *testing.T) {
	t.Parallel()

	app := openFake(t, newFakeGnoDevice())
	app.device = replyDevice{reply: []byte{gnoPubKeyLen, 1, 2, 3}}
	_, err := app.GetPublicKeySECP256K1(path)
	assert.ErrorContains(t, err, "malformed")

	app.device = replyDevice{reply: make([]byte, 63)}
	_, err = app.SignSECP256K1(path, []byte("payload"), 0)
	assert.ErrorContains(t, err, "malformed")

	app.device = replyDevice{reply: make([]byte, 64)}
	_, err = app.SignSECP256K1(path, []byte("payload"), 0)
	assert.ErrorContains(t, err, "out of range")
}

// errorDevice answers every command with the same status word.
type errorDevice struct{ sw uint16 }

func (d errorDevice) Exchange([]byte) ([]byte, error) { return nil, swError(d.sw) }
func (errorDevice) Close() error                      { return nil }

// failingDevice fails every transfer, as an unresponsive device does.
type failingDevice struct{}

func (failingDevice) Exchange([]byte) ([]byte, error) {
	return nil, errors.New("hidapi: unknown failure")
}
func (failingDevice) Close() error { return nil }

// replyDevice answers every command with the same reply.
type replyDevice struct{ reply []byte }

func (d replyDevice) Exchange([]byte) ([]byte, error) { return d.reply, nil }
func (replyDevice) Close() error                      { return nil }
