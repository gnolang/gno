package client

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type VerifyCfg struct {
	RootCfg *BaseCfg

	DocPath   string
	SigPath   string
	Signature string

	ChainID         string
	AccountNumber   uint64
	AccountSequence uint64
	Offline         bool
}

func NewVerifyCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &VerifyCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "verify",
			ShortUsage: "verify [flags] <key-name or address>",
			ShortHelp:  "verifies the transaction signature",
			LongHelp:   "Verifies a std.Tx signature against <key-name or address> in your local keybase. The sign bytes are derived from the tx using --chain-id, --account-number, and --account-sequence; these must match the values used when the signature was created. If --account-number and --account-sequence are 0 and --offline=false, the command queries the chain (via --remote) to fill them; if the query fails, 0 is used. Provide the signature via --sigpath or --signature (mutually exclusive); otherwise the first signature in the tx (tx.Signatures[0]) is used. The tx is read from --docpath or stdin.",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execVerify(ctx, cfg, args, io)
		},
	)
}

func (c *VerifyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.DocPath,
		"docpath",
		"",
		"path of transaction file to verify in Amino JSON format (if empty, read from stdin)",
	)
	fs.StringVar(
		&c.SigPath,
		"sigpath",
		"",
		"path of signature file in Amino JSON format (mutually exclusive with -signature flag)",
	)
	fs.StringVar(
		&c.Signature,
		"signature",
		"",
		"base64-encoded signature string (mutually exclusive with -sigpath flag)",
	)
	fs.StringVar(
		&c.ChainID,
		"chain-id",
		"dev",
		"network chain ID used for signing",
	)
	fs.Uint64Var(
		&c.AccountNumber,
		"account-number",
		0,
		"account number of the signing account",
	)
	fs.Uint64Var(
		&c.AccountSequence,
		"account-sequence",
		0,
		"account sequence of the signing account",
	)
	fs.BoolVar(
		&c.Offline,
		"offline",
		false,
		"offline mode: do not query the chain for account number and account sequence",
	)
}

func execVerify(ctx context.Context, cfg *VerifyCfg, args []string, io commands.IO) error {
	var (
		kb  keys.Keybase
		err error
	)

	if len(args) != 1 {
		return flag.ErrHelp
	}

	// Fetch the key info from the keybase.
	kb, err = keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return err
	}

	info, err := kb.GetByNameOrAddress(args[0])
	if err != nil {
		return fmt.Errorf("unable to get key from keybase, %w", err)
	}

	// Get the transaction to verify.
	tx, err := getTransaction(cfg.DocPath, io)
	if err != nil {
		return err
	}

	// Verify signature.
	sig, err := getSignature(cfg, tx)
	if err != nil {
		return err
	}

	// Get account number and sequence if needed.
	signBytes, err := getSignBytes(ctx, cfg, info, tx, io)
	if err != nil {
		return err
	}

	err = kb.Verify(info.GetName(), signBytes, sig)
	if err == nil {
		if !cfg.RootCfg.BaseOptions.Quiet {
			io.Println("Valid signature!")
		}
	}
	return err
}

func getTransaction(docPath string, io commands.IO) (*std.Tx, error) {
	var msg []byte

	// Read document to sign.
	if docPath == "" { // From stdin.
		msgstr, err := io.GetString(
			"Enter document to sign.",
		)
		if err != nil {
			return nil, err
		}
		msg = []byte(msgstr)
	} else { // from file
		var err error
		msg, err = os.ReadFile(docPath)
		if err != nil {
			return nil, err
		}
	}

	// Unmarshal Amino JSON transaction.
	var tx std.Tx
	if err := amino.UnmarshalJSON(msg, &tx); err != nil {
		return nil, fmt.Errorf("unable to unmarshal transaction, %w", err)
	}
	return &tx, nil
}

func getSignature(cfg *VerifyCfg, tx *std.Tx) ([]byte, error) {
	// Exclude -sigpath and -signature flags set at the same time.
	if cfg.SigPath != "" && cfg.Signature != "" {
		return nil, errors.New("only one of -sigpath or -signature flags can be set")
	}

	// From -sigpath flag.
	if cfg.SigPath != "" {
		sigbz, err := os.ReadFile(cfg.SigPath)
		if err != nil {
			return nil, err
		}

		// Unmarshal Amino JSON signature.
		var sig std.Signature
		if err := amino.UnmarshalJSON(sigbz, &sig); err != nil {
			return nil, fmt.Errorf("unable to unmarshal signature, %w", err)
		}

		if sig.Signature == nil {
			return nil, errors.New("no signature found in the signature file")
		}

		return sig.Signature, nil
	}

	// From -signature flag.
	if cfg.Signature != "" {
		sig, err := base64.StdEncoding.DecodeString(cfg.Signature)
		if err != nil {
			return nil, fmt.Errorf("unable to decode signature, %w", err)
		}
		return sig, nil
	}

	// Default: from tx.
	if len(tx.Signatures) > 0 {
		return tx.Signatures[0].Signature, nil
	}

	return nil, errors.New("no signature found in the transaction")
}

func getSignBytes(ctx context.Context, cfg *VerifyCfg, info keys.Info, tx *std.Tx, io commands.IO) ([]byte, error) {
	// Query account number and sequence if needed.
	if !cfg.Offline && cfg.AccountNumber == 0 && cfg.AccountSequence == 0 {
		if !cfg.RootCfg.BaseOptions.Quiet {
			io.Println("Querying account from chain...")
		}

		// Query the account from the chain.
		baseAccount, err := queryBaseAccount(ctx, cfg, info)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not query account from chain, use default values: %v\n", err)
		} else {
			// Update cfg with queried account number and sequence.
			cfg.AccountNumber = baseAccount.AccountNumber
			cfg.AccountSequence = baseAccount.Sequence
			if !cfg.RootCfg.BaseOptions.Quiet {
				io.Printf("account-number set to %d\n", cfg.AccountNumber)
				io.Printf("account-sequence set to %d\n", cfg.AccountSequence)
			}
		}
	}

	// Get the bytes to verify.
	signBytes, err := tx.GetSignBytes(
		cfg.ChainID,
		cfg.AccountNumber,
		cfg.AccountSequence,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get signature bytes, %w", err)
	}

	return signBytes, nil
}

func queryBaseAccount(ctx context.Context, cfg *VerifyCfg, info keys.Info) (*std.BaseAccount, error) {
	remote := cfg.RootCfg.Remote
	if remote == "" {
		return nil, errors.New("missing remote url")
	}

	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, errors.Wrap(err, "new http client")
	}

	address := crypto.AddressToBech32(info.GetAddress())
	path := fmt.Sprintf("auth/accounts/%s", address)
	data := []byte{}

	qres, err := cli.ABCIQuery(ctx, path, data)
	if err != nil {
		return nil, errors.Wrap(err, "query account")
	}
	if len(qres.Response.Data) == 0 || string(qres.Response.Data) == "null" {
		return nil, errors.Wrap(err, "unknown address: "+address)
	}

	var qret struct{ BaseAccount std.BaseAccount }
	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
	if err != nil {
		return nil, err
	}

	return &qret.BaseAccount, nil
}
