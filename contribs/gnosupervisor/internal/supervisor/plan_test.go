package supervisor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/upgrades"
)

func TestResolve(t *testing.T) {
	t.Parallel()
	p := newPlan(fixtureLedger(t, testSHA,
		fixtureEntry{version: "v1.0.0"},
		fixtureEntry{version: "v1.1.0", halt: 100},
		fixtureEntry{version: "v1.1.1", halt: rolling},
		fixtureEntry{version: "v1.1.2", halt: rolling},
		fixtureEntry{version: "v1.2.0-rc.2", halt: 200},
	))
	for _, tc := range []struct {
		height  int64
		version string
		from    int64
		to      int64
	}{
		{1, "v1.0.0", 1, 100},
		{100, "v1.0.0", 1, 100}, // the halt height is the old release's last block
		{101, "v1.1.2", 101, 200},
		{200, "v1.1.2", 101, 200}, // the last rolling patch serves the range
		{201, "v1.2.0-rc.2", 201, 0},
		{1 << 40, "v1.2.0-rc.2", 201, 0},
	} {
		leg, err := p.Resolve(tc.height)
		require.NoError(t, err, "height %d", tc.height)
		assert.Equal(t, tc.version, leg.Version, "height %d", tc.height)
		assert.Equal(t, tc.version, leg.Entry.Version, "height %d", tc.height)
		assert.Equal(t, tc.from, leg.Range.From, "height %d", tc.height)
		assert.Equal(t, tc.to, leg.Range.To, "height %d", tc.height)
		assert.Equal(t, tc.to == 0, leg.Last())
	}

	_, err := p.Resolve(0)
	assert.ErrorContains(t, err, "first block is 1")
}

func TestResolvePendingLastEntry(t *testing.T) {
	t.Parallel()
	p := newPlan(fixtureLedger(t, testSHA,
		fixtureEntry{version: "v1.0.0"},
		fixtureEntry{version: "v1.1.0", halt: 50, pending: true},
	))
	leg, err := p.Resolve(50)
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", leg.Version)

	_, err = p.Resolve(51)
	assert.ErrorContains(t, err, "no image digest yet")
}

// A rolling patch whose build is still pending does not hold the range up:
// the release it patches serves it meanwhile.
func TestResolvePendingRollingPatch(t *testing.T) {
	t.Parallel()
	p := newPlan(fixtureLedger(t, testSHA,
		fixtureEntry{version: "v1.0.0"},
		fixtureEntry{version: "v1.1.0", halt: 50},
		fixtureEntry{version: "v1.1.1", halt: rolling, pending: true},
	))
	leg, err := p.Resolve(60)
	require.NoError(t, err)
	assert.Equal(t, "v1.1.0", leg.Version)
}

// The mainnet ledger in this repository, resolved at its edges: the joining
// path VALIDATOR.md and UPGRADES.md describe, with rolling patches serving
// their ranges.
func TestResolveMainnetLedger(t *testing.T) {
	t.Parallel()
	l, err := upgrades.Load(filepath.Join("..", "..", "..", "..", "misc", "deployments", "mainnet.gno.land"))
	require.NoError(t, err)
	require.NoError(t, l.Validate())
	p := newPlan(l)
	for height, want := range map[int64]string{
		1:       "v1.2.0",
		36300:   "v1.2.0",
		36301:   "v1.3.0",
		113000:  "v1.3.0",
		113001:  "v1.4.0",
		162200:  "v1.4.0",
		162201:  "v1.5.1",
		1183000: "v1.5.1",
		1183001: "v1.7.1",
		1651500: "v1.7.1",
		1651501: "v1.8.0",
	} {
		leg, err := p.Resolve(height)
		require.NoError(t, err, "height %d", height)
		assert.Equal(t, want, leg.Version, "height %d", height)
		assert.NotNil(t, leg.Entry.Image.Digest)
	}
}

func TestLoadPlan(t *testing.T) {
	t.Parallel()
	entries := []fixtureEntry{{version: "v1.0.0"}, {version: "v1.1.0", halt: 10}}

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		ledger, genesis := writeFixture(t, t.TempDir(), entries...)
		p, err := LoadPlan(ledger, genesis)
		require.NoError(t, err)
		assert.Equal(t, testChainID, p.ChainID())
	})

	t.Run("genesis of another chain", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		ledger, genesis := writeFixture(t, dir, entries...)
		require.NoError(t, os.WriteFile(genesis, []byte(`{"chain_id":"other-1"}`), 0o644))
		_, err := LoadPlan(ledger, genesis)
		assert.ErrorContains(t, err, `is for chain "other-1"`)
	})

	t.Run("genesis with other content", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		ledger, genesis := writeFixture(t, dir, entries...)
		data, err := os.ReadFile(genesis)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(genesis, append(data, ' '), 0o644))
		_, err = LoadPlan(ledger, genesis)
		assert.ErrorContains(t, err, "the ledger records")
	})

	t.Run("invalid ledger", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		ledger, genesis := writeFixture(t, dir, entries...)
		data, err := os.ReadFile(ledger)
		require.NoError(t, err)
		// Versions out of order: Validate refuses it, so the supervisor does.
		data = []byte(strings.Replace(string(data), `"version": "v1.1.0"`, `"version": "v0.9.0"`, 1))
		data = []byte(strings.Replace(string(data), `gnoland:v1.1.0`, `gnoland:v0.9.0`, 1))
		data = []byte(strings.Replace(string(data), `tag/v1.1.0`, `tag/v0.9.0`, 1))
		require.NoError(t, os.WriteFile(ledger, data, 0o644))
		_, err = LoadPlan(ledger, genesis)
		assert.ErrorContains(t, err, "does not increase")
	})

	t.Run("unknown field", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		ledger, genesis := writeFixture(t, dir, entries...)
		data, err := os.ReadFile(ledger)
		require.NoError(t, err)
		data = []byte(strings.Replace(string(data), `"chain_id"`, `"chainid": "x", "chain_id"`, 1))
		require.NoError(t, os.WriteFile(ledger, data, 0o644))
		_, err = LoadPlan(ledger, genesis)
		assert.ErrorContains(t, err, "unknown field")
	})
}

const testSHA = "0000000000000000000000000000000000000000000000000000000000000000"
