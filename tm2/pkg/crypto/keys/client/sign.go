package client

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errInvalidTxFile = errors.New("invalid transaction file")

type signOpts struct {
	chainID         string
	accountSequence uint64
	accountNumber   uint64
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
	OutputDocument string
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

	fs.StringVar(
		&c.OutputDocument,
		"output-document",
		"",
		"the signature json document to save. If empty, outputs the signature in the terminal",
	)
}

func execSign(cfg *SignCfg, args []string, io commands.IO) error {
	// Make sure the key name is provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	if cfg.RootCfg.Json {
		io.ErrPrintln("warning: -json flag has no effect on `sign` command")
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
	}

	kOpts := keyOpts{
		keyName:     args[0],
		decryptPass: password,
	}

	// Generate the signature
	signature, err := generateSignature(&tx, kb, sOpts, kOpts)
	if err != nil {
		return fmt.Errorf("unable to sign transaction, %w", err)
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
	signBytes, err := tx.GetSignBytes(
		signOpts.chainID,
		signOpts.accountNumber,
		signOpts.accountSequence,
	)
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
		return nil, fmt.Errorf("unable to sign transaction bytes, %w", err)
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
			PubKey:    sig.PubKey,
			Signature: sig.Signature,
		}

		return nil
	}

	// Append the signature, since it wasn't
	// present before
	tx.Signatures = append(
		tx.Signatures, std.Signature{
			PubKey:    sig.PubKey,
			Signature: sig.Signature,
		},
	)

	// Validate the tx after signing
	if err := tx.ValidateBasic(); err != nil {
		return fmt.Errorf("unable to validate transaction, %w", err)
	}

	return nil
}
