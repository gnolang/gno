package bech32_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
)

func TestEncodeAndDecode(t *testing.T) {
	t.Parallel()

	sum := sha256.Sum256([]byte("hello world\n"))

	bech, err := bech32.ConvertAndEncode("shasum", sum[:])
	if err != nil {
		t.Error(err)
	}
	hrp, data, err := bech32.DecodeAndConvert(bech)
	if err != nil {
		t.Error(err)
	}
	if hrp != "shasum" {
		t.Error("Invalid hrp")
	}
	if !bytes.Equal(data, sum[:]) {
		t.Error("Invalid decode")
	}
}

var pubkeyBech32 = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqg5y7u93gpzug38k2p8s8322zpdm96t0ch87ax88sre4vnclz2jcy8uyhst"

// amino marshaled pubkey bytes. pubkey.Bytes()
var pubkeyBytes = "0A132F746D2E5075624B6579536563703235366B3112230A2102284F70B14045C444F6504F03C54A105BB2E96FC5CFEE98E780F3564F1F12A582"

func TestEncode(t *testing.T) {
	t.Parallel()

	bz, err := hex.DecodeString(pubkeyBytes)

	assert.NoError(t, err)

	p, err := bech32.Encode(crypto.Bech32PubKeyPrefix(), bz)

	assert.NoError(t, err)
	assert.Equal(t, pubkeyBech32, p)
}

func TestDecode(t *testing.T) {
	t.Parallel()

	hrp, b1, err := bech32.Decode(pubkeyBech32)

	assert.NoError(t, err)
	assert.Equal(t, crypto.Bech32PubKeyPrefix(), hrp)

	b2, err := hex.DecodeString(pubkeyBytes)

	assert.NoError(t, err)
	assert.Equal(t, b1, b2)
}
