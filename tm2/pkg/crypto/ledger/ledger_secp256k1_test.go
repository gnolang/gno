package ledger

import (
	"errors"
	"strconv"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/internal/ledger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// refusingDevice records whether anything was asked of the device.
type refusingDevice struct {
	ledger.MockLedger
	asked bool
}

func (d *refusingDevice) GetPublicKeySECP256K1(path []uint32) ([]byte, error) {
	d.asked = true

	return nil, nil
}

func (d *refusingDevice) SignSECP256K1(path []uint32, msg []byte, mode byte) ([]byte, error) {
	d.asked = true

	return nil, nil
}

// A CALLER HAS TO BE ABLE TO TELL THESE APART. All three failures mean the
// payload is too large for the device to parse and display, and all three
// arrive as opaque text saying nothing about size: a front end can only offer a
// remedy if this package classifies them. The remedy itself is not here, since
// gnokey has a flag to suggest and gnoclient does not.
func TestExplainSignError(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		err  error
		size int
		want error
	}{
		{
			name: "the app refuses the payload outright",
			err:  errors.New("APDU Error Code from Ledger Device: 0x6988"),
			size: 12000,
			want: ErrPayloadTooLarge,
		},
		{
			name: "the app runs out of JSON tokens",
			err:  errors.New("not enough tokens were provided"),
			size: 4000,
			want: ErrTooManyJSONValues,
		},
		{
			name: "the transport gives up on a large payload",
			err:  errors.New("hidapi: unknown failure"),
			size: 8840,
			want: ErrLargePayloadRefused,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := explainSignError(tc.err, tc.size)
			require.ErrorIs(t, got, tc.want)

			// The device's own words are kept: the classification is a guess
			// about the cause, and a signer debugging a cable needs the
			// original.
			assert.ErrorContains(t, got, tc.err.Error())
			require.ErrorIs(t, got, tc.err)
		})
	}
}

// NO ADVICE LEAKS DOWN HERE. This package cannot import the gnokey client that
// defines -sign-mode, and gnoclient reaches the same code with no flags at all,
// so naming one would be advice half the callers cannot act on.
func TestLedgerErrorsNameNoClientFlag(t *testing.T) {
	t.Parallel()

	messages := []string{
		explainSignError(errors.New("APDU Error Code from Ledger Device: 0x6988"), 12000).Error(),
		explainSignError(errors.New("not enough tokens were provided"), 4000).Error(),
		explainSignError(errors.New("hidapi: unknown failure"), 8840).Error(),
		ErrPayloadTooLarge.Error(),
		ErrTooManyJSONValues.Error(),
		ErrLargePayloadRefused.Error(),
	}
	for _, msg := range messages {
		assert.NotContains(t, msg, "sign-mode", "the device layer is naming a client flag")
	}
}

// The 0x6988 arm quotes no ceiling. The pre-check in sign already refused
// anything over MaxPayloadSize, so a payload the DEVICE refuses for capacity is
// under the number this package would otherwise cite, and an error quoting both
// would argue with itself.
func TestDeviceRefusalQuotesNoCeiling(t *testing.T) {
	t.Parallel()

	got := explainSignError(errors.New("APDU Error Code from Ledger Device: 0x6988"), 12000)

	assert.ErrorContains(t, got, "12000")
	assert.NotContains(t, got.Error(), strconv.Itoa(MaxPayloadSize))
}

// AND AN ORDINARY FAILURE MUST NOT COLLECT SIZE ADVICE. A rejected prompt or an
// unplugged device on a small payload has nothing to do with the sign mode, and
// telling someone to change it would send them after the wrong thing.
func TestExplainSignErrorLeavesOrdinaryFailuresAlone(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		err  error
		size int
	}{
		{"user rejected", errors.New("[APDU_CODE_COMMAND_NOT_ALLOWED] Command not allowed / User Rejected (no current EF)"), 420},
		{"app not open", errors.New("[APDU_CODE_APP_NOT_OPEN] Ledger Connected but Chain Specific App Not Open"), 420},
		{"transport failure on a small payload", errors.New("hidapi: unknown failure"), largePayloadSize},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := explainSignError(tc.err, tc.size)
			assert.Equal(t, tc.err, got)
			assert.NotContains(t, got.Error(), "sign-mode")
		})
	}
}

// The pre-check refuses an oversized payload before the device is asked, and
// names both sizes: a signer whose transaction is twice the ceiling is deciding
// whether to split it, and needs the numbers.
func TestSignRefusesAnOversizedPayload(t *testing.T) {
	t.Parallel()

	device := &refusingDevice{}

	_, err := sign(device, PrivKeyLedgerSecp256k1{}, make([]byte, MaxPayloadSize+1))

	require.ErrorIs(t, err, ErrPayloadTooLarge)
	assert.ErrorContains(t, err, strconv.Itoa(MaxPayloadSize+1))
	assert.ErrorContains(t, err, strconv.Itoa(MaxPayloadSize))
	assert.False(t, device.asked, "the device was asked about a payload it cannot hold")
}
