package ledger

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	ledger_go "github.com/zondax/ledger-go"
)

// GnoAppName is the name the Gno Ledger app returns for GET_APP_NAME.
const GnoAppName = "Gno"

// APDU protocol of the Gno Ledger app (github.com/clockworkgr/ledger-app-gno,
// docs/APDU.md).
const (
	gnoCLA = 0xE0

	gnoInsGetAppName   = 0x04
	gnoInsGetPublicKey = 0x05
	gnoInsSignTx       = 0x06

	gnoP1PubKeySilent  = 0x00
	gnoP1PubKeyDisplay = 0x01
	gnoP1SignFirst     = 0x00
	gnoP1SignNext      = 0x01
	gnoP2SignLast      = 0x00
	gnoP2SignMore      = 0x80

	gnoMaxChunk     = 255
	gnoPubKeyLen    = 33
	gnoAddressLen   = 40
	gnoSignatureLen = 64
)

// Status words the Gno app returns, besides 0x9000.
const (
	gnoSWCommandNotAccepted = 0x6901
	swDeviceLocked          = 0x5515
)

var gnoStatusWords = map[uint16]string{
	swDeviceLocked:          "the device is locked: unlock it",
	gnoSWCommandNotAccepted: "the device is busy with the previous command",
	0x6985:                  "rejected on the device",
	0x6A86:                  "invalid P1/P2",
	0x6D00:                  "instruction not supported",
	0x6E00:                  "CLA not supported",
	0x6E03:                  "malformed data length",
	0xB002:                  "the address could not be displayed",
	0xB005:                  "the transaction is not one the Gno app accepts",
	0xB008:                  "signing failed",
	0xB009:                  "key derivation failed",
	0xB00A:                  "the app version could not be encoded",
	0xB00B:                  "the path is outside m/44'/118'/account'/0/index",
	0xB00C:                  "no signing session is open",
	0xB00D:                  "a message is signed by an address other than the device's",
	0xB00E:                  "a displayed value or the review exceeds the device's limits",
	0xB00F:                  "the transaction can run code: enable \"Packages\" in the Gno app's settings",
}

// GnoStatusError is a status word other than 0x9000 returned by the Gno app
// or by the device.
type GnoStatusError struct {
	SW uint16
}

func (e *GnoStatusError) Error() string {
	return fmt.Sprintf("Ledger: %s (0x%04X)", gnoStatusWords[e.SW], e.SW)
}

// gnoStatus recovers the status word behind an error from ledger-go, which
// reports only the text of ErrorMessage(sw).
func gnoStatus(err error) (uint16, bool) {
	for sw := range gnoStatusWords {
		if err.Error() == ledger_go.ErrorMessage(sw) {
			return sw, true
		}
	}
	return 0, false
}

// GnoApp talks to the Gno Ledger app. It implements [SECP256K1].
type GnoApp struct {
	device ledger_go.LedgerDevice

	// After a review, the device answers 0x6901 until its status page closes.
	// Such commands were not processed, so they are resent.
	busyRetry   time.Duration
	busyTimeout time.Duration
	sleep       func(time.Duration)
}

var errNotGnoApp = errors.New("the open Ledger app is not the Gno app")

// openGnoApp returns a [GnoApp] on device if the Gno app is the one open,
// and errNotGnoApp if another app or none is.
func openGnoApp(device ledger_go.LedgerDevice) (*GnoApp, error) {
	app := &GnoApp{
		device:      device,
		busyRetry:   250 * time.Millisecond,
		busyTimeout: 10 * time.Second,
		sleep:       time.Sleep,
	}
	name, err := app.exchange(gnoInsGetAppName, 0, 0, nil)
	var status *GnoStatusError
	if errors.As(err, &status) && status.SW == swDeviceLocked {
		return nil, err
	}
	if err != nil || string(name) != GnoAppName {
		return nil, errNotGnoApp
	}
	return app, nil
}

func (a *GnoApp) exchange(ins, p1, p2 byte, data []byte) ([]byte, error) {
	if len(data) > gnoMaxChunk {
		return nil, fmt.Errorf("APDU data of %d bytes exceeds %d", len(data), gnoMaxChunk)
	}
	apdu := append([]byte{gnoCLA, ins, p1, p2, byte(len(data))}, data...)

	var waited time.Duration
	for {
		resp, err := a.device.Exchange(apdu)
		if err == nil {
			return resp, nil
		}
		sw, ok := gnoStatus(err)
		if !ok {
			return nil, err
		}
		if sw == gnoSWCommandNotAccepted && waited < a.busyTimeout {
			a.sleep(a.busyRetry)
			waited += a.busyRetry
			continue
		}
		return nil, &GnoStatusError{SW: sw}
	}
}

// gnoPath encodes a BIP44 path given as unhardened components, the form
// [hd.BIP44Params.DerivationPath] returns. The first three are hardened.
func gnoPath(path []uint32) ([]byte, error) {
	if len(path) != 5 {
		return nil, fmt.Errorf("path must have 5 components, got %d", len(path))
	}
	out := make([]byte, 1, 1+4*len(path))
	out[0] = byte(len(path))
	for i, c := range path {
		if c >= 0x8000_0000 {
			return nil, fmt.Errorf("path component %d is out of range", i)
		}
		if i < 3 {
			c |= 0x8000_0000
		}
		out = binary.BigEndian.AppendUint32(out, c)
	}
	return out, nil
}

// getPublicKey returns the compressed public key and the address for path.
func (a *GnoApp) getPublicKey(path []uint32, display bool) ([]byte, string, error) {
	data, err := gnoPath(path)
	if err != nil {
		return nil, "", err
	}
	p1 := byte(gnoP1PubKeySilent)
	if display {
		p1 = gnoP1PubKeyDisplay
	}
	resp, err := a.exchange(gnoInsGetPublicKey, p1, 0, data)
	if err != nil {
		return nil, "", err
	}
	// pubkey length | pubkey | address length | address
	if len(resp) != 2+gnoPubKeyLen+gnoAddressLen ||
		resp[0] != gnoPubKeyLen || resp[1+gnoPubKeyLen] != gnoAddressLen {
		return nil, "", fmt.Errorf("malformed GET_PUBLIC_KEY response of %d bytes", len(resp))
	}
	return resp[1 : 1+gnoPubKeyLen], string(resp[2+gnoPubKeyLen:]), nil
}

// GetPublicKeySECP256K1 returns the compressed public key for path.
func (a *GnoApp) GetPublicKeySECP256K1(path []uint32) ([]byte, error) {
	pub, _, err := a.getPublicKey(path, false)
	return pub, err
}

// GetAddressPubKeySECP256K1 shows the address for path on the device and
// returns it with the compressed public key once the user approves it. The
// app shows gno addresses only, so hrp must be their prefix.
func (a *GnoApp) GetAddressPubKeySECP256K1(path []uint32, hrp string) ([]byte, string, error) {
	pub, addr, err := a.getPublicKey(path, true)
	if err != nil {
		return nil, "", err
	}
	if !strings.HasPrefix(addr, hrp+"1") {
		return nil, "", fmt.Errorf("the Gno Ledger app shows %s, not an address with prefix %q", addr, hrp)
	}
	return pub, addr, nil
}

// SignSECP256K1 has the user review msg, a transaction's signature payload,
// and returns its DER-encoded signature. The Gno app has a single signing
// mode, so p2 must be 0.
func (a *GnoApp) SignSECP256K1(path []uint32, msg []byte, p2 byte) ([]byte, error) {
	if p2 != 0 {
		return nil, fmt.Errorf("the Gno Ledger app has no signing mode %d", p2)
	}
	if len(msg) == 0 {
		return nil, errors.New("nothing to sign")
	}
	data, err := gnoPath(path)
	if err != nil {
		return nil, err
	}
	if _, err := a.exchange(gnoInsSignTx, gnoP1SignFirst, gnoP2SignMore, data); err != nil {
		return nil, err
	}
	var resp []byte
	for len(msg) > 0 {
		n := min(len(msg), gnoMaxChunk)
		p2 := byte(gnoP2SignMore)
		if n == len(msg) {
			p2 = gnoP2SignLast
		}
		if resp, err = a.exchange(gnoInsSignTx, gnoP1SignNext, p2, msg[:n]); err != nil {
			return nil, err
		}
		msg = msg[n:]
	}
	if len(resp) != gnoSignatureLen {
		return nil, fmt.Errorf("malformed signature of %d bytes", len(resp))
	}
	// The app returns R || S with S already low; callers expect DER.
	var r, s secp.ModNScalar
	if r.SetByteSlice(resp[:32]) || s.SetByteSlice(resp[32:]) || r.IsZero() || s.IsZero() {
		return nil, errors.New("signature values out of range")
	}
	return ecdsa.NewSignature(&r, &s).Serialize(), nil
}

// Close closes the connection to the device.
func (a *GnoApp) Close() error {
	return a.device.Close()
}
