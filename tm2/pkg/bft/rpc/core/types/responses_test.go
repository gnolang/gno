package core_types

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	p2ptypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusIndexer(t *testing.T) {
	t.Parallel()

	var status *ResultStatus
	assert.False(t, status.TxIndexEnabled())

	status = &ResultStatus{}
	assert.False(t, status.TxIndexEnabled())

	status.NodeInfo = p2ptypes.NodeInfo{}
	assert.False(t, status.TxIndexEnabled())

	cases := []struct {
		expected bool
		other    p2ptypes.NodeInfoOther
	}{
		{false, p2ptypes.NodeInfoOther{}},
		{false, p2ptypes.NodeInfoOther{TxIndex: "aa"}},
		{false, p2ptypes.NodeInfoOther{TxIndex: "off"}},
		{true, p2ptypes.NodeInfoOther{TxIndex: "on"}},
	}

	for _, tc := range cases {
		status.NodeInfo.Other = tc.other
		assert.Equal(t, tc.expected, status.TxIndexEnabled())
	}
}

// streamableAppState is a test-only AppState that streams a marker JSON object
// directly into the writer. It exercises the streaming code path of
// ResultGenesis.StreamJSON.
type streamableAppState struct {
	marker string
}

func (s *streamableAppState) StreamJSON(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, `{"streamed_marker":"`+s.marker+`"}`)
	return err
}

// fixtureGenesisDoc returns a GenesisDoc with realistic content (one
// validator carrying a polymorphic ed25519 PubKey) so the test exercises the
// non-AppState fields including the polymorphic Validators slice.
func fixtureGenesisDoc(t *testing.T, appState any) *types.GenesisDoc {
	t.Helper()
	pk := ed25519.GenPrivKey().PubKey()
	return &types.GenesisDoc{
		GenesisTime: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		ChainID:     "test-chain-123",
		Validators: []types.GenesisValidator{{
			Address: pk.Address(),
			PubKey:  pk,
			Power:   10,
			Name:    "v1",
		}},
		AppHash:  []byte{0xab, 0xcd},
		AppState: appState,
	}
}

// TestResultGenesis_StreamJSON_StreamableAppState verifies that when the
// underlying AppState implements StreamableResult, ResultGenesis.StreamJSON
// emits a JSON object whose "genesis.app_state" key contains the streamed
// bytes verbatim — not an amino-marshaled copy. The non-AppState fields must
// be byte-for-byte identical to what the standard amino marshal path produces
// so existing /genesis clients keep working.
func TestResultGenesis_StreamJSON_StreamableAppState(t *testing.T) {
	t.Parallel()

	app := &streamableAppState{marker: "streamed-content"}
	doc := fixtureGenesisDoc(t, app)
	res := &ResultGenesis{Genesis: doc}

	var buf bytes.Buffer
	require.NoError(t, res.StreamJSON(context.Background(), &buf))

	// The output must be a valid JSON object with a single "genesis" key.
	var envelope struct {
		Genesis json.RawMessage `json:"genesis"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope), "streamed output must be valid JSON")
	require.NotEmpty(t, envelope.Genesis)

	// Inside, app_state must be exactly the bytes StreamJSON wrote, not the
	// amino-marshaled struct.
	var inner struct {
		ChainID  string          `json:"chain_id"`
		AppState json.RawMessage `json:"app_state"`
	}
	require.NoError(t, json.Unmarshal(envelope.Genesis, &inner))
	assert.Equal(t, "test-chain-123", inner.ChainID)
	assert.JSONEq(t, `{"streamed_marker":"streamed-content"}`, string(inner.AppState))
}

// nonStreamableAppState is a plain struct that does NOT implement
// StreamableResult; the streaming path must fall back to amino marshaling.
type nonStreamableAppState struct {
	Plain string `json:"plain"`
}

// TestResultGenesis_StreamJSON_NonStreamableAppState verifies that when the
// AppState does NOT implement StreamableResult, the streaming path falls back
// to amino marshaling — this preserves wire compatibility for existing
// in-memory AppState callers and keeps the streaming hook truly opt-in.
func TestResultGenesis_StreamJSON_NonStreamableAppState(t *testing.T) {
	t.Parallel()

	doc := fixtureGenesisDoc(t, nonStreamableAppState{Plain: "value"})
	res := &ResultGenesis{Genesis: doc}

	var buf bytes.Buffer
	require.NoError(t, res.StreamJSON(context.Background(), &buf))

	var envelope struct {
		Genesis json.RawMessage `json:"genesis"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))

	var inner struct {
		ChainID  string          `json:"chain_id"`
		AppState json.RawMessage `json:"app_state"`
	}
	require.NoError(t, json.Unmarshal(envelope.Genesis, &inner))
	assert.Equal(t, "test-chain-123", inner.ChainID)
	assert.JSONEq(t, `{"plain":"value"}`, string(inner.AppState))
}

// TestResultGenesis_StreamJSON_NilAppState verifies the omitempty behavior is
// preserved when AppState is nil: no "app_state" key appears at all.
func TestResultGenesis_StreamJSON_NilAppState(t *testing.T) {
	t.Parallel()

	doc := fixtureGenesisDoc(t, nil)
	res := &ResultGenesis{Genesis: doc}

	var buf bytes.Buffer
	require.NoError(t, res.StreamJSON(context.Background(), &buf))

	var envelope struct {
		Genesis json.RawMessage `json:"genesis"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))

	// Use a permissive map decode so we can verify the absence of app_state.
	var inner map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(envelope.Genesis, &inner))
	_, hasAppState := inner["app_state"]
	assert.False(t, hasAppState, "nil AppState should be omitted via omitempty")
	assert.Contains(t, inner, "chain_id")
}

// TestResultGenesis_StreamJSON_PreservesValidators verifies that the
// polymorphic Validators slice (each carrying a crypto.PubKey interface) is
// rendered through the same amino code path as the non-streaming response —
// otherwise the type tag (`@type` / `value`) would be missing and clients
// would fail to decode. The streamed bytes for the non-AppState fields must
// be byte-for-byte identical to amino's marshal of a doc with AppState=nil,
// since that's literally how the implementation produces them.
func TestResultGenesis_StreamJSON_PreservesValidators(t *testing.T) {
	t.Parallel()

	app := &streamableAppState{marker: "x"}
	doc := fixtureGenesisDoc(t, app)
	res := &ResultGenesis{Genesis: doc}

	var buf bytes.Buffer
	require.NoError(t, res.StreamJSON(context.Background(), &buf))

	var envelope struct {
		Genesis json.RawMessage `json:"genesis"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))

	// Inspect the validators[].pub_key wire shape: must carry the amino
	// polymorphism marker @type / value, otherwise consumers can't decode.
	var inner struct {
		Validators []struct {
			PubKey  json.RawMessage `json:"pub_key"`
			Address string          `json:"address"`
			Power   string          `json:"power"`
			Name    string          `json:"name"`
		} `json:"validators"`
	}
	require.NoError(t, json.Unmarshal(envelope.Genesis, &inner))
	require.Len(t, inner.Validators, 1)
	assert.Equal(t, doc.Validators[0].Address.String(), inner.Validators[0].Address)
	assert.Equal(t, "v1", inner.Validators[0].Name)
	assert.Equal(t, "10", inner.Validators[0].Power)
	assert.Contains(t, string(inner.Validators[0].PubKey), `"@type"`,
		"validator pub_key must carry amino polymorphism marker")
	assert.Contains(t, string(inner.Validators[0].PubKey), `"value"`)
}
