package std_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LookalikeMsg is a Msg shaped like a sign-data message: in a transaction it
// differs from one by its "@type" alone.
type LookalikeMsg struct {
	MsgType string        `json:"type"`
	Value   LookalikeData `json:"value"`
}

type LookalikeData struct {
	Data    string `json:"data"`
	Purpose string `json:"purpose"`
	Signer  string `json:"signer"`
	URL     string `json:"url"`
}

func (LookalikeMsg) Route() string                { return "sign" }
func (LookalikeMsg) Type() string                 { return "MsgSignData" }
func (LookalikeMsg) ValidateBasic() error         { return nil }
func (LookalikeMsg) GetSignBytes() []byte         { return nil }
func (LookalikeMsg) GetSigners() []crypto.Address { return nil }

var _ = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/std_test",
	"std_test",
	amino.GetCallersDirname(),
).WithTypes(
	LookalikeMsg{}, "MsgSignData",
))

// A sign-data proof must never be a transaction signature. With a zero fee, a
// transaction's sign bytes match a proof's in every field but the messages, and
// the only thing setting those apart is the "@type" amino writes for a Msg in
// an interface slot, which a sign-data message lacks. If either side changes
// (signDocPayload.Msgs made concrete, the sign-data message registered as a
// Msg), every proof becomes a signed transaction and this fails.
func TestSignDataPayloadIsNoTransactionPayload(t *testing.T) {
	t.Parallel()

	d := std.SignedData{
		ChainID: "dev",
		URL:     "https://example.com",
		Purpose: "test",
		Signer:  crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		Data:    "hello",
	}
	proof, err := std.SignDataPayload(d)
	require.NoError(t, err)

	for _, msg := range msgKeys(t, proof) {
		assert.NotContains(t, msg, "@type", "a sign-data message reads as a transaction's")
	}

	lookalike := LookalikeMsg{
		MsgType: std.SignDataMsgType,
		Value: LookalikeData{
			Data:    base64.StdEncoding.EncodeToString([]byte(d.Data)),
			Purpose: d.Purpose,
			Signer:  d.Signer.String(),
			URL:     d.URL,
		},
	}
	msgSets := map[string][]std.Msg{
		"no message":      nil,
		"a lookalike":     {lookalike},
		"two lookalikes":  {lookalike, lookalike},
		"an empty struct": {LookalikeMsg{}},
	}

	for name, msgs := range msgSets {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// A node accepts a signature over either rendering.
			doc := std.SignDoc{ChainID: d.ChainID, Msgs: msgs}
			current, err := std.GetSignaturePayload(doc)
			require.NoError(t, err)
			legacy, err := std.GetSignaturePayloadLegacy(doc)
			require.NoError(t, err)

			for _, tx := range [][]byte{current, legacy} {
				assert.False(t, bytes.Equal(tx, proof), "a transaction's sign bytes equal a proof's:\n%s", tx)
				for _, msg := range msgKeys(t, tx) {
					assert.Contains(t, msg, "@type", "a transaction's message reads as a sign-data one:\n%s", tx)
				}
			}
		})
	}
}

// msgKeys returns the keys of each message in a sign document.
func msgKeys(t *testing.T, payload []byte) []map[string]json.RawMessage {
	t.Helper()

	var doc struct {
		Msgs []map[string]json.RawMessage `json:"msgs"`
	}
	require.NoError(t, json.Unmarshal(payload, &doc))

	return doc.Msgs
}
