package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
)

var testProposal *Proposal

func init() {
	stamp, err := time.Parse(TimeFormat, "2018-02-11T07:09:22.765Z")
	if err != nil {
		panic(err)
	}
	testProposal = &Proposal{
		Height:    12345,
		Round:     23456,
		BlockID:   BlockID{[]byte{1, 2, 3}, PartSetHeader{111, []byte("blockparts")}},
		POLRound:  -1,
		Timestamp: stamp,
	}
}

func TestProposalSignable(t *testing.T) {
	t.Parallel()

	chainID := "test_chain_id"
	signBytes := testProposal.SignBytes(chainID)

	expected, err := amino.MarshalSized(CanonicalizeProposal(chainID, testProposal))
	require.NoError(t, err)
	require.Equal(t, expected, signBytes, "Got unexpected sign bytes for Proposal")
}

// TestProposalSignBytesTestVectors locks down the exact on-wire sign-bytes
// for a known CanonicalProposal. A wire-format regression in the fixed64
// Height/Round/POLRound fields or the field ordering would invalidate every
// historical proposal signature — catching that here is cheaper than
// discovering it via cross-validator signature failure. Mirrors the coverage
// TestVoteSignBytesTestVectors provides for Vote.
func TestProposalSignBytesTestVectors(t *testing.T) {
	t.Parallel()

	// Height=1, Round=1, POLRound=-1 (a real "no prior precommit" proposal).
	// Zero Timestamp (Go zero time.Time) — the default sign-bytes pattern.
	prop := &Proposal{Type: ProposalType, Height: 1, Round: 1, POLRound: -1}

	want := []byte{
		0x2a,       // length = 42
		0x08, 0x20, // Type = ProposalType (0x20)
		0x11, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Height = 1 (fixed64 LE)
		0x19, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Round = 1 (fixed64 LE)
		0x21, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // POLRound = -1 (fixed64 LE)
		0x32, 0x0b, // Timestamp field + length
		0x08, 0x80, 0x92, 0xb8, 0xc3, 0x98, 0xfe, 0xff, 0xff, 0xff, 0x01,
	}
	require.Equal(t, want, prop.SignBytes(""))
}

func TestProposalString(t *testing.T) {
	t.Parallel()

	str := testProposal.String()
	expected := `Proposal{12345/23456 (010203:111:626C6F636B70, -1) 000000000000 @ 2018-02-11T07:09:22.765Z}`
	if str != expected {
		t.Errorf("Got unexpected string for Proposal. Expected:\n%v\nGot:\n%v", expected, str)
	}
}

func TestProposalVerifySignature(t *testing.T) {
	t.Parallel()

	privVal := NewMockPV()
	pubKey := privVal.PubKey()

	prop := NewProposal(
		4, 2, 2,
		BlockID{[]byte{1, 2, 3}, PartSetHeader{777, []byte("proper")}})
	signBytes := prop.SignBytes("test_chain_id")

	// sign it
	err := privVal.SignProposal("test_chain_id", prop)
	require.NoError(t, err)

	// verify the same proposal
	valid := pubKey.VerifyBytes(signBytes, prop.Signature)
	require.True(t, valid)

	// serialize, deserialize and verify again....
	newProp := new(Proposal)
	bs, err := amino.MarshalSized(prop)
	require.NoError(t, err)
	err = amino.UnmarshalSized(bs, &newProp)
	require.NoError(t, err)

	// verify the transmitted proposal
	newSignBytes := newProp.SignBytes("test_chain_id")
	require.Equal(t, string(signBytes), string(newSignBytes))
	valid = pubKey.VerifyBytes(newSignBytes, newProp.Signature)
	require.True(t, valid)
}

func BenchmarkProposalWriteSignBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testProposal.SignBytes("test_chain_id")
	}
}

func BenchmarkProposalSign(b *testing.B) {
	privVal := NewMockPV()
	for i := 0; i < b.N; i++ {
		err := privVal.SignProposal("test_chain_id", testProposal)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkProposalVerifySignature(b *testing.B) {
	privVal := NewMockPV()
	err := privVal.SignProposal("test_chain_id", testProposal)
	require.Nil(b, err)

	for i := 0; i < b.N; i++ {
		privVal.PubKey().VerifyBytes(testProposal.SignBytes("test_chain_id"), testProposal.Signature)
	}
}

func TestProposalValidateBasic(t *testing.T) {
	t.Parallel()

	privVal := NewMockPV()
	testCases := []struct {
		testName         string
		malleateProposal func(*Proposal)
		expectErr        bool
	}{
		{"Good Proposal", func(p *Proposal) {}, false},
		{"Invalid Type", func(p *Proposal) { p.Type = PrecommitType }, true},
		{"Invalid Height", func(p *Proposal) { p.Height = -1 }, true},
		{"Invalid Round", func(p *Proposal) { p.Round = -1 }, true},
		{"Invalid POLRound", func(p *Proposal) { p.POLRound = -2 }, true},
		{"Invalid BlockId", func(p *Proposal) {
			p.BlockID = BlockID{[]byte{1, 2, 3}, PartSetHeader{111, []byte("blockparts")}}
		}, true},
		{"Invalid Signature", func(p *Proposal) {
			p.Signature = make([]byte, 0)
		}, true},
		{"Too big Signature", func(p *Proposal) {
			p.Signature = make([]byte, MaxSignatureSize+1)
		}, true},
	}
	blockID := makeBlockID(tmhash.Sum([]byte("blockhash")), 10, tmhash.Sum([]byte("partshash")))

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			prop := NewProposal(
				4, 2, 2,
				blockID)
			err := privVal.SignProposal("test_chain_id", prop)
			require.NoError(t, err)
			tc.malleateProposal(prop)
			assert.Equal(t, tc.expectErr, prop.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}
