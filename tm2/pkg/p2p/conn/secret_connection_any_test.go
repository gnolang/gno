package conn

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/async"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func makeSecretConnPairAny(tb testing.TB, fooPriv, barPriv crypto.PrivKey) (
	fooSecConn, barSecConn *SecretConnection,
	fooRemote, barRemote crypto.PubKey,
) {
	tb.Helper()

	fooConn, barConn := makeKVStoreConnPair()
	fooPub := fooPriv.PubKey()
	barPub := barPriv.PubKey()

	trs, ok := async.Parallel(
		func(_ int) (val any, err error, abort bool) {
			fsc, rem, err := MakeSecretConnectionAny(fooConn, fooPriv)
			if err != nil {
				tb.Errorf("foo handshake: %v", err)
				return nil, err, true
			}
			if !rem.Equals(barPub) {
				err = fmt.Errorf("foo saw remote pubkey %v, expected %v", rem, barPub)
				tb.Error(err)
				return nil, err, false
			}
			fooSecConn = fsc
			fooRemote = rem
			return nil, nil, false
		},
		func(_ int) (val any, err error, abort bool) {
			bsc, rem, err := MakeSecretConnectionAny(barConn, barPriv)
			if err != nil {
				tb.Errorf("bar handshake: %v", err)
				return nil, err, true
			}
			if !rem.Equals(fooPub) {
				err = fmt.Errorf("bar saw remote pubkey %v, expected %v", rem, fooPub)
				tb.Error(err)
				return nil, err, false
			}
			barSecConn = bsc
			barRemote = rem
			return nil, nil, false
		},
	)

	require.Nil(tb, trs.FirstError())
	require.True(tb, ok, "unexpected task abortion")
	return
}

func TestSecretConnectionAny_Handshake(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		foo  crypto.PrivKey
		bar  crypto.PrivKey
	}{
		{
			name: "ed25519 both sides",
			foo:  ed25519.GenPrivKey(),
			bar:  ed25519.GenPrivKey(),
		},
		{
			name: "secp256k1 both sides",
			foo:  secp256k1.GenPrivKey(),
			bar:  secp256k1.GenPrivKey(),
		},
		{
			name: "ed25519 + secp256k1 mixed",
			foo:  ed25519.GenPrivKey(),
			bar:  secp256k1.GenPrivKey(),
		},
		{
			name: "secp256k1 + ed25519 mixed (swapped)",
			foo:  secp256k1.GenPrivKey(),
			bar:  ed25519.GenPrivKey(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fooSC, barSC, _, _ := makeSecretConnPairAny(t, tt.foo, tt.bar)
			require.NotNil(t, fooSC)
			require.NotNil(t, barSC)
			require.NoError(t, fooSC.Close())
			require.NoError(t, barSC.Close())
		})
	}
}

func TestSecretConnectionAny_NilPrivKey(t *testing.T) {
	t.Parallel()

	fooConn, _ := makeKVStoreConnPair()
	sc, pub, err := MakeSecretConnectionAny(fooConn, nil)
	require.Error(t, err)
	require.Nil(t, sc)
	require.Nil(t, pub)
}

func TestSecretConnectionAny_ReadWrite(t *testing.T) {
	t.Parallel()

	fooSC, barSC, _, _ := makeSecretConnPairAny(t,
		secp256k1.GenPrivKey(),
		ed25519.GenPrivKey(),
	)
	defer fooSC.Close()
	defer barSC.Close()

	msg := []byte("hello from a mixed-scheme channel")
	done := make(chan error, 1)
	go func() {
		buf := make([]byte, len(msg))
		_, err := barSC.Read(buf)
		if err != nil {
			done <- err
			return
		}
		if string(buf) != string(msg) {
			done <- fmt.Errorf("got %q want %q", buf, msg)
			return
		}
		done <- nil
	}()

	_, err := fooSC.Write(msg)
	require.NoError(t, err)
	require.NoError(t, <-done)
}
