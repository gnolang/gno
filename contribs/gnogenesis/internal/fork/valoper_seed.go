package fork

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// monikerRe mirrors r/gnops/valopers/valopers.gno's validateMonikerRe
// (^[a-zA-Z0-9][\w -]{0,30}[a-zA-Z0-9]$ — alphanumeric start/end, with
// alphanumerics/spaces/hyphens/underscores in between, total length
// 2..32). The realm panics on regex failure during chain replay; this
// pre-flight catches misformatted monikers at CSV-validation time so
// the .jsonl is never emitted with rows that would explode at boot.
var monikerRe = regexp.MustCompile(`^[a-zA-Z0-9][\w -]{0,30}[a-zA-Z0-9]$`)

// validServerTypes mirrors r/gnops/valopers ServerType*. Kept in sync
// with the realm; if the realm grows a new variant, add it here.
var validServerTypes = map[string]struct{}{
	"cloud":       {},
	"on-prem":     {},
	"data-center": {},
}

const (
	valopersPkgPath = "gno.land/r/gnops/valopers"
	registerFunc    = "Register"

	// gas budget per Register tx — enough headroom for the realm's
	// signingRegistry insert + cross-call into v3 NotifyValoperChanged.
	// Tuned against the txtar e2e tests (60M was sufficient there).
	defaultRegisterGasWanted = 60_000_000
)

type valoperSeedCfg struct {
	csvPath string
	output  string
}

type seedRow struct {
	OperatorAddr  string
	SigningPubKey string
	Moniker       string
	Description   string
	ServerType    string
}

// newValoperSeedCmd registers `gnogenesis fork valoper-seed`. It
// validates a CSV of valoper-registration rows and emits a deterministic
// .jsonl of gnoland.TxWithMetadata entries. Each line is a MsgCall to
// valopers.Register; gnogenesis fork generate consumes the .jsonl via
// --migration-tx after historical replay.
//
// Validation is fail-fast: any bad row aborts before any output is
// written, so the .jsonl is always either complete or absent.
func newValoperSeedCmd(io commands.IO) *commands.Command {
	cfg := &valoperSeedCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "valoper-seed",
			ShortUsage: "valoper-seed [flags]",
			ShortHelp:  "build a valoper migration .jsonl from a CSV",
			LongHelp: `Build a deterministic .jsonl of valopers.Register migration txs from
a CSV of (operator_addr, signing_pubkey, moniker, description, server_type).

The output is intended for gnogenesis fork generate's --migration-tx flag,
which appends migration txs after historical replay. Each emitted tx is a
genesis-mode MsgCall to gno.land/r/gnops/valopers.Register, with Caller
set to operator_addr (so OriginCaller during genesis-mode replay equals
the operator address — the realm's post-genesis squat guard is bypassed
via ChainHeight()==0 during migration replay).

CSV schema (header row required, exact column names):

  operator_addr,signing_pubkey,moniker,description,server_type

Validations (fail-fast, no partial output):
  - operator_addr is a valid bech32 g1 address
  - signing_pubkey is a valid bech32 gpub1 that decodes to a non-nil PubKey
  - moniker non-empty (1..32 chars; the realm's regex is the source of truth)
  - description non-empty
  - server_type ∈ {cloud, on-prem, data-center}
  - no duplicate operator_addr
  - no duplicate signing_pubkey

Output is sorted by operator_addr so the same CSV produces a byte-equal
.jsonl across runs. Idempotent: re-running with the same CSV is safe.

PREREQUISITE: each Register call cross-calls
gno.land/r/sys/validators/v3.NotifyValoperChanged, and gnoland's
InitChainer auto-runs v3.AssertGenesisValopersConsistent at end of
genesis-mode replay (when PastChainIDs is set). v3 must therefore
already be deployed at genesis. If the source chain (the one being
forked from) does not have v3 deployed in its genesis-mode addpkg
txs, use 'gnogenesis fork addpkg' to produce a separate .jsonl
that deploys v3 (and any other new realms valopers transitively
imports), and pass it BEFORE this seed via repeated --migration-tx
flags. Order matters — addpkg first, then this seed:

  gnogenesis fork addpkg --output addpkg-v3.jsonl examples/gno.land/r/sys/validators/v3
  gnogenesis fork valoper-seed --csv valopers.csv --output valoper-seed.jsonl
  gnogenesis fork generate \
      --source ... \
      --migration-tx addpkg-v3.jsonl \
      --migration-tx valoper-seed.jsonl \
      --output genesis.json

If the source chain already has v3 deployed (e.g., a fresh launch or
a fork from a chain where v3 was already live), pass --patch-realm
on the existing addpkg and skip the addpkg step.

Example (full flow, source is gnoland-1 with v3 NOT pre-deployed):

  gnogenesis fork addpkg --output addpkg-v3.jsonl examples/gno.land/r/sys/validators/v3
  gnogenesis fork valoper-seed --csv valopers.csv --output valoper-seed.jsonl
  gnogenesis fork generate --source ... \
      --migration-tx addpkg-v3.jsonl --migration-tx valoper-seed.jsonl \
      --patch-realm gno.land/r/gnops/valopers=examples/gno.land/r/gnops/valopers \
      --patch-realm gno.land/r/gnops/valopers/proposal=examples/gno.land/r/gnops/valopers/proposal \
      --output genesis.json`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execValoperSeed(ctx, cfg, io)
		},
	)
}

func (c *valoperSeedCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.csvPath, "csv", "", "input CSV path (required)")
	fs.StringVar(&c.output, "output", "", "output .jsonl path (required)")
}

func execValoperSeed(_ context.Context, cfg *valoperSeedCfg, io commands.IO) error {
	if cfg.csvPath == "" {
		return errors.New("--csv is required")
	}
	if cfg.output == "" {
		return errors.New("--output is required")
	}

	rows, err := loadAndValidateCSV(cfg.csvPath)
	if err != nil {
		return err
	}

	// Sort by operator address so the output is deterministic across
	// runs regardless of CSV row order.
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].OperatorAddr < rows[j].OperatorAddr
	})

	var buf strings.Builder
	for _, r := range rows {
		tx := buildRegisterTx(r)
		line, err := amino.MarshalJSON(tx)
		if err != nil {
			return fmt.Errorf("marshal tx for operator %s: %w", r.OperatorAddr, err)
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}

	// AssertGenesisValopersConsistent runs unconditionally in
	// gnoland's InitChainer for hardfork-mode boots, so the .jsonl
	// itself need not include it.

	if err := os.WriteFile(cfg.output, []byte(buf.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", cfg.output, err)
	}

	io.Printfln("wrote %d valoper Register txs to %s", len(rows), cfg.output)
	return nil
}

// loadAndValidateCSV reads the CSV file, parses rows, and validates
// every row. On any error returns (nil, err) without writing output.
func loadAndValidateCSV(path string) ([]seedRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = 5

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	if err := validateHeader(header); err != nil {
		return nil, err
	}

	var (
		rows        []seedRow
		seenOps     = map[string]int{} // operator -> CSV row index
		seenPubKeys = map[string]int{} // pubkey -> CSV row index
	)
	for i := 0; ; i++ {
		rec, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read row %d: %w", i+2, err) // +2: 1-indexed + header
		}
		row := seedRow{
			OperatorAddr:  strings.TrimSpace(rec[0]),
			SigningPubKey: strings.TrimSpace(rec[1]),
			Moniker:       strings.TrimSpace(rec[2]),
			Description:   strings.TrimSpace(rec[3]),
			ServerType:    strings.TrimSpace(rec[4]),
		}
		if err := validateRow(&row, i+2); err != nil {
			return nil, err
		}
		if prev, dup := seenOps[row.OperatorAddr]; dup {
			return nil, fmt.Errorf("duplicate operator_addr %s (rows %d and %d)", row.OperatorAddr, prev, i+2)
		}
		if prev, dup := seenPubKeys[row.SigningPubKey]; dup {
			return nil, fmt.Errorf("duplicate signing_pubkey %s (rows %d and %d)", row.SigningPubKey, prev, i+2)
		}
		seenOps[row.OperatorAddr] = i + 2
		seenPubKeys[row.SigningPubKey] = i + 2
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil, errors.New("csv has no data rows")
	}

	return rows, nil
}

// validateHeader ensures the CSV columns match the documented schema
// exactly. Off-by-one column shifts otherwise produce silently-bad
// .jsonl that fails opaquely at chain replay.
func validateHeader(header []string) error {
	want := []string{"operator_addr", "signing_pubkey", "moniker", "description", "server_type"}
	if len(header) != len(want) {
		return fmt.Errorf("header has %d columns, want %d (%s)", len(header), len(want), strings.Join(want, ","))
	}
	for i, c := range header {
		if strings.TrimSpace(c) != want[i] {
			return fmt.Errorf("header column %d is %q, want %q", i+1, c, want[i])
		}
	}
	return nil
}

// validateRow checks all per-row invariants and **canonicalizes** the
// operator_addr and signing_pubkey strings on the row in-place.
// Canonicalization defeats case-aliasing dedup bypass: bech32 accepts
// both lowercase and uppercase encodings of the same payload, so two
// rows with the same canonical operator but different cases would
// otherwise pass the seenOps dedup check and produce duplicate Valoper
// profiles for the same canonical operator.
func validateRow(row *seedRow, csvRow int) error {
	if row.Moniker == "" {
		return fmt.Errorf("row %d: moniker is empty", csvRow)
	}
	if len(row.Moniker) > 32 {
		return fmt.Errorf("row %d: moniker %q exceeds 32 characters", csvRow, row.Moniker)
	}
	if !monikerRe.MatchString(row.Moniker) {
		return fmt.Errorf("row %d: moniker %q must match the realm regex (2..32 chars, alphanumeric start/end, alphanumeric/space/hyphen/underscore middle)", csvRow, row.Moniker)
	}
	if row.Description == "" {
		return fmt.Errorf("row %d: description is empty", csvRow)
	}
	if _, ok := validServerTypes[row.ServerType]; !ok {
		return fmt.Errorf("row %d: server_type %q not in {cloud, on-prem, data-center}", csvRow, row.ServerType)
	}

	addr, err := crypto.AddressFromBech32(row.OperatorAddr)
	if err != nil {
		return fmt.Errorf("row %d: invalid operator_addr %q: %w", csvRow, row.OperatorAddr, err)
	}
	row.OperatorAddr = addr.String() // canonicalize (lowercase bech32)

	pk, err := crypto.PubKeyFromBech32(row.SigningPubKey)
	if err != nil {
		return fmt.Errorf("row %d: invalid signing_pubkey %q: %w", csvRow, row.SigningPubKey, err)
	}
	if pk == nil {
		return fmt.Errorf("row %d: signing_pubkey %q decoded to nil PubKey", csvRow, row.SigningPubKey)
	}
	// Canonicalize the pubkey by re-encoding from the parsed PubKey.
	// PubKeyToBech32 always emits lowercase, matching what the realm
	// stores at write time.
	row.SigningPubKey = crypto.PubKeyToBech32(pk)

	// Reject operator_addr == derive(signing_pubkey). Collapsing the
	// two identities into one address makes signing-key compromise
	// equivalent to operator-slot compromise: anyone holding the
	// validator's private key could call valopers entrypoints as the
	// operator. The whole point of separating the operator profile
	// from the consensus signing key is that their security domains
	// stay distinct. Catch the misconfiguration here (cheapest layer)
	// before it ships into a hardfork ceremony's migration .jsonl.
	if addr == pk.Address() {
		return fmt.Errorf("row %d: operator_addr %s equals the address derived from signing_pubkey — operator identity must be distinct from the consensus signing key", csvRow, row.OperatorAddr)
	}

	return nil
}

// buildRegisterTx produces a TxWithMetadata for a single valoper.Register
// MsgCall. Caller is the operator address; OriginCaller during genesis-
// mode replay therefore equals the operator address, satisfying the
// realm's squat guard via the ChainHeight()==0 bypass.
func buildRegisterTx(row seedRow) gnoland.TxWithMetadata {
	caller, _ := crypto.AddressFromBech32(row.OperatorAddr)

	msg := vm.MsgCall{
		Caller:  caller,
		PkgPath: valopersPkgPath,
		Func:    registerFunc,
		Args: []string{
			row.Moniker,
			row.Description,
			row.ServerType,
			row.OperatorAddr,
			row.SigningPubKey,
		},
	}

	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(defaultRegisterGasWanted, std.NewCoin("ugnot", 0)),
		Signatures: []std.Signature{},
	}

	return gnoland.TxWithMetadata{
		Tx: tx,
		Metadata: &gnoland.GnoTxMetadata{
			BlockHeight: 0, // genesis-mode replay
		},
	}
}
