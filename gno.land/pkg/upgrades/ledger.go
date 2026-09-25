// Package upgrades is the ledger of every binary a gno.land network has run.
//
// misc/deployments/<chain>/upgrades.json is the source of truth; UPGRADES.md
// next to it carries a table rendered from the JSON between two markers. This
// package owns the format: parsing, the invariants a replaying supervisor
// relies on, the rendering, and the check that the two files agree. The rules
// live in Go, tested, rather than in a shell script forked per chain.
//
// The contract: a version runs the blocks from the previous entry's
// halt_height + 1 (1 for genesis) up to and including its own successor's
// halt_height. Entries are in chain order: strictly increasing versions and
// strictly increasing halt heights.
package upgrades

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	// LedgerFile and DocFile are the two files a deployment directory holds.
	LedgerFile = "upgrades.json"
	DocFile    = "UPGRADES.md"

	// BeginMarker and EndMarker delimit the generated table inside DocFile.
	BeginMarker = "<!-- BEGIN GENERATED (gno.land/pkg/upgrades) -->"
	EndMarker   = "<!-- END GENERATED -->"

	// SchemaVersion is the only format this package reads and writes.
	SchemaVersion = 1

	pendingCell = "*(pending)*"
	notSetCell  = "*(not set)*"
	noneCell    = "—"
	repoURL     = "https://github.com/gnolang/gno"
	govdaoURL   = "https://gno.land/r/gov/dao"
	imagePrefix = "ghcr.io/gnolang/gno/gnoland:"
)

// Kind says what an entry is: the genesis the chain started on, or a
// coordinated upgrade that changed the binary at a halt height.
type Kind string

const (
	KindGenesis Kind = "genesis"
	KindUpgrade Kind = "upgrade"
)

// Ledger is one chain's upgrades.json.
type Ledger struct {
	Schema        string    `json:"$schema,omitempty"`
	SchemaVersion int       `json:"schema_version"`
	ChainID       string    `json:"chain_id"`
	GenesisSHA256 string    `json:"genesis_sha256"`
	GenesisTime   time.Time `json:"genesis_time"`
	Upgrades      []Entry   `json:"upgrades"`
}

// Entry is one binary the network ran. Pointer fields are null in the JSON
// until the fact they record has happened: the digest until CI built the
// image, the halt time until the halt, the proposal until it was created.
// HaltMinVersion is null when the proposal set no floor; an empty string is
// refused, so "the gate was off" and "someone forgot" cannot be confused.
type Entry struct {
	Kind           Kind              `json:"kind"`
	Version        string            `json:"version"`
	Commit         string            `json:"commit"`
	HaltHeight     *int64            `json:"halt_height"`
	HaltTime       *time.Time        `json:"halt_time"`
	HaltMinVersion *string           `json:"halt_min_version"`
	Proposal       *int64            `json:"proposal"`
	Image          Image             `json:"image"`
	Binaries       map[string]string `json:"binaries"`
	RanAs          *string           `json:"ran_as"`
	Release        string            `json:"release"`
}

// Image is the gnoland container image for an entry: the tag operators pin,
// and the digest that tag resolved to when the ledger was written.
type Image struct {
	Ref    string  `json:"ref"`
	Digest *string `json:"digest"`
}

// Range is the blocks one entry's version produced or replays.
// To is 0 for the current version, which has no successor yet.
type Range struct {
	Version  string
	From, To int64
}

var (
	hex40     = regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64     = regexp.MustCompile(`^[0-9a-f]{64}$`)
	digestRE  = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	platform  = regexp.MustCompile(`^[a-z0-9]+/[a-z0-9]+$`)
	checksumQ = regexp.MustCompile(`[?&]checksum=sha256:[0-9a-f]{64}$`)
)

// Parse decodes a ledger. Unknown fields are an error: a misspelled field
// would otherwise vanish silently, which for a null-able field looks exactly
// like "pending".
func Parse(data []byte) (*Ledger, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var l Ledger
	if err := dec.Decode(&l); err != nil {
		return nil, fmt.Errorf("parse %s: %w", LedgerFile, err)
	}
	return &l, nil
}

// Load reads and parses dir/upgrades.json.
func Load(dir string) (*Ledger, error) {
	data, err := os.ReadFile(filepath.Join(dir, LedgerFile))
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// ParseVersion normalises a release tag the way the node does
// (gno.land/pkg/gnoland.parseReleaseVersion): vMAJOR.MINOR.PATCH with an
// optional pre-release, build metadata dropped. The ledger only ever holds
// tags the node can order, because halt_min_version is compared with them.
func ParseVersion(v string) (string, bool) {
	if !semver.IsValid(v) {
		return "", false
	}
	v = strings.TrimSuffix(v, semver.Build(v))
	if semver.Canonical(v) != v {
		return "", false
	}
	return v, true
}

// finalVersion strips a pre-release suffix: v1.6.0-rc.2 rehearses v1.6.0.
func finalVersion(v string) string {
	if c, ok := ParseVersion(v); ok {
		return strings.TrimSuffix(c, semver.Prerelease(c))
	}
	return v
}

// Validate checks every invariant and reports all violations at once.
func (l *Ledger) Validate() error {
	var problems []string
	fail := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	if l.SchemaVersion != SchemaVersion {
		fail("schema_version is %d, this tooling reads %d", l.SchemaVersion, SchemaVersion)
	}
	if l.ChainID == "" {
		fail("chain_id is empty")
	}
	if !hex64.MatchString(l.GenesisSHA256) {
		fail("genesis_sha256 %q is not 64 lowercase hex characters", l.GenesisSHA256)
	}
	if l.GenesisTime.IsZero() {
		fail("genesis_time is missing")
	}
	if len(l.Upgrades) == 0 {
		fail("no entries: the ledger needs at least the genesis")
		return errors.New(strings.Join(problems, "\n"))
	}

	var (
		prevVersion string
		prevHeight  int64
		prevTime    time.Time
		genesisSeen int
	)
	for i, e := range l.Upgrades {
		at := fmt.Sprintf("upgrades[%d] (%s)", i, e.Version)

		switch e.Kind {
		case KindGenesis:
			genesisSeen++
			if i != 0 {
				fail("%s: kind genesis is only valid for the first entry", at)
			}
			if e.HaltHeight != nil {
				fail("%s: genesis has no halt_height", at)
			}
		case KindUpgrade:
			if i == 0 {
				fail("%s: the first entry must be the genesis, got kind %q", at, e.Kind)
			}
			if e.HaltHeight == nil {
				fail("%s: an upgrade needs the halt_height that activated it", at)
			} else if *e.HaltHeight <= prevHeight {
				fail("%s: halt_height %d does not increase on the previous entry's %d", at, *e.HaltHeight, prevHeight)
			}
		default:
			fail("%s: unknown kind %q", at, e.Kind)
		}
		if e.HaltHeight != nil {
			prevHeight = *e.HaltHeight
		}

		canon, ok := ParseVersion(e.Version)
		switch {
		case !ok:
			fail("%s: version %q is not a release tag the node can order", at, e.Version)
		case prevVersion != "" && semver.Compare(canon, prevVersion) <= 0:
			fail("%s: version does not increase on the previous entry's %s", at, prevVersion)
		}
		if ok {
			prevVersion = canon
		}

		if !hex40.MatchString(e.Commit) {
			fail("%s: commit %q is not a 40-character lowercase sha", at, e.Commit)
		}
		if e.HaltTime != nil {
			if !prevTime.IsZero() && !e.HaltTime.After(prevTime) {
				fail("%s: halt_time %s does not follow the previous entry's %s", at, e.HaltTime.UTC().Format(time.RFC3339), prevTime.UTC().Format(time.RFC3339))
			}
			prevTime = *e.HaltTime
		}
		if e.HaltMinVersion != nil {
			switch mv, mok := ParseVersion(*e.HaltMinVersion); {
			case *e.HaltMinVersion == "":
				fail("%s: halt_min_version is \"\"; use null when the proposal set no floor", at)
			case !mok:
				fail("%s: halt_min_version %q is not a release tag the node can parse", at, *e.HaltMinVersion)
			case ok && semver.Compare(mv, canon) > 0:
				fail("%s: halt_min_version %s is above the version that ran (%s)", at, *e.HaltMinVersion, e.Version)
			}
		}
		if e.Proposal != nil && *e.Proposal < 0 {
			fail("%s: proposal %d is negative", at, *e.Proposal)
		}
		if e.Image.Ref != imagePrefix+e.Version {
			fail("%s: image.ref %q is not %s%s", at, e.Image.Ref, imagePrefix, e.Version)
		}
		if e.Image.Digest != nil && !digestRE.MatchString(*e.Image.Digest) {
			fail("%s: image.digest %q is not sha256: followed by 64 hex characters", at, *e.Image.Digest)
		}
		if len(e.Binaries) == 0 {
			fail("%s: binaries is empty; a supervisor needs one download per platform", at)
		}
		for plat, url := range e.Binaries {
			if !platform.MatchString(plat) {
				fail("%s: binaries key %q is not a platform of the form os/arch", at, plat)
			}
			if !checksumQ.MatchString(url) {
				fail("%s: binaries[%s] has no ?checksum=sha256:<64 hex> suffix", at, plat)
			}
		}
		if e.RanAs != nil && !digestRE.MatchString(*e.RanAs) {
			fail("%s: ran_as %q is not sha256: followed by 64 hex characters", at, *e.RanAs)
		}
		if e.Release == "" {
			fail("%s: release link is empty", at)
		}
	}
	if genesisSeen != 1 {
		fail("exactly one genesis entry is expected, found %d", genesisSeen)
	}

	if len(problems) == 0 {
		return nil
	}
	return errors.New(strings.Join(problems, "\n"))
}

// Has reports whether the ledger has an entry for v, ignoring a pre-release
// suffix on either side: a release candidate rehearses the final version's
// entry rather than getting one of its own.
func (l *Ledger) Has(v string) bool {
	want := finalVersion(v)
	for _, e := range l.Upgrades {
		if e.Version == v || finalVersion(e.Version) == want {
			return true
		}
	}
	return false
}

// BlockRanges is the contract made explicit: which version produced which
// blocks. The last range's To is 0 because the current version is still running.
func (l *Ledger) BlockRanges() []Range {
	ranges := make([]Range, 0, len(l.Upgrades))
	for i, e := range l.Upgrades {
		from := int64(1)
		if i > 0 && l.Upgrades[i-1].HaltHeight != nil {
			from = *l.Upgrades[i-1].HaltHeight + 1
		}
		if i > 0 && e.HaltHeight != nil {
			from = *e.HaltHeight + 1
		}
		var to int64
		if i+1 < len(l.Upgrades) && l.Upgrades[i+1].HaltHeight != nil {
			to = *l.Upgrades[i+1].HaltHeight
		}
		ranges = append(ranges, Range{Version: e.Version, From: from, To: to})
	}
	return ranges
}

// RenderTable renders the Markdown table, markers included, with a readable
// word in every cell whose fact has not happened yet.
func (l *Ledger) RenderTable() string {
	var b strings.Builder
	b.WriteString(BeginMarker + "\n")
	b.WriteString("| Version | Commit | Halt height | Halt time (UTC) | halt_min_version | GovDAO proposal | gnoland image digest |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, e := range l.Upgrades {
		height, when, minVersion, proposal := "genesis", l.GenesisTime.UTC().Format(time.RFC3339), noneCell, noneCell
		if e.Kind != KindGenesis {
			height, when, minVersion, proposal = pendingCell, pendingCell, notSetCell, pendingCell
			if e.HaltHeight != nil {
				height = fmt.Sprint(*e.HaltHeight)
			}
			if e.HaltTime != nil {
				when = e.HaltTime.UTC().Format(time.RFC3339)
			}
			if e.HaltMinVersion != nil {
				minVersion = "`" + *e.HaltMinVersion + "`"
			}
			if e.Proposal != nil {
				proposal = fmt.Sprintf("[#%d](%s:%d)", *e.Proposal, govdaoURL, *e.Proposal)
			}
		}
		digest := pendingCell
		if e.Image.Digest != nil {
			digest = "`" + (*e.Image.Digest)[:19] + "…`"
		}
		fmt.Fprintf(&b, "| [%s](%s) | [%s](%s/commit/%s) | %s | %s | %s | %s | %s |\n",
			e.Version, e.Release, e.Commit[:9], repoURL, e.Commit, height, when, minVersion, proposal, digest)
	}
	b.WriteString(EndMarker)
	return b.String()
}

// Splice replaces the generated block of doc, markers included, with table.
func Splice(doc []byte, table string) ([]byte, error) {
	begin := bytes.Index(doc, []byte(BeginMarker))
	if begin < 0 {
		return nil, fmt.Errorf("%s has no %q marker", DocFile, BeginMarker)
	}
	end := bytes.Index(doc[begin:], []byte(EndMarker))
	if end < 0 {
		return nil, fmt.Errorf("%s has no %q marker after the begin marker", DocFile, EndMarker)
	}
	end += begin + len(EndMarker)
	out := make([]byte, 0, len(doc)-(end-begin)+len(table))
	out = append(out, doc[:begin]...)
	out = append(out, table...)
	out = append(out, doc[end:]...)
	return out, nil
}

// rendered returns dir's document with a fresh table spliced in, after
// validating the ledger: a ledger the rules refuse is never rendered.
func rendered(dir string) (current, fresh []byte, err error) {
	l, err := Load(dir)
	if err != nil {
		return nil, nil, err
	}
	if err := l.Validate(); err != nil {
		return nil, nil, fmt.Errorf("%s/%s is invalid:\n%w", dir, LedgerFile, err)
	}
	current, err = os.ReadFile(filepath.Join(dir, DocFile))
	if err != nil {
		return nil, nil, err
	}
	fresh, err = Splice(current, l.RenderTable())
	return current, fresh, err
}

// Render rewrites dir/UPGRADES.md from dir/upgrades.json.
func Render(dir string) error {
	_, fresh, err := rendered(dir)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, DocFile), fresh, 0o644)
}

// Check fails if dir/upgrades.json is invalid or dir/UPGRADES.md is stale.
func Check(dir string) error {
	current, fresh, err := rendered(dir)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, fresh) {
		return fmt.Errorf("%s/%s is stale: render it from %s and commit the result", dir, DocFile, LedgerFile)
	}
	return nil
}
