package supervisor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/upgrades"
)

const testChainID = "test-chain-1"

// leg describes one ledger entry of a fixture: halt 0 is the genesis, -1 a
// rolling release.
type fixtureEntry struct {
	version string
	halt    int64
	pending bool // no image digest nor binaries yet (last entry only)
}

const rolling = -1

// writeFixture writes a ledger and a matching genesis into dir and returns
// their paths. Every field the ledger rules require is filled in from the
// version, so the fixture passes upgrades.Ledger.Validate.
func writeFixture(t *testing.T, dir string, entries ...fixtureEntry) (ledgerPath, genesisPath string) {
	t.Helper()
	genesis := []byte(`{"genesis_time":"2026-01-01T00:00:00Z","chain_id":"` + testChainID + `","app_state":{}}` + "\n")
	genesisPath = filepath.Join(dir, "genesis.json")
	require.NoError(t, os.WriteFile(genesisPath, genesis, 0o644))
	sum := sha256.Sum256(genesis)

	l := fixtureLedger(t, hex.EncodeToString(sum[:]), entries...)
	data, err := json.MarshalIndent(l, "", "  ")
	require.NoError(t, err)
	ledgerPath = filepath.Join(dir, upgrades.LedgerFile)
	require.NoError(t, os.WriteFile(ledgerPath, data, 0o644))
	return ledgerPath, genesisPath
}

func fixtureLedger(t *testing.T, genesisSHA string, entries ...fixtureEntry) *upgrades.Ledger {
	t.Helper()
	l := &upgrades.Ledger{
		SchemaVersion: upgrades.SchemaVersion,
		ChainID:       testChainID,
		GenesisSHA256: genesisSHA,
		GenesisTime:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	at := l.GenesisTime
	for i, f := range entries {
		e := upgrades.Entry{
			Kind:    upgrades.KindUpgrade,
			Version: f.version,
			Commit:  strings.Repeat(fmt.Sprintf("%x", i%16), 40),
			Image:   upgrades.Image{Ref: "ghcr.io/gnolang/gno/gnoland:" + f.version},
			Release: "https://github.com/gnolang/gno/releases/tag/" + f.version,
		}
		switch {
		case i == 0:
			e.Kind = upgrades.KindGenesis
		case f.halt == rolling:
			e.Kind = upgrades.KindRolling
		default:
			h := f.halt
			at = at.Add(time.Hour)
			when := at
			p := int64(i)
			e.HaltHeight, e.HaltTime, e.Proposal = &h, &when, &p
		}
		if !f.pending {
			d := "sha256:" + strings.Repeat(fmt.Sprintf("%x", (i+1)%16), 64)
			e.Image.Digest = &d
			e.Binaries = map[string]string{
				"linux/amd64": "https://example.invalid/gnoland?checksum=sha256:" + strings.Repeat("a", 64),
			}
		}
		l.Upgrades = append(l.Upgrades, e)
	}
	require.NoError(t, l.Validate())
	return l
}
