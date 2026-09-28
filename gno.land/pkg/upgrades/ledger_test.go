package upgrades

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validLedger = `{
  "$schema": "../upgrades/upgrades.schema.json",
  "schema_version": 1,
  "chain_id": "gnoland-1",
  "genesis_sha256": "ea22691003130eae3ba975b7d16460706b5d75ce6c04ae82c0c4faeab7de91f0",
  "genesis_time": "2026-09-12T15:00:00Z",
  "upgrades": [
    {
      "kind": "genesis",
      "version": "v1.2.0",
      "commit": "9c8eb132e483d6fd324d92c193e629ad65a98a37",
      "halt_height": null,
      "halt_time": null,
      "halt_min_version": null,
      "proposal": null,
      "image": {"ref": "ghcr.io/gnolang/gno/gnoland:v1.2.0", "digest": "sha256:4b161a2b5d5fcceab4badd4d519ea73465c078a17e6fb1e9b0bae2e01c1a7b87"},
      "binaries": {
        "linux/amd64": "https://github.com/gnolang/gno/releases/download/v1.2.0/gnoland_linux_amd64?checksum=sha256:1111111111111111111111111111111111111111111111111111111111111111"
      },
      "ran_as": null,
      "release": "https://github.com/gnolang/gno/releases/tag/v1.2.0"
    },
    {
      "kind": "upgrade",
      "version": "v1.3.0",
      "commit": "31b6650a100d9baf14e7669f8f0df924f1f841e0",
      "halt_height": 36300,
      "halt_time": "2026-09-14T09:17:06Z",
      "halt_min_version": null,
      "proposal": 0,
      "image": {"ref": "ghcr.io/gnolang/gno/gnoland:v1.3.0", "digest": null},
      "binaries": {
        "linux/amd64": "https://github.com/gnolang/gno/releases/download/v1.3.0/gnoland_linux_amd64?checksum=sha256:2222222222222222222222222222222222222222222222222222222222222222"
      },
      "ran_as": "sha256:7fffc5ac2d21d6608bed35b31cfc3f2aaefed0d6bb3cdba1d5edbfd905127a3f",
      "release": "https://github.com/gnolang/gno/releases/tag/v1.3.0"
    },
    {
      "kind": "upgrade",
      "version": "v1.4.0-rc.1",
      "commit": "00417a1be97b9a311d9669ae7aa9585b277ee594",
      "halt_height": 113000,
      "halt_time": null,
      "halt_min_version": "v1.4.0-rc.1",
      "proposal": null,
      "image": {"ref": "ghcr.io/gnolang/gno/gnoland:v1.4.0-rc.1", "digest": null},
      "binaries": {
        "linux/amd64": "https://github.com/gnolang/gno/releases/download/v1.4.0-rc.1/gnoland_linux_amd64?checksum=sha256:3333333333333333333333333333333333333333333333333333333333333333"
      },
      "ran_as": null,
      "release": "https://github.com/gnolang/gno/releases/tag/v1.4.0-rc.1"
    }
  ]
}`

func parseValid(t *testing.T) *Ledger {
	t.Helper()
	l, err := Parse([]byte(validLedger))
	require.NoError(t, err)
	return l
}

func TestParseAndValidate_acceptsTheReferenceLedger(t *testing.T) {
	t.Parallel()
	l := parseValid(t)
	require.NoError(t, l.Validate())
	assert.Equal(t, "gnoland-1", l.ChainID)
	assert.Len(t, l.Upgrades, 3)
	assert.Equal(t, KindGenesis, l.Upgrades[0].Kind)
	assert.Nil(t, l.Upgrades[0].HaltHeight)
	require.NotNil(t, l.Upgrades[1].HaltHeight)
	assert.EqualValues(t, 36300, *l.Upgrades[1].HaltHeight)
}

// Every rule the renderer and a replaying supervisor rely on, one mutation each.
func TestValidate_refusesEachBrokenInvariant(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		mutate func(l *Ledger)
		want   string
	}{
		{"unknown schema version", func(l *Ledger) { l.SchemaVersion = 2 }, "schema_version"},
		{"empty chain id", func(l *Ledger) { l.ChainID = "" }, "chain_id"},
		{"genesis sha not 64 hex", func(l *Ledger) { l.GenesisSHA256 = "abc" }, "genesis_sha256"},
		{"no entries", func(l *Ledger) { l.Upgrades = nil }, "no entries"},
		{"first entry is not genesis", func(l *Ledger) { l.Upgrades[0].Kind = KindUpgrade; h := int64(1); l.Upgrades[0].HaltHeight = &h }, "first entry"},
		{"two genesis entries", func(l *Ledger) { l.Upgrades[1].Kind = KindGenesis; l.Upgrades[1].HaltHeight = nil }, "genesis"},
		{"unknown kind", func(l *Ledger) { l.Upgrades[1].Kind = "hotfix" }, "kind"},
		{"genesis with a halt height", func(l *Ledger) { h := int64(5); l.Upgrades[0].HaltHeight = &h }, "halt_height"},
		{"upgrade without a halt height", func(l *Ledger) { l.Upgrades[1].HaltHeight = nil }, "halt_height"},
		{"halt heights not increasing", func(l *Ledger) { h := int64(36300); l.Upgrades[2].HaltHeight = &h }, "halt_height"},
		{"versions not increasing", func(l *Ledger) { l.Upgrades[2].Version = "v1.2.5" }, "version"},
		{"duplicate version", func(l *Ledger) { l.Upgrades[2].Version = "v1.3.0" }, "version"},
		{"version not a release tag", func(l *Ledger) { l.Upgrades[1].Version = "chain/mainnet" }, "version"},
		{"commit not 40 hex", func(l *Ledger) { l.Upgrades[1].Commit = "31b6650a1" }, "commit"},
		{"digest without prefix", func(l *Ledger) {
			d := "7fffc5ac2d21d6608bed35b31cfc3f2aaefed0d6bb3cdba1d5edbfd905127a3f"
			l.Upgrades[1].Image.Digest = &d
		}, "digest"},
		{"empty halt_min_version is ambiguous", func(l *Ledger) { s := ""; l.Upgrades[1].HaltMinVersion = &s }, "halt_min_version"},
		{"halt_min_version not a release tag", func(l *Ledger) { s := "latest"; l.Upgrades[1].HaltMinVersion = &s }, "halt_min_version"},
		{"halt_min_version above the version that ran", func(l *Ledger) { s := "v9.0.0"; l.Upgrades[1].HaltMinVersion = &s }, "halt_min_version"},
		{"halt times not increasing", func(l *Ledger) { t2 := l.Upgrades[1].HaltTime.Add(-1); l.Upgrades[2].HaltTime = &t2 }, "halt_time"},
		{"negative proposal", func(l *Ledger) { p := int64(-1); l.Upgrades[1].Proposal = &p }, "proposal"},
		{"image ref for another version", func(l *Ledger) { l.Upgrades[1].Image.Ref = "ghcr.io/gnolang/gno/gnoland:v1.9.0" }, "image.ref"},
		{"no binaries", func(l *Ledger) { l.Upgrades[1].Binaries = nil }, "binaries"},
		{"binary without checksum", func(l *Ledger) { l.Upgrades[1].Binaries["linux/amd64"] = "https://example.com/gnoland" }, "checksum"},
		{"binary platform not os/arch", func(l *Ledger) { l.Upgrades[1].Binaries["linux"] = l.Upgrades[1].Binaries["linux/amd64"] }, "platform"},
		{"ran_as malformed", func(l *Ledger) { s := "sha-31b6650"; l.Upgrades[1].RanAs = &s }, "ran_as"},
		{"missing release link", func(l *Ledger) { l.Upgrades[1].Release = "" }, "release"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			l := parseValid(t)
			tc.mutate(l)
			err := l.Validate()
			require.Error(t, err, "mutation %q was accepted", tc.name)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestValidate_reportsEveryProblemAtOnce(t *testing.T) {
	t.Parallel()
	l := parseValid(t)
	l.ChainID = ""
	l.Upgrades[1].Commit = "short"
	err := l.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chain_id")
	assert.Contains(t, err.Error(), "commit")
}

func TestParse_refusesUnknownFields(t *testing.T) {
	t.Parallel()
	_, err := Parse([]byte(strings.Replace(validLedger, `"chain_id"`, `"chain_idd"`, 1)))
	require.Error(t, err, "a typo in a field name must not silently drop the field")
}

func TestRender_pendingCellsNeverShowNull(t *testing.T) {
	t.Parallel()
	table := parseValid(t).RenderTable()
	assert.NotContains(t, table, "null")
	assert.NotContains(t, table, "<nil>")
	lines := strings.Split(strings.TrimSpace(table), "\n")
	require.Len(t, lines, 7, "BEGIN marker, header, separator, three rows, END marker")
	assert.Equal(t, BeginMarker, lines[0])
	assert.Equal(t, EndMarker, lines[len(lines)-1])
	genesis, upgrade := lines[3], lines[4]
	assert.Contains(t, genesis, "| genesis |")
	assert.Contains(t, genesis, "2026-09-12T15:00:00Z", "the genesis row shows the genesis time")
	assert.Contains(t, upgrade, "| 36300 |")
	assert.Contains(t, upgrade, "*(pending)*", "a missing digest renders as pending")
	assert.Contains(t, upgrade, "[#0](https://gno.land/r/gov/dao:0)")
	assert.Contains(t, upgrade, "*(not set)*", "a null halt_min_version says so")
	assert.Contains(t, lines[5], "`v1.4.0-rc.1`")
	assert.Contains(t, lines[5], "*(pending)*", "a null halt_time renders as pending")
}

func TestSplice_replacesOnlyTheGeneratedBlock(t *testing.T) {
	t.Parallel()
	doc := "# title\n\nprose before\n\n" + BeginMarker + "\n| old |\n" + EndMarker + "\n\nprose after\n"
	out, err := Splice([]byte(doc), "TABLE")
	require.NoError(t, err)
	assert.Equal(t, "# title\n\nprose before\n\nTABLE\n\nprose after\n", string(out))

	_, err = Splice([]byte("no markers here\n"), "TABLE")
	require.Error(t, err)
}

func TestCheck_detectsAStaleDocument(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, LedgerFile), []byte(validLedger), 0o644))
	doc := "notes\n\n" + BeginMarker + "\n| stale |\n" + EndMarker + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, DocFile), []byte(doc), 0o644))

	err := Check(dir)
	require.Error(t, err, "a hand-edited table must fail the check")
	assert.Contains(t, err.Error(), "stale")

	require.NoError(t, Render(dir))
	require.NoError(t, Check(dir), "a freshly rendered document passes")

	before, _ := os.ReadFile(filepath.Join(dir, DocFile))
	require.NoError(t, Render(dir))
	after, _ := os.ReadFile(filepath.Join(dir, DocFile))
	assert.Equal(t, string(before), string(after), "rendering is idempotent")
}

func TestHasVersion_stripsThePreReleaseSuffix(t *testing.T) {
	t.Parallel()
	l := parseValid(t)
	assert.True(t, l.Has("v1.3.0"))
	assert.True(t, l.Has("v1.3.0-rc.2"), "an rc rehearses the final version's entry")
	assert.True(t, l.Has("v1.4.0-rc.1"))
	assert.False(t, l.Has("v1.9.0"))
}

func TestBlockRanges_followTheContract(t *testing.T) {
	t.Parallel()
	l := parseValid(t)
	ranges := l.BlockRanges()
	require.Len(t, ranges, 3)
	assert.Equal(t, [2]int64{1, 36300}, [2]int64{ranges[0].From, ranges[0].To}, "genesis runs from block 1 to the first halt")
	assert.Equal(t, [2]int64{36301, 113000}, [2]int64{ranges[1].From, ranges[1].To})
	assert.Equal(t, int64(113001), ranges[2].From)
	assert.Equal(t, int64(0), ranges[2].To, "the current version has no upper bound yet")
}

// Every ledger committed under misc/deployments must validate and its rendered
// table must be current: this is the CI check.
func TestRepositoryLedgers(t *testing.T) {
	t.Parallel()
	dirs, err := filepath.Glob(filepath.Join("..", "..", "..", "misc", "deployments", "*", LedgerFile))
	require.NoError(t, err)
	require.NotEmpty(t, dirs, "no ledger found under misc/deployments; the mainnet one has moved?")
	for _, ledger := range dirs {
		dir := filepath.Dir(ledger)
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, Check(dir))
		})
	}
}

// The JSON Schema shipped for third-party readers must describe the same
// fields this package reads, or a consumer validating against it is misled.
func TestSchemaDescribesTheStructs(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "misc", "deployments", "upgrades", "upgrades.schema.json"))
	require.NoError(t, err)
	var schema struct {
		Properties map[string]any `json:"properties"`
		Defs       struct {
			Entry struct {
				Properties map[string]any `json:"properties"`
			} `json:"entry"`
		} `json:"$defs"`
	}
	require.NoError(t, json.Unmarshal(raw, &schema))

	assert.ElementsMatch(t, jsonTags(reflect.TypeFor[Ledger]()), keys(schema.Properties), "root properties")
	assert.ElementsMatch(t, jsonTags(reflect.TypeFor[Entry]()), keys(schema.Defs.Entry.Properties), "entry properties")
}

func jsonTags(typ reflect.Type) []string {
	var tags []string
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		tags = append(tags, strings.Split(tag, ",")[0])
	}
	return tags
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
