package fork

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validPubKey is a deterministic ed25519 pubkey usable across cases.
// Re-used from existing v3 test fixtures so it's a known-good string.
const (
	validPubKeyA = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq3ds6sdvc0shfkq02h6xx5g0jp04aadexfnpsmgjxu72xz9y30aqfrlpny"
	validPubKeyB = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
	validPubKeyC = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqddddqg2glc8x4fl7vxjlnr7p5a3czm5kcdp4239sg6yqdc4rc2r5cjrffs"

	// Valid g1 addresses (any valid bech32, used as operator addrs).
	opAddrA = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	opAddrB = "g1c0j899h88nwyvnzvh5jagpq6fkkyuj76nld6t0"
	opAddrC = "g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h"
)

const validHeader = "operator_addr,signing_pubkey,moniker,description,server_type"

func writeCSV(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "valopers.csv")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func runSeed(t *testing.T, dir, csvContent string) (string, error) {
	t.Helper()
	csvPath := writeCSV(t, dir, csvContent)
	outPath := filepath.Join(dir, "out.jsonl")
	cfg := &valoperSeedCfg{csvPath: csvPath, output: outPath}
	io := commands.NewTestIO()
	if err := execValoperSeed(t.Context(), cfg, io); err != nil {
		return "", err
	}
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	return string(data), nil
}

func TestValoperSeed_HappyPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	csvContent := validHeader + "\n" +
		opAddrA + "," + validPubKeyA + ",alice-validator,Alice's validator,cloud\n" +
		opAddrB + "," + validPubKeyB + ",bob-validator,Bob's validator,on-prem\n"

	out, err := runSeed(t, dir, csvContent)
	require.NoError(t, err)

	// Two Register lines, no tail-line assertion (gnoland InitChainer
	// runs the assertion unconditionally in hardfork mode).
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	require.Len(t, lines, 2)

	// Output is sorted by operator addr; opAddrB < opAddrA lexically
	// because "g1c0j..." < "g1jg8...".
	require.True(t, opAddrB < opAddrA, "fixture ordering assumption")

	var first gnoland.TxWithMetadata
	require.NoError(t, amino.UnmarshalJSON([]byte(lines[0]), &first))
	require.Len(t, first.Tx.Msgs, 1)
	msg, ok := first.Tx.Msgs[0].(vm.MsgCall)
	require.True(t, ok, "first msg is MsgCall")
	assert.Equal(t, "gno.land/r/gnops/valopers", msg.PkgPath)
	assert.Equal(t, "Register", msg.Func)
	assert.Equal(t, opAddrB, msg.Caller.String())
	require.Len(t, msg.Args, 5)
	assert.Equal(t, "bob-validator", msg.Args[0])
	assert.Equal(t, opAddrB, msg.Args[3])
	assert.Equal(t, validPubKeyB, msg.Args[4])
	require.NotNil(t, first.Metadata)
	assert.Equal(t, int64(0), first.Metadata.BlockHeight)

	var second gnoland.TxWithMetadata
	require.NoError(t, amino.UnmarshalJSON([]byte(lines[1]), &second))
	msg2 := second.Tx.Msgs[0].(vm.MsgCall)
	assert.Equal(t, opAddrA, msg2.Caller.String())
}

func TestValoperSeed_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Same CSV in two different orderings — output must be byte-equal
	// because the tool sorts by operator address.
	csv1 := validHeader + "\n" +
		opAddrA + "," + validPubKeyA + ",alice,Alice,cloud\n" +
		opAddrB + "," + validPubKeyB + ",bob,Bob,on-prem\n" +
		opAddrC + "," + validPubKeyC + ",carol,Carol,data-center\n"
	csv2 := validHeader + "\n" +
		opAddrC + "," + validPubKeyC + ",carol,Carol,data-center\n" +
		opAddrA + "," + validPubKeyA + ",alice,Alice,cloud\n" +
		opAddrB + "," + validPubKeyB + ",bob,Bob,on-prem\n"

	dir1 := t.TempDir()
	out1, err := runSeed(t, dir1, csv1)
	require.NoError(t, err)
	dir2 := t.TempDir()
	out2, err := runSeed(t, dir2, csv2)
	require.NoError(t, err)
	_ = dir

	hash := func(s string) string {
		h := sha256.Sum256([]byte(s))
		return hex.EncodeToString(h[:])
	}
	assert.Equal(t, hash(out1), hash(out2), "different row orders must produce byte-equal output")
}

func TestValoperSeed_RejectsDuplicateOperator(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	csv := validHeader + "\n" +
		opAddrA + "," + validPubKeyA + ",alice,Alice,cloud\n" +
		opAddrA + "," + validPubKeyB + ",alice2,Alice2,on-prem\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate operator_addr")

	// No output file produced (fail-fast).
	_, statErr := os.Stat(filepath.Join(dir, "out.jsonl"))
	assert.True(t, os.IsNotExist(statErr), "no partial output on validation failure")
}

func TestValoperSeed_RejectsDuplicateSigningPubKey(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	csv := validHeader + "\n" +
		opAddrA + "," + validPubKeyA + ",alice,Alice,cloud\n" +
		opAddrB + "," + validPubKeyA + ",bob,Bob,on-prem\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate signing_pubkey")
}

func TestValoperSeed_RejectsBadPubKey(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	csv := validHeader + "\n" +
		opAddrA + "," + "gpub1notreallyapubkey" + ",alice,Alice,cloud\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signing_pubkey")
}

func TestValoperSeed_RejectsBadOperatorAddr(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	csv := validHeader + "\n" +
		"not-bech32" + "," + validPubKeyA + ",alice,Alice,cloud\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operator_addr")
}

func TestValoperSeed_RejectsBadServerType(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	csv := validHeader + "\n" +
		opAddrA + "," + validPubKeyA + ",alice,Alice,bare-metal\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server_type")
}

func TestValoperSeed_RejectsEmptyMoniker(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	csv := validHeader + "\n" +
		opAddrA + "," + validPubKeyA + ",,Alice,cloud\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "moniker is empty")
}

func TestValoperSeed_RejectsTooLongMoniker(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	long := strings.Repeat("a", 33)
	csv := validHeader + "\n" +
		opAddrA + "," + validPubKeyA + "," + long + ",Alice,cloud\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds 32 characters")
}

func TestValoperSeed_RejectsMissingHeader(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Wrong column order.
	csv := "signing_pubkey,operator_addr,moniker,description,server_type\n" +
		validPubKeyA + "," + opAddrA + ",alice,Alice,cloud\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "header column")
}

func TestValoperSeed_RejectsEmptyCSV(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	_, err := runSeed(t, dir, validHeader+"\n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no data rows")
}

func TestValoperSeed_RejectsEmptyDescription(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	csv := validHeader + "\n" +
		opAddrA + "," + validPubKeyA + ",alice,,cloud\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "description is empty")
}

func TestValoperSeed_RejectsBadMoniker(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		moniker string
	}{
		{"too-short", "a"},
		{"trailing-hyphen", "alice-"},
		{"leading-hyphen", "-alice"},
		{"special-char", "alice!"},
		// Note: leading/trailing whitespace passes through CSV
		// because validateRow trims. That's intentional ergonomics.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			csv := validHeader + "\n" +
				opAddrA + "," + validPubKeyA + "," + tc.moniker + ",Alice,cloud\n"
			_, err := runSeed(t, dir, csv)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "moniker")
		})
	}
}

func TestValoperSeed_RejectsOperatorEqualsDerivedSigningAddress(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Pair where chain.PubKeyAddress(pubkey) derives to opAddrA
	// (g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5) — same fixture used
	// in valopers.txtar's "by coincidence" register case. With the
	// new check, valoper-seed must refuse this row at CSV-validation
	// time so the misconfiguration never ships in a migration .jsonl.
	const pubKeyDerivingToOpA = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj"

	csv := validHeader + "\n" +
		opAddrA + "," + pubKeyDerivingToOpA + ",alice,Alice,cloud\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err, "operator_addr equal to derived signing address must be rejected")
	assert.Contains(t, err.Error(), "equals the address derived from signing_pubkey")
}

func TestValoperSeed_MonikerRegexDoesNotDrift(t *testing.T) {
	t.Parallel()

	// The realm derives its moniker regex from MonikerMaxLength
	// (`^[a-zA-Z0-9][\w -]{0,MonikerMaxLength-2}[a-zA-Z0-9]$`).
	// gnogenesis hardcodes the middle-bound integer for performance
	// and to avoid pulling Gno into Go. If MonikerMaxLength ever
	// changes on the realm side without the gnogenesis hardcode
	// being updated, the pre-flight would silently accept inputs
	// the realm rejects (or vice versa), producing migration .jsonls
	// that explode at chain replay. This test pins the two together.
	//
	// Source of truth: examples/gno.land/r/gnops/valopers/valopers.gno's
	// `MonikerMaxLength` constant. Read with regex (Go can't import
	// Gno) and compare to the bound encoded in monikerRe.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// Walk up from contribs/gnogenesis/internal/fork to repo root
	// (4 levels: fork -> internal -> gnogenesis -> contribs -> gno).
	root := filepath.Join(wd, "..", "..", "..", "..")
	gnoPath := filepath.Join(root, "examples", "gno.land", "r", "gnops", "valopers", "valopers.gno")

	data, err := os.ReadFile(gnoPath)
	if err != nil {
		t.Fatalf("read %s: %v", gnoPath, err)
	}

	re := regexp.MustCompile(`(?m)^\s*MonikerMaxLength\s*=\s*(\d+)`)
	m := re.FindSubmatch(data)
	require.Len(t, m, 2, "could not parse MonikerMaxLength from %s", gnoPath)
	gnoMax, err := strconv.Atoi(string(m[1]))
	require.NoError(t, err)

	// Realm regex middle bound = MonikerMaxLength - 2 (subtracts the
	// leading + trailing alphanumeric chars).
	wantMiddle := gnoMax - 2

	// Extract the middle bound from monikerRe by parsing the integer
	// inside `{0,N}`.
	reBound := regexp.MustCompile(`\{0,(\d+)\}`)
	gotBound := reBound.FindStringSubmatch(monikerRe.String())
	require.Len(t, gotBound, 2, "could not parse middle bound from monikerRe %q", monikerRe.String())
	gotMiddle, err := strconv.Atoi(gotBound[1])
	require.NoError(t, err)

	assert.Equal(t, wantMiddle, gotMiddle,
		"moniker regex drift: realm MonikerMaxLength=%d implies middle-bound=%d, gnogenesis monikerRe encodes middle-bound=%d (update the hardcode in valoper_seed.go to match)",
		gnoMax, wantMiddle, gotMiddle)
}

func TestValoperSeed_DedupsCaseAliasedAddress(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Two rows with the same canonical operator address but different
	// cases — bech32 accepts both. Without normalization, the dedup
	// check passes both rows and produces duplicate Valoper profiles
	// for the same canonical operator. With normalization, the second
	// row is rejected as a duplicate.
	upper := strings.ToUpper(opAddrA)
	csv := validHeader + "\n" +
		opAddrA + "," + validPubKeyA + ",alice,Alice,cloud\n" +
		upper + "," + validPubKeyB + ",alice2,Alice2,on-prem\n"

	_, err := runSeed(t, dir, csv)
	require.Error(t, err, "case-aliased operator must trip the duplicate check")
	assert.Contains(t, err.Error(), "duplicate operator_addr")
}
