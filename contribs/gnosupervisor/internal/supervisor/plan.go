package supervisor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/upgrades"
)

// Leg is one stretch of the replay: the release that runs the blocks From..To
// (To is 0 for the last range, which has no end yet).
type Leg struct {
	// Range is the ledger range, as upgrades.BlockRanges returns it.
	Range upgrades.Range
	// Version is the release that runs the range: the range's own version,
	// or the last rolling patch released into it.
	Version string
	// Entry is the ledger entry for Version; its image digest names the
	// artifact.
	Entry upgrades.Entry
}

// Last reports whether the leg is the open-ended last range.
func (l Leg) Last() bool { return l.Range.To == 0 }

// Plan is a validated ledger, checked against the node's genesis.
type Plan struct {
	ledger *upgrades.Ledger
	ranges []upgrades.Range
}

// LoadPlan reads and validates the ledger at ledgerPath and checks that
// genesisPath is the genesis it records: same chain id, same sha256.
func LoadPlan(ledgerPath, genesisPath string) (*Plan, error) {
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		return nil, fmt.Errorf("read ledger: %w", err)
	}
	l, err := upgrades.Parse(data)
	if err != nil {
		return nil, err
	}
	if err := l.Validate(); err != nil {
		return nil, fmt.Errorf("ledger %s is invalid:\n%w", ledgerPath, err)
	}
	if err := checkGenesis(l, genesisPath); err != nil {
		return nil, err
	}
	return newPlan(l), nil
}

// newPlan wraps a ledger that has passed Validate.
func newPlan(l *upgrades.Ledger) *Plan {
	return &Plan{ledger: l, ranges: l.BlockRanges()}
}

// ChainID is the chain the ledger describes.
func (p *Plan) ChainID() string { return p.ledger.ChainID }

// Resolve returns the leg that must execute block height. The range rule is
// upgrades.BlockRanges; this only picks the range that holds height.
func (p *Plan) Resolve(height int64) (Leg, error) {
	if height < 1 {
		return Leg{}, fmt.Errorf("block height %d: the first block is 1", height)
	}
	for _, r := range p.ranges {
		if height < r.From || (r.To != 0 && height > r.To) {
			continue
		}
		// A rolling patch changes no consensus code and fixes something in the
		// release it patches, so the last one that is built serves the range.
		// Only the ledger's last entry can still be pending its build.
		candidates := append([]string{r.Version}, r.Rolling...)
		for i := len(candidates) - 1; i >= 0; i-- {
			version := candidates[i]
			e, ok := p.ledger.Entry(version)
			if !ok || e.Version != version {
				return Leg{}, fmt.Errorf("ledger has no entry for %s", version)
			}
			if e.Image.Digest != nil {
				return Leg{Range: r, Version: version, Entry: e}, nil
			}
		}
		return Leg{}, fmt.Errorf("%s runs block %d, but its ledger entry has no image digest yet (pending its build)", r.Version, height)
	}
	// BlockRanges starts at 1 and its last range is open, so this means the
	// ledger changed shape under us.
	return Leg{}, fmt.Errorf("no ledger range holds block %d", height)
}

// HaltAt reports whether the ledger records a coordinated halt at height h,
// the end of a range.
func (p *Plan) HaltAt(h int64) bool {
	leg, err := p.Resolve(h)
	return err == nil && !leg.Last() && leg.Range.To == h
}

// checkGenesis refuses a genesis file that is not the one the ledger was
// written for: replaying another chain's history with this ledger's binaries
// would fail at best and fork at worst.
func checkGenesis(l *upgrades.Ledger, genesisPath string) error {
	f, err := os.Open(genesisPath)
	if err != nil {
		return fmt.Errorf("open genesis: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	var doc struct {
		ChainID string `json:"chain_id"`
	}
	if err := json.NewDecoder(io.TeeReader(f, h)).Decode(&doc); err != nil {
		return fmt.Errorf("decode genesis %s: %w", genesisPath, err)
	}
	// The decoder may stop before EOF; hash the rest so the sum covers the file.
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("read genesis %s: %w", genesisPath, err)
	}
	if doc.ChainID != l.ChainID {
		return fmt.Errorf("genesis %s is for chain %q, the ledger is for %q", genesisPath, doc.ChainID, l.ChainID)
	}
	if sum := hex.EncodeToString(h.Sum(nil)); sum != l.GenesisSHA256 {
		return errors.New("genesis " + genesisPath + " has sha256 " + sum + ", the ledger records " + l.GenesisSHA256)
	}
	return nil
}
