package fork

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteTxsJSONL_RoundTrip verifies amino round-trip preserves
// std.Msg interface types in JSONL output.
func TestWriteTxsJSONL_RoundTrip(t *testing.T) {
	t.Parallel()

	// Create a tx with a concrete Msg (bank.MsgSend).
	msg := bank.MsgSend{
		FromAddress: crypto.AddressFromPreimage([]byte("sender")),
		ToAddress:   crypto.AddressFromPreimage([]byte("receiver")),
		Amount:      std.NewCoins(std.NewCoin("ugnot", 1000)),
	}
	tx := std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(50000, std.NewCoin("ugnot", 1000)),
	}
	original := []gnoland.TxWithMetadata{
		{
			Tx: tx,
			Metadata: &gnoland.GnoTxMetadata{
				Timestamp:   1234567890,
				BlockHeight: 42,
				ChainID:     "test-chain",
			},
		},
	}

	// Write to JSONL.
	dir := t.TempDir()
	path := filepath.Join(dir, "txs.jsonl")
	require.NoError(t, writeTxsJSONL(path, original))

	// Read back line-by-line using amino.UnmarshalJSON (the correct decoder
	// for amino-registered interfaces like std.Msg).
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	var decoded []gnoland.TxWithMetadata
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var tx gnoland.TxWithMetadata
		require.NoError(t, amino.UnmarshalJSON(line, &tx), "amino should unmarshal JSONL line")
		decoded = append(decoded, tx)
	}
	require.NoError(t, scanner.Err())

	require.Len(t, decoded, 1, "should decode exactly one tx")

	// The Msg should round-trip correctly with its type preserved.
	require.Len(t, decoded[0].Tx.Msgs, 1, "should have one msg")
	_, ok := decoded[0].Tx.Msgs[0].(bank.MsgSend)
	require.True(t, ok, "Msg should be bank.MsgSend after round-trip, got %T", decoded[0].Tx.Msgs[0])

	// Metadata should survive.
	require.NotNil(t, decoded[0].Metadata)
	assert.Equal(t, int64(42), decoded[0].Metadata.BlockHeight)
	assert.Equal(t, "test-chain", decoded[0].Metadata.ChainID)
}

// TestVerifyGenesisFile_Invalid verifies that verifyGenesisFile returns an
// error for a malformed genesis file (so the calling tool can abort).
func TestVerifyGenesisFile_Invalid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		err := verifyGenesisFile(filepath.Join(dir, "does-not-exist.json"))
		require.Error(t, err)
	})

	t.Run("malformed json", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "bad.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"not_valid": `), 0o644))
		err := verifyGenesisFile(path)
		require.Error(t, err)
	})
}
