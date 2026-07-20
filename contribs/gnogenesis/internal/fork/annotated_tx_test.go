package fork

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleTx(memo string) std.Tx {
	return std.Tx{
		Msgs: []std.Msg{
			bank.MsgSend{
				FromAddress: crypto.Address{0x01},
				ToAddress:   crypto.Address{0x02},
				Amount:      std.NewCoins(std.NewCoin(ugnot.Denom, 100)),
			},
		},
		Fee:  std.Fee{GasWanted: 10, GasFee: std.NewCoin(ugnot.Denom, 1000000)},
		Memo: memo,
	}
}

func TestAnnotatedTx_AminoJSONRoundtrip(t *testing.T) {
	t.Parallel()

	at := AnnotatedTx{
		Tx: sampleTx("patched-body"),
		Metadata: &gnoland.GnoTxMetadata{
			BlockHeight: 1950,
			ChainID:     "gnoland-1",
		},
		Reason: "API drift on params.ProposeAddUnrestrictedAcctsRequest (post-#5669)",
	}

	bz, err := amino.MarshalJSON(at)
	require.NoError(t, err)
	assert.Contains(t, string(bz), `"reason"`)
	assert.Contains(t, string(bz), `"API drift on params`)

	var got AnnotatedTx
	require.NoError(t, amino.UnmarshalJSON(bz, &got))
	assert.Equal(t, at.Reason, got.Reason)
	assert.Equal(t, at.Tx.Memo, got.Tx.Memo)
	require.NotNil(t, got.Metadata)
	assert.Equal(t, int64(1950), got.Metadata.BlockHeight)
}

func TestAnnotatedTx_ReasonOmitemptyWhenUnset(t *testing.T) {
	t.Parallel()

	at := AnnotatedTx{Tx: sampleTx("no-reason")}
	bz, err := amino.MarshalJSON(at)
	require.NoError(t, err)
	assert.NotContains(t, string(bz), `"reason"`)
}

func TestReadAnnotatedTxs(t *testing.T) {
	t.Parallel()

	t.Run("multi-line jsonl with reasons", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "patches.jsonl")

		entries := []AnnotatedTx{
			{Tx: sampleTx("a"), Reason: "first reason"},
			{Tx: sampleTx("b"), Reason: "second reason"},
		}
		var buf strings.Builder
		for _, e := range entries {
			data, err := amino.MarshalJSON(e)
			require.NoError(t, err)
			buf.Write(data)
			buf.WriteByte('\n')
		}
		require.NoError(t, os.WriteFile(path, []byte(buf.String()), 0o644))

		got, err := readAnnotatedTxs(path)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "first reason", got[0].Reason)
		assert.Equal(t, "second reason", got[1].Reason)
		assert.Equal(t, "a", got[0].Tx.Memo)
		assert.Equal(t, "b", got[1].Tx.Memo)
	})

	t.Run("blank lines and comments ignored", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "patches.jsonl")

		at, err := amino.MarshalJSON(AnnotatedTx{Tx: sampleTx("only"), Reason: "r"})
		require.NoError(t, err)
		content := "\n# leading comment\n" + string(at) + "\n\n# trailing comment\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

		got, err := readAnnotatedTxs(path)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "r", got[0].Reason)
	})

	t.Run("plain TxWithMetadata line parses with empty reason", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "legacy.jsonl")

		legacy := gnoland.TxWithMetadata{
			Tx:       sampleTx("legacy"),
			Metadata: &gnoland.GnoTxMetadata{BlockHeight: 42},
		}
		bz, err := amino.MarshalJSON(legacy)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(path, append(bz, '\n'), 0o644))

		got, err := readAnnotatedTxs(path)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "", got[0].Reason)
		assert.Equal(t, "legacy", got[0].Tx.Memo)
		require.NotNil(t, got[0].Metadata)
		assert.Equal(t, int64(42), got[0].Metadata.BlockHeight)
	})

	t.Run("invalid json line surfaces line number", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "broken.jsonl")

		at, err := amino.MarshalJSON(AnnotatedTx{Tx: sampleTx("ok"), Reason: "r"})
		require.NoError(t, err)
		content := string(at) + "\n!not valid json!\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

		_, err = readAnnotatedTxs(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "line 2")
	})
}
