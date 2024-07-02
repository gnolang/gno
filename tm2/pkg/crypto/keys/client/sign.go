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

	TxPath        string
	ChainID       string
	AccountNumber uint64
	Sequence      uint64
	NameOrBech32  string
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
}

func execSign(cfg *SignCfg, args []string, io commands.IO) error {
	// Make sure the key name is provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	// saveTx saves the given transaction to the given path (Amino-encoded JSON)
	saveTx := func(tx *std.Tx, path string) error {
		// Encode the transaction
		encodedTx, err := amino.MarshalJSON(tx)
		if err != nil {
			return fmt.Errorf("unable ot marshal tx to JSON, %w", err)
		}

		// Save the transaction
		if err := os.WriteFile(path, encodedTx, 0o644); err != nil {
			return fmt.Errorf("unable to write tx to %s, %w", path, err)
		}

		io.Printf("\nTx successfully signed and saved to %s\n", path)

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

	// Check if accountNumber or sequence is zero to determine if account information needs to be fetched
	if cfg.AccountNumber == 0 || cfg.Sequence == 0 {
		accountAddr := info.GetAddress()

		qopts := &QueryCfg{
			RootCfg: cfg.RootCfg,
			Path:    fmt.Sprintf("auth/accounts/%s", accountAddr),
		}
		qres, err := QueryHandler(qopts)
		if err == nil {
			var qret struct{ BaseAccount std.BaseAccount }
			if err := amino.UnmarshalJSON(qres.Response.Data, &qret); err == nil {
				cfg.AccountNumber = qret.BaseAccount.AccountNumber
				cfg.Sequence = qret.BaseAccount.Sequence
			}
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

	// Sign the transaction
	if err := signTx(&tx, kb, sOpts, kOpts); err != nil {
		return fmt.Errorf("unable to sign transaction, %w", err)
	}

	return saveTx(&tx, cfg.TxPath)
}

// signTx generates the transaction signature,
// and saves it to the given transaction
func signTx(
	tx *std.Tx,
	kb keys.Keybase,
	signOpts signOpts,
	keyOpts keyOpts,
) error {
	// Save the signature
	signers := tx.GetSigners()
	if tx.Signatures == nil {
		for range signers {
			tx.Signatures = append(tx.Signatures, std.Signature{
				PubKey:    nil, // Zero signature
				Signature: nil, // Zero signature
			})
		}
	}

	// Validate the tx after signing
	if err := tx.ValidateBasic(); err != nil {
		return fmt.Errorf("unable to validate transaction, %w", err)
	}

	signBytes, err := tx.GetSignBytes(
		signOpts.chainID,
		signOpts.accountNumber,
		signOpts.accountSequence,
	)
	if err != nil {
		return fmt.Errorf("unable to get signature bytes, %w", err)
	}

	// Sign the transaction data
	sig, pub, err := kb.Sign(
		keyOpts.keyName,
		keyOpts.decryptPass,
		signBytes,
	)
	if err != nil {
		return fmt.Errorf("unable to sign transaction bytes, %w", err)
	}

	addr := pub.Address()

	found := false
	for i := range tx.Signatures {
		if signers[i] == addr {
			found = true
			tx.Signatures[i] = std.Signature{
				PubKey:    pub,
				Signature: sig,
			}
		}
	}

	if !found {
		return fmt.Errorf("address %v (%s) not in signer set", addr, keyOpts.keyName)
	}

	return nil
}
