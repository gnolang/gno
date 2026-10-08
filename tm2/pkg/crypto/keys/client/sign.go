package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/ledger"
	tm2errors "github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errInvalidTxFile = tm2errors.New("invalid transaction file")

type signOpts struct {
	chainID         string
	accountSequence uint64
	accountNumber   uint64
	signMode        string
}

// These are the valid values of signOpts.signMode, and of the -sign-mode flag
// that sets it.
const (
	// SignModeFull signs the whole transaction as rendered. The device displays
	// every field, and refuses the transaction if it does not fit.
	SignModeFull = "full"
	// SignModeDigest signs the rendering in which strings over
	// std.MaxSignDocStringLen are replaced by their digests, which is how a
	// transaction too large for a Ledger gets signed at all. The signer gives
	// up reading those fields on the device and compares digests instead.
	SignModeDigest = "digest"
)

// signBytes renders tx for signing in the mode opts asks for. An unset mode is
// full, so a caller that builds a config itself rather than parsing flags
// keeps the behaviour it had before the mode existed.
func (opts signOpts) signBytes(tx *std.Tx) ([]byte, error) {
	if opts.signMode == SignModeDigest {
		return tx.GetSignBytesDigest(opts.chainID, opts.accountNumber, opts.accountSequence)
	}

	return tx.GetSignBytes(opts.chainID, opts.accountNumber, opts.accountSequence)
}

// explainLedgerSignError adds the remedy gnokey can offer to a device refusal
// that means "this payload is too large to display".
//
// The device layer classifies the failure and deliberately names no flag: it
// cannot import this package, and gnoclient reaches the same code with no flags
// to offer. So the remedy is phrased here, where -sign-mode exists, and only
// when the signer is not already using it.
func explainLedgerSignError(err error, mode string) error {
	if mode == SignModeDigest {
		// Already the smallest rendering gnokey can produce. Sending them
		// round the same loop would be worse than saying nothing.
		return err
	}

	switch {
	case errors.Is(err, ledger.ErrPayloadTooLarge),
		errors.Is(err, ledger.ErrLargePayloadRefused):
		return fmt.Errorf("%w\nretry with -sign-mode digest, which replaces oversized "+
			"fields by hashes the device can display, or sign with a session key instead", err)
	case errors.Is(err, ledger.ErrTooManyJSONValues):
		return fmt.Errorf("%w\nretry with -sign-mode digest, or split the transaction "+
			"into smaller ones", err)
	default:
		return err
	}
}

// validateSignMode rejects an unknown mode rather than silently signing in
// full mode, which would leave a signer who misspelled the flag facing the
// device error the flag exists to avoid. The empty string is the unset mode,
// not a misspelling, and means full.
func validateSignMode(mode string) error {
	switch mode {
	case "", SignModeFull, SignModeDigest:
		return nil
	default:
		return fmt.Errorf("invalid sign mode %q, want %q or %q", mode, SignModeFull, SignModeDigest)
	}
}

// reportDigestedFields prints what signing in digest mode hides: the fields the
// device will show as a digest instead of as their contents.
//
// This is the only place a signer learns which digest to expect on the screen,
// and the comparison is the whole security of the mode, so it prints before the
// device prompts. It is a no-op in any other mode, so a command adopting the
// flag cannot forget to warn; the test cannot move into generateSignature,
// which maketx calls a second time for the simulated transaction and would
// print the table twice.
func reportDigestedFields(tx *std.Tx, opts signOpts, io commands.IO) error {
	if opts.signMode != SignModeDigest {
		return nil
	}

	doc := tx.SignDoc(opts.chainID, opts.accountNumber, opts.accountSequence)

	_, fields, err := std.GetSignaturePayloadDigest(doc)
	if err != nil {
		return fmt.Errorf("unable to compute digested fields, %w", err)
	}

	if len(fields) == 0 {
		io.Printfln("\nsign mode digest: nothing to digest, the device will show every field")

		return nil
	}

	io.Printfln("\nsign mode digest: %d field(s) will reach the device as a digest only.", len(fields))
	io.Printfln("Compare each one against the device screen before approving.\n")
	printDigestedFields(io, fields)

	return nil
}

// printDigestedFields lists fields the way a signer has to read them: one line
// naming the field, one holding the digest.
//
// Shared with gnokey verify, which prints the same list for a signature already
// made, so that the two cannot drift into describing the same transaction
// differently. The docs promise a signer can compare one against the other.
func printDigestedFields(io commands.IO, fields []std.Digested) {
	for _, field := range fields {
		io.Printfln("  %s (%d bytes)", field.Path, field.Length)
		io.Printfln("    %s", groupHex(field.Sum))
	}
}

// groupHex breaks a digest into eight-character groups, so that comparing it
// against a device screen character by character is possible at all.
func groupHex(sum string) string {
	var out strings.Builder
	for i := 0; i < len(sum); i += 8 {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(sum[i:min(i+8, len(sum))])
	}

	return out.String()
}

type keyOpts struct {
	keyName     string
	decryptPass string
}

type SignCfg struct {
	RootCfg *BaseCfg

	TxPath         string
	ChainID        string
	AccountNumber  uint64
	Sequence       uint64
	NameOrBech32   string
	Session        bool
	OutputDocument string
	SignMode       string
}

func NewSignCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &SignCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "sign",
			ShortUsage: "sign [flags] <key-name or address>",
			ShortHelp:  "signs the given tx document and saves it to disk",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSign(cfg, args, io)
		},
	)
}

func (c *SignCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.TxPath,
		"tx-path",
		"",
		"path to the Amino JSON-encoded tx (file) to sign",
	)

	fs.StringVar(
		&c.ChainID,
		"chainid",
		"dev",
		"the ID of the chain",
	)

	fs.Uint64Var(
		&c.AccountNumber,
		"account-number",
		0,
		"account number to sign with",
	)

	fs.Uint64Var(
		&c.Sequence,
		"account-sequence",
		0,
		"account sequence to sign with",
	)

	fs.BoolVar(
		&c.Session,
		"session",
		false,
		"the key is a session account",
	)

	fs.StringVar(
		&c.OutputDocument,
		"output-document",
		"",
		"the signature json document to save. If empty, updates the tx (file) being signed",
	)

	fs.StringVar(
		&c.SignMode,
		"sign-mode",
		SignModeFull,
		"how the tx is rendered for signing: full, or digest to replace oversized "+
			"strings by their hashes so a Ledger can parse the payload",
	)
}

func execSign(cfg *SignCfg, args []string, io commands.IO) error {
	// Make sure the key name is provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	if err := validateSignMode(cfg.SignMode); err != nil {
		return err
	}

	// saveSignature saves the given transaction signature to the given path (Amino-encoded JSON)
	saveSignature := func(signature *std.Signature, path string) error {
		// Encode the signature
		encodedSig, err := amino.MarshalJSON(signature)
		if err != nil {
			return fmt.Errorf("unable to marshal signature to JSON, %w", err)
		}

		// Save the signature
		if err := os.WriteFile(path, encodedSig, 0o644); err != nil {
			return fmt.Errorf("unable to write signature to %s, %w", path, err)
		}

		io.Printf("\nSignature generated and successfully saved to %s\n", path)

		return nil
	}

	// Load the keybase
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return fmt.Errorf("unable to load keybase, %w", err)
	}

	// Fetch the key info from the keybase
	info, err := kb.GetByNameOrAddress(args[0])
	if err != nil {
		return fmt.Errorf("unable to get key from keybase, %w", err)
	}

	// Get the transaction bytes
	txRaw, err := os.ReadFile(cfg.TxPath)
	if err != nil {
		return fmt.Errorf("unable to read transaction file")
	}

	// Make sure there is something to actually sign
	if len(txRaw) == 0 {
		return errInvalidTxFile
	}

	// Make sure the tx is valid Amino JSON
	var tx std.Tx
	if err := amino.UnmarshalJSON(txRaw, &tx); err != nil {
		return fmt.Errorf("unable to unmarshal transaction, %w", err)
	}

	var password string

	// Check if we need to get a decryption password.
	// This is only required for local keys
	if info.GetType() != keys.TypeLedger {
		// Get the keybase decryption password
		prompt := "Enter password to decrypt key"
		if cfg.RootCfg.Quiet {
			prompt = "" // No prompt
		}

		password, err = io.GetPassword(
			prompt,
			cfg.RootCfg.InsecurePasswordStdin,
		)
		if err != nil {
			return fmt.Errorf("unable to get decryption key, %w", err)
		}
	}

	// Prepare the signature ops
	sOpts := signOpts{
		chainID:         cfg.ChainID,
		accountSequence: cfg.Sequence,
		accountNumber:   cfg.AccountNumber,
		signMode:        cfg.SignMode,
	}

	kOpts := keyOpts{
		keyName:     args[0],
		decryptPass: password,
	}

	// Tell the signer what the device will not be able to show them, before
	// the device asks them to approve it.
	if err := reportDigestedFields(&tx, sOpts, io); err != nil {
		return err
	}

	// Generate the signature
	signature, err := generateSignature(&tx, kb, sOpts, kOpts)
	if err != nil {
		return fmt.Errorf("unable to sign transaction, %w", err)
	}

	if cfg.Session {
		signature.SessionAddr = info.GetAddress()
	}

	if cfg.OutputDocument != "" {
		// Don't save the signature in-place, separate it
		return saveSignature(signature, cfg.OutputDocument)
	}

	// Add the signature to the tx
	if err = addSignature(&tx, signature); err != nil {
		return fmt.Errorf("unable to add signature: %w", err)
	}

	// Save the tx to disk
	if err = saveTx(&tx, cfg.TxPath); err != nil {
		return fmt.Errorf("unable to save tx: %w", err)
	}

	io.Printf("\nTx successfully signed and saved to %s\n", cfg.TxPath)

	return nil
}

// generateSignature generates the transaction signature
func generateSignature(
	tx *std.Tx,
	kb keys.Keybase,
	signOpts signOpts,
	keyOpts keyOpts,
) (*std.Signature, error) {
	signBytes, err := signOpts.signBytes(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to get signature bytes, %w", err)
	}

	// Sign the transaction data
	sig, pub, err := kb.Sign(
		keyOpts.keyName,
		keyOpts.decryptPass,
		signBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to sign transaction bytes, %w",
			explainLedgerSignError(err, signOpts.signMode))
	}

	return &std.Signature{
		PubKey:    pub,
		Signature: sig,
	}, nil
}

// addSignature generates the transaction signature,
// and saves it to the given transaction
func addSignature(tx *std.Tx, sig *std.Signature) error {
	// Save the signature
	if tx.Signatures == nil {
		tx.Signatures = make([]std.Signature, 0, 1)
	}

	// Check if the signature needs to be overwritten
	for index, signature := range tx.Signatures {
		if !signature.PubKey.Equals(sig.PubKey) {
			continue
		}

		// Save the signature
		tx.Signatures[index] = std.Signature{
			PubKey:      sig.PubKey,
			Signature:   sig.Signature,
			SessionAddr: sig.SessionAddr,
		}

		return nil
	}

	// Append the signature, since it wasn't
	// present before
	tx.Signatures = append(
		tx.Signatures, std.Signature{
			PubKey:      sig.PubKey,
			Signature:   sig.Signature,
			SessionAddr: sig.SessionAddr,
		},
	)

	// Validate the tx after signing
	if err := tx.ValidateBasic(); err != nil {
		return fmt.Errorf("unable to validate transaction, %w", err)
	}

	return nil
}
