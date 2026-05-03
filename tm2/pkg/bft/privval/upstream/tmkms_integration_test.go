//go:build tmkms_integration

package upstream_test

// tmkms_integration_test.go: end-to-end test against a real tmkms
// binary. Gated behind the tmkms_integration build tag (and a check
// for the binary on PATH) so it doesn't run in the default CI.
//
// To run locally:
//
//	go test -tags=tmkms_integration -count=1 ./tm2/pkg/bft/privval/upstream/...
//
// The dedicated CI workflow at .github/workflows/ci-tmkms-integration.yml
// installs a pinned tmkms release and runs this test.
//
// The test orchestrates tmkms with the softsign backend:
//   1. Generate three ed25519 keys: gnoland-identity (our SecretConn
//      identity), tmkms-identity (their SecretConn identity), and
//      consensus (the validator key tmkms holds).
//   2. Write them as base64 files (tmkms's softsign format) and a
//      tmkms.toml that pins protocol_version = "v0.34" and points
//      at our listener.
//   3. Start an upstream.SignerListenerEndpoint + SignerClient.
//   4. Spawn `tmkms start -c <toml>` as a subprocess pointed at us.
//   5. Init() blocks until tmkms dials in; verify the cached pubkey
//      matches the consensus key we wrote.
//   6. SignVote at heights 1 and 2 (no double-sign), SignProposal at
//      height 3. Verify each signature with the consensus pubkey.
//   7. Tear down: kill tmkms, close the endpoint.

import (
	"context"
	stded25519 "crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testChainID = "gno-tmkms-it"

	// Generous wait — tmkms cold-start + first dial can take a few
	// seconds in CI containers.
	testWaitForConnection = 20 * time.Second
	testRPCTimeout        = 10 * time.Second
)

func TestTmkmsIntegration_FullSigningFlow(t *testing.T) {
	tmkmsBin, err := exec.LookPath("tmkms")
	if err != nil {
		t.Skip("tmkms binary not on PATH — install tmkms or build from iqlusioninc/tmkms to run this test")
	}
	t.Logf("using tmkms binary: %s", tmkmsBin)

	tmpDir := t.TempDir()

	// --- 1. Keys --------------------------------------------------

	gnolandIdentitySeed := mustRandomSeed(t)
	tmkmsIdentitySeed := mustRandomSeed(t)
	consensusSeed := mustRandomSeed(t)

	gnolandIdentity := ed25519PrivFromSeed(gnolandIdentitySeed)
	tmkmsIdentityPub := stded25519.NewKeyFromSeed(tmkmsIdentitySeed[:]).Public().(stded25519.PublicKey)
	consensusPub := stded25519.NewKeyFromSeed(consensusSeed[:]).Public().(stded25519.PublicKey)

	// tmkms expects 32-byte ed25519 seed, base64-encoded, in a file.
	tmkmsIdentityKeyPath := filepath.Join(tmpDir, "kms-identity.key")
	consensusKeyPath := filepath.Join(tmpDir, "consensus.key")
	mustWriteBase64Seed(t, tmkmsIdentityKeyPath, tmkmsIdentitySeed[:])
	mustWriteBase64Seed(t, consensusKeyPath, consensusSeed[:])

	// --- 2. tmkms.toml -------------------------------------------

	listenAddr, listenPort := pickFreePort(t)
	tmkmsStateFile := filepath.Join(tmpDir, "consensus_state.json")
	tmkmsTomlPath := filepath.Join(tmpDir, "tmkms.toml")
	tmkmsToml := fmt.Sprintf(`
[[chain]]
id = "%s"
key_format = { type = "hex" }
state_file = "%s"

[[providers.softsign]]
chain_ids = ["%s"]
key_type = "consensus"
key_format = { type = "base64" }
path = "%s"

[[validator]]
chain_id = "%s"
addr = "tcp://127.0.0.1:%d"
secret_key = "%s"
protocol_version = "v0.34"
reconnect = false
`, testChainID, tmkmsStateFile, testChainID, consensusKeyPath, testChainID, listenPort, tmkmsIdentityKeyPath)
	require.NoError(t, os.WriteFile(tmkmsTomlPath, []byte(tmkmsToml), 0o600))

	// --- 3. Listener + endpoint + client ------------------------

	allowlist := []ed25519.PubKeyEd25519{}
	var pk ed25519.PubKeyEd25519
	copy(pk[:], tmkmsIdentityPub)
	allowlist = append(allowlist, pk)

	rawLn, err := net.Listen("tcp", listenAddr)
	require.NoError(t, err)
	tcpLn := rawLn.(*net.TCPListener)
	compatLn := upstream.NewTCPListener(tcpLn, gnolandIdentity, allowlist,
		upstream.TCPListenerTimeoutReadWrite(5*time.Second))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	endpoint := upstream.NewSignerListenerEndpoint(logger, compatLn,
		upstream.SignerListenerEndpointTimeoutReadWrite(5*time.Second))

	sc, err := upstream.NewSignerClient(endpoint, testChainID)
	require.NoError(t, err)

	// --- 4. Spawn tmkms ------------------------------------------

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tmkmsCmd := exec.CommandContext(ctx, tmkmsBin, "start", "-c", tmkmsTomlPath)
	tmkmsCmd.Stdout = &testLogWriter{t: t, prefix: "tmkms-stdout"}
	tmkmsCmd.Stderr = &testLogWriter{t: t, prefix: "tmkms-stderr"}
	require.NoError(t, tmkmsCmd.Start(),
		"failed to start tmkms — verify your tmkms build supports softsign + protocol v0.34")

	t.Cleanup(func() {
		cancel()
		// Allow a brief grace period for tmkms to exit cleanly.
		done := make(chan error, 1)
		go func() { done <- tmkmsCmd.Wait() }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = tmkmsCmd.Process.Kill()
		}
		_ = sc.Close()
	})

	// --- 5. Init blocks for tmkms dial-in + pubkey fetch ---------

	require.NoError(t, sc.Init(testWaitForConnection),
		"tmkms did not complete the SecretConnection handshake within %s", testWaitForConnection)

	gotPub := sc.PubKey().Bytes()
	wantPub := []byte(consensusPub)
	require.Equal(t, wantPub, gotPub,
		"validator pubkey reported by tmkms must match the consensus key we wrote to softsign")

	// --- 6. SignVote × 2 (monotonic), SignProposal ---------------

	vote1 := &types.Vote{
		Type:             types.PrecommitType,
		Height:           1,
		Round:            0,
		BlockID:          types.BlockID{Hash: bytesOfLen(0xaa, 32)},
		ValidatorAddress: addrOfLen(0x01),
	}
	mustSignAndVerify(t, sc, vote1, consensusPub)

	vote2 := &types.Vote{
		Type:             types.PrecommitType,
		Height:           2,
		Round:            0,
		BlockID:          types.BlockID{Hash: bytesOfLen(0xbb, 32)},
		ValidatorAddress: addrOfLen(0x01),
	}
	mustSignAndVerify(t, sc, vote2, consensusPub)

	prop := &types.Proposal{
		Type:     types.ProposalType,
		Height:   3,
		Round:    0,
		POLRound: -1,
		BlockID:  types.BlockID{Hash: bytesOfLen(0xcc, 32)},
	}
	pbBefore, err := proposalSignBytes(prop)
	require.NoError(t, err)
	require.NoError(t, sc.SignProposal(testChainID, prop), "SignProposal must round-trip against tmkms")
	require.NotEmpty(t, prop.Signature, "tmkms must populate Proposal.Signature")
	assert.True(t,
		stded25519.Verify(consensusPub, pbBefore, prop.Signature),
		"proposal signature must verify against the consensus pubkey")
}

// --- helpers ------------------------------------------------------

func mustSignAndVerify(t *testing.T, sc *upstream.SignerClient, vote *types.Vote, consPub stded25519.PublicKey) {
	t.Helper()
	signBytes := vote.SignBytes(testChainID)
	require.NoError(t, sc.SignVote(testChainID, vote),
		"SignVote h=%d r=%d must succeed", vote.Height, vote.Round)
	require.NotEmpty(t, vote.Signature, "tmkms must populate Vote.Signature for h=%d r=%d", vote.Height, vote.Round)
	assert.True(t,
		stded25519.Verify(consPub, signBytes, vote.Signature),
		"vote signature h=%d r=%d must verify against the consensus pubkey", vote.Height, vote.Round)
}

func proposalSignBytes(p *types.Proposal) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()
	return p.SignBytes(testChainID), nil
}

func mustRandomSeed(t *testing.T) [32]byte {
	t.Helper()
	var seed [32]byte
	_, err := rand.Read(seed[:])
	require.NoError(t, err)
	return seed
}

// ed25519PrivFromSeed expands a 32-byte seed into tm2's
// ed25519.PrivKeyEd25519 ([64]byte: seed || pubkey). Matches what
// stdlib ed25519.NewKeyFromSeed produces.
func ed25519PrivFromSeed(seed [32]byte) ed25519.PrivKeyEd25519 {
	stdPriv := stded25519.NewKeyFromSeed(seed[:])
	var out ed25519.PrivKeyEd25519
	copy(out[:], stdPriv)
	return out
}

func mustWriteBase64Seed(t *testing.T, path string, seed []byte) {
	t.Helper()
	enc := base64.StdEncoding.EncodeToString(seed)
	require.NoError(t, os.WriteFile(path, []byte(enc), 0o600))
}

func pickFreePort(t *testing.T) (addr string, port int) {
	t.Helper()
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port = probe.Addr().(*net.TCPAddr).Port
	require.NoError(t, probe.Close())
	return fmt.Sprintf("127.0.0.1:%d", port), port
}

func bytesOfLen(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}

func addrOfLen(b byte) (a [20]byte) {
	for i := range a {
		a[i] = b
	}
	return
}

// testLogWriter pipes a subprocess's stdout/stderr into the test log
// so we get tmkms diagnostics on failure.
type testLogWriter struct {
	t      *testing.T
	prefix string
	closed atomic.Bool
}

func (w *testLogWriter) Write(p []byte) (int, error) {
	if w.closed.Load() {
		return 0, errors.New("closed")
	}
	w.t.Logf("[%s] %s", w.prefix, string(p))
	return len(p), nil
}
