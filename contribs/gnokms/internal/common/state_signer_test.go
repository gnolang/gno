package common

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSigner is a deterministic ed25519-backed signer used by tests. It also
// counts how many times Sign() was called so tests can assert that the
// HRSGuardedSigner short-circuits before reaching the inner signer.
type fakeSigner struct {
	priv      ed25519.PrivKeyEd25519
	signCount int
	failNext  bool
}

func newFakeSigner() *fakeSigner {
	return &fakeSigner{priv: ed25519.GenPrivKey()}
}

func (f *fakeSigner) PubKey() crypto.PubKey { return f.priv.PubKey() }
func (f *fakeSigner) Close() error          { return nil }
func (f *fakeSigner) Sign(b []byte) ([]byte, error) {
	f.signCount++
	if f.failNext {
		f.failNext = false
		return nil, errors.New("fake signer: induced failure")
	}
	return f.priv.Sign(b)
}

// makeVoteBytes builds amino-encoded SignBytes for a CanonicalVote at the
// given (height, round, voteType).
func makeVoteBytes(t *testing.T, height int64, round int64, voteType types.SignedMsgType, chainID string) []byte {
	t.Helper()
	v := types.CanonicalVote{
		Type:      voteType,
		Height:    height,
		Round:     round,
		BlockID:   types.CanonicalBlockID{Hash: []byte{0xab, 0xcd}},
		Timestamp: time.Unix(1700000000, 0).UTC(),
		ChainID:   chainID,
	}
	bz, err := amino.MarshalSized(v)
	require.NoError(t, err)
	return bz
}

// makeVoteBytesWithTime is makeVoteBytes with caller-controlled timestamp,
// used to test the same-HRS-different-bytes (timestamp drift) rejection path.
func makeVoteBytesWithTime(t *testing.T, height int64, round int64, voteType types.SignedMsgType, chainID string, ts time.Time) []byte {
	t.Helper()
	v := types.CanonicalVote{
		Type:      voteType,
		Height:    height,
		Round:     round,
		BlockID:   types.CanonicalBlockID{Hash: []byte{0xab, 0xcd}},
		Timestamp: ts.UTC(),
		ChainID:   chainID,
	}
	bz, err := amino.MarshalSized(v)
	require.NoError(t, err)
	return bz
}

func makeProposalBytes(t *testing.T, height int64, round int64, chainID string) []byte {
	t.Helper()
	p := types.CanonicalProposal{
		Type:      types.ProposalType,
		Height:    height,
		Round:     round,
		POLRound:  -1,
		BlockID:   types.CanonicalBlockID{Hash: []byte{0xde, 0xad}},
		Timestamp: time.Unix(1700000000, 0).UTC(),
		ChainID:   chainID,
	}
	bz, err := amino.MarshalSized(p)
	require.NoError(t, err)
	return bz
}

func newGuardedSigner(t *testing.T) (*HRSGuardedSigner, *fakeSigner, string) {
	t.Helper()
	dir := t.TempDir()
	statePath := filepath.Join(dir, "signer_state.json")
	inner := newFakeSigner()
	g, err := NewHRSGuardedSigner(inner, statePath, nil)
	require.NoError(t, err)
	return g, inner, statePath
}

// Monotonic HRS sequences are signed and persisted in order.
func TestHRSGuard_Monotonic(t *testing.T) {
	t.Parallel()
	g, inner, _ := newGuardedSigner(t)

	// Heights advance.
	for h := int64(1); h <= 3; h++ {
		bz := makeVoteBytes(t, h, 0, types.PrecommitType, "test")
		sig, err := g.Sign(bz)
		require.NoError(t, err)
		assert.NotEmpty(t, sig)
	}
	assert.Equal(t, 3, inner.signCount)

	// At a fixed height, step advances Prevote -> Precommit.
	g, inner, _ = newGuardedSigner(t)
	prevote := makeVoteBytes(t, 1, 0, types.PrevoteType, "test")
	precommit := makeVoteBytes(t, 1, 0, types.PrecommitType, "test")
	_, err := g.Sign(prevote)
	require.NoError(t, err)
	_, err = g.Sign(precommit)
	require.NoError(t, err)
	assert.Equal(t, 2, inner.signCount)

	// At a fixed height/step, round advances 0 -> 1.
	g, inner, _ = newGuardedSigner(t)
	r0 := makeVoteBytes(t, 1, 0, types.PrecommitType, "test")
	r1 := makeVoteBytes(t, 1, 1, types.PrecommitType, "test")
	_, err = g.Sign(r0)
	require.NoError(t, err)
	_, err = g.Sign(r1)
	require.NoError(t, err)
	assert.Equal(t, 2, inner.signCount)

	// A proposal at h+1 after a precommit at h is allowed.
	g, _, _ = newGuardedSigner(t)
	_, err = g.Sign(makeVoteBytes(t, 5, 0, types.PrecommitType, "test"))
	require.NoError(t, err)
	_, err = g.Sign(makeProposalBytes(t, 6, 0, "test"))
	require.NoError(t, err)
}

// Height regression is refused; no signature is produced; inner is not called.
func TestHRSGuard_HeightRegression(t *testing.T) {
	t.Parallel()
	g, inner, _ := newGuardedSigner(t)

	_, err := g.Sign(makeVoteBytes(t, 10, 0, types.PrecommitType, "test"))
	require.NoError(t, err)
	require.Equal(t, 1, inner.signCount)

	_, err = g.Sign(makeVoteBytes(t, 9, 0, types.PrecommitType, "test"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "height regression")
	assert.Equal(t, 1, inner.signCount, "inner signer must not be called on regression")
}

func TestHRSGuard_RoundRegression(t *testing.T) {
	t.Parallel()
	g, inner, _ := newGuardedSigner(t)
	_, err := g.Sign(makeVoteBytes(t, 1, 5, types.PrecommitType, "test"))
	require.NoError(t, err)

	_, err = g.Sign(makeVoteBytes(t, 1, 4, types.PrecommitType, "test"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "round regression")
	assert.Equal(t, 1, inner.signCount)
}

func TestHRSGuard_StepRegression(t *testing.T) {
	t.Parallel()
	g, inner, _ := newGuardedSigner(t)
	_, err := g.Sign(makeVoteBytes(t, 1, 0, types.PrecommitType, "test"))
	require.NoError(t, err)

	// Same height/round, lower step (Prevote=2 < Precommit=3).
	_, err = g.Sign(makeVoteBytes(t, 1, 0, types.PrevoteType, "test"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "step regression")
	assert.Equal(t, 1, inner.signCount)
}

// Same HRS with byte-identical SignBytes returns the cached signature
// without invoking the inner signer (idempotent retransmit).
func TestHRSGuard_SameHRSIdempotent(t *testing.T) {
	t.Parallel()
	g, inner, _ := newGuardedSigner(t)

	bz := makeVoteBytes(t, 7, 2, types.PrecommitType, "test")
	sig1, err := g.Sign(bz)
	require.NoError(t, err)
	require.Equal(t, 1, inner.signCount)

	sig2, err := g.Sign(bz)
	require.NoError(t, err)
	assert.Equal(t, sig1, sig2, "idempotent retransmit must return cached signature")
	assert.Equal(t, 1, inner.signCount, "inner must not be called for byte-identical replay")
}

// Same HRS with non-identical SignBytes (e.g., timestamp drift, or hostile
// content swap) is refused. This is the slashing-prevention guarantee.
func TestHRSGuard_SameHRSConflict(t *testing.T) {
	t.Parallel()
	g, inner, _ := newGuardedSigner(t)

	t1 := time.Unix(1700000000, 0)
	t2 := time.Unix(1700000005, 0) // 5s later

	bz1 := makeVoteBytesWithTime(t, 7, 2, types.PrecommitType, "test", t1)
	bz2 := makeVoteBytesWithTime(t, 7, 2, types.PrecommitType, "test", t2)

	_, err := g.Sign(bz1)
	require.NoError(t, err)
	require.Equal(t, 1, inner.signCount)

	_, err = g.Sign(bz2)
	require.ErrorIs(t, err, ErrSameHRSConflict)
	assert.Equal(t, 1, inner.signCount, "inner must not be called for same-HRS conflict")
}

// State persists across NewHRSGuardedSigner calls — i.e., a gnokms restart
// cannot be tricked into double-signing by simply re-creating the wrapper.
func TestHRSGuard_PersistAcrossRestart(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	statePath := filepath.Join(dir, "signer_state.json")

	// Initial run: sign at height 100.
	inner1 := newFakeSigner()
	g1, err := NewHRSGuardedSigner(inner1, statePath, nil)
	require.NoError(t, err)
	_, err = g1.Sign(makeVoteBytes(t, 100, 0, types.PrecommitType, "test"))
	require.NoError(t, err)
	require.NoError(t, g1.Close())

	// Restart: re-open the same state file with a fresh inner signer.
	// (In reality the inner signer would be the same key; for the test we
	// only care that the state file gates regression.)
	inner2 := newFakeSigner()
	g2, err := NewHRSGuardedSigner(inner2, statePath, nil)
	require.NoError(t, err)

	// Attempting to sign at H=99 must be refused — the state file
	// remembers that H=100 was already signed.
	_, err = g2.Sign(makeVoteBytes(t, 99, 0, types.PrecommitType, "test"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "height regression")
	assert.Equal(t, 0, inner2.signCount)

	// H=101 is allowed.
	_, err = g2.Sign(makeVoteBytes(t, 101, 0, types.PrecommitType, "test"))
	require.NoError(t, err)
	assert.Equal(t, 1, inner2.signCount)
}

// Garbage SignBytes are rejected without consulting the inner signer. This
// covers the case where a hostile (or buggy) client sends bytes that aren't
// a CanonicalVote or CanonicalProposal — gnokms must not blindly sign them.
func TestHRSGuard_RejectsGarbage(t *testing.T) {
	t.Parallel()
	g, inner, _ := newGuardedSigner(t)

	cases := [][]byte{
		nil,
		{},
		{0x00},
		{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		[]byte("hello world"),
	}
	for _, bz := range cases {
		_, err := g.Sign(bz)
		require.Error(t, err, "garbage bytes %x must be refused", bz)
	}
	assert.Equal(t, 0, inner.signCount, "inner signer must not be called for any garbage input")
}

// classifySignBytes correctly discriminates between votes and proposals, and
// rejects shapes that decode but carry the wrong Type.
func TestClassifySignBytes(t *testing.T) {
	t.Parallel()

	t.Run("prevote", func(t *testing.T) {
		t.Parallel()
		bz := makeVoteBytes(t, 5, 1, types.PrevoteType, "test")
		h, r, s, err := classifySignBytes(bz)
		require.NoError(t, err)
		assert.Equal(t, int64(5), h)
		assert.Equal(t, 1, r)
		assert.EqualValues(t, 2, s) // StepPrevote = 2
	})

	t.Run("precommit", func(t *testing.T) {
		t.Parallel()
		bz := makeVoteBytes(t, 5, 1, types.PrecommitType, "test")
		h, r, s, err := classifySignBytes(bz)
		require.NoError(t, err)
		assert.Equal(t, int64(5), h)
		assert.Equal(t, 1, r)
		assert.EqualValues(t, 3, s) // StepPrecommit = 3
	})

	t.Run("proposal", func(t *testing.T) {
		t.Parallel()
		bz := makeProposalBytes(t, 5, 1, "test")
		h, r, s, err := classifySignBytes(bz)
		require.NoError(t, err)
		assert.Equal(t, int64(5), h)
		assert.Equal(t, 1, r)
		assert.EqualValues(t, 1, s) // StepPropose = 1
	})

	t.Run("garbage", func(t *testing.T) {
		t.Parallel()
		_, _, _, err := classifySignBytes([]byte("not a canonical message"))
		require.ErrorIs(t, err, ErrUnparseableSignBytes)
	})
}

// Constructor input validation.
func TestNewHRSGuardedSigner_BadInputs(t *testing.T) {
	t.Parallel()

	_, err := NewHRSGuardedSigner(nil, filepath.Join(t.TempDir(), "s.json"), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inner signer is nil")

	_, err = NewHRSGuardedSigner(newFakeSigner(), "", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "state file path is empty")
}
