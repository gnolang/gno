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
			LongHelp: `
DESCRIPTION
  Verify that a signature over a transaction (std.Tx, Amino JSON) matches the public
  key stored in your local keybase entry identified by <key-name or address>.

  The bytes to verify are computed exactly as during signing, i.e. using:
    - the Chain ID (-chain-id),
    - the account number (-account-number),
    - the account sequence (-account-sequence).

  If either -account-number or -account-sequence is 0 (the default) and -offline=false,
  the command will query the connected chain to fill in the missing values for the
  provided key's address. If the query fails, the defaults (0) are used and verification
  proceeds against those values.

USAGE
  gnokey verify [flags] <key-name or address>

ARGUMENTS
  <key-name or address>
      A local key name OR a bech32 address present in your keybase. The corresponding
      public key is used to verify the signature.

INPUTS
  - Transaction (required):
      Use -docpath to point to a file containing the transaction in Amino JSON (std.Tx).
      If -docpath is empty, the transaction is read from stdin (you will be prompted).

  - Signature (optional, source precedence):
      1) -sigpath: path to a file with a std.Signature (Amino JSON). The "Signature"
         field is used.
      2) -signature: a base64-encoded signature string.
      3) If neither is provided, the command uses tx.Signatures[0].Signature from the
         provided transaction.
      Note: Setting both -sigpath and -signature is an error.

FLAGS
  -docpath string
      Path to the transaction file (Amino JSON std.Tx). If empty, read from stdin.

  -sigpath string
      Path to a signature file (Amino JSON std.Signature). Mutually exclusive with -signature.

  -signature string
      Base64-encoded signature string. Mutually exclusive with -sigpath.

  -chain-id string (default: "dev")
      Chain ID used when reconstructing the sign bytes.

  -account-number uint (default: 0)
      Account number used when reconstructing the sign bytes. If 0 and -offline=false,
      it will be fetched from the chain.

  -account-sequence uint (default: 0)
      Account sequence used when reconstructing the sign bytes. If 0 and -offline=false,
      it will be fetched from the chain.

  -offline
      Do not query the chain for account number/sequence. If provided values are 0, they
      remain 0 for verification.

ROOT-LEVEL OPTIONS (examples)
  -home string
      Keybase/home directory.
  -remote string
      RPC endpoint to query account data when not offline (e.g. -remote http://127.0.0.1:26657).
  -q
      Quiet mode. Suppresses "Valid signature!" on success.

BEHAVIOR & NOTES
  • Only one signature is verified. If the transaction contains multiple signatures,
    the first one (index 0) is used unless -sigpath or -signature is provided.
  • Verification will only succeed if the chain-id, account-number, and account-sequence
    exactly match those used when the signature was produced.
  • The command does not modify your keybase or the chain; it only reads local data
    and (unless -offline) queries the account endpoint to populate missing parameters.

EXIT STATUS
  0  if the signature is valid.
  1  if verification fails or on other errors.

EXAMPLES
  # Verify using the signature embedded in the transaction, auto-fetching account params
  gnokey -remote http://127.0.0.1:26657 verify alice -docpath tx.json -chain-id dev

  # Verify using a separate signature file (Amino JSON std.Signature)
  gnokey verify bob -docpath tx.json -sigpath sig.json -chain-id mychain

  # Verify using a base64 signature string in fully offline mode
  gnokey verify g1xyz... -docpath tx.json -signature AbCdEf== -chain-id mychain -account-number 12 -account-sequence 34 -offline

  # Read the transaction from stdin
  echo "$(cat tx.json)\n" | gnokey verify alice -chain-id dev
`,
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
		"path of transaction file to verify in amino JSON format (if empty, read from stdin)",
	)
	fs.StringVar(
		&c.SigPath,
		"sigpath",
		"",
		"path of signature file in Amino JSON format",
	)
	fs.StringVar(
		&c.Signature,
		"signature",
		"",
		"base64-encoded signature string",
	)
	fs.StringVar(
		&c.ChainID,
		"chain-id",
		"dev",
		"The network chain ID",
	)
	fs.Uint64Var(
		&c.AccountNumber,
		"account-number",
		0,
		"The account number of the signing account. If not provided, it will be queried from the chain unless --offline is set.",
	)
	fs.Uint64Var(
		&c.AccountSequence,
		"account-sequence",
		0,
		"The account sequence of the signing account. If not provided, it will be queried from the chain unless --offline is set.",
	)
	fs.BoolVar(
		&c.Offline,
		"offline",
		false,
		"Offline mode. Do not query the chain for account number and account sequence.",
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

	// Fetch the key info from the keybase
	kb, err = keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return err
	}

	info, err := kb.GetByNameOrAddress(args[0])
	if err != nil {
		return fmt.Errorf("unable to get key from keybase, %w", err)
	}

	// get the transaction to verify.
	tx, err := getTransaction(cfg, io)
	if err != nil {
		return err
	}

	// verify signature.
	sig, err := getSignature(cfg, tx)
	if err != nil {
		return err
	}

	// Get account number and sequence if needed.
	getTxParameters(ctx, cfg, info, io)

	// Get the bytes to verify
	signBytes, err := tx.GetSignBytes(
		cfg.ChainID,
		cfg.AccountNumber,
		cfg.AccountSequence,
	)
	if err != nil {
		return fmt.Errorf("unable to get signature bytes, %w", err)
	}

	err = kb.Verify(info.GetName(), signBytes, sig)
	if err == nil {
		if !cfg.RootCfg.BaseOptions.Quiet {
			io.Println("Valid signature!")
		}
	}
	return err
}

func getTransaction(cfg *VerifyCfg, io commands.IO) (*std.Tx, error) {
	var msg []byte

	// read document to sign
	if cfg.DocPath == "" { // from stdin.
		msgstr, err := io.GetString(
			"Enter document to sign.",
		)
		if err != nil {
			return nil, err
		}
		msg = []byte(msgstr)
	} else { // from file
		var err error
		msg, err = os.ReadFile(cfg.DocPath)
		if err != nil {
			return nil, err
		}
	}

	// Unmarshal transaction Amino JSON
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

	// from -sigpath flag
	if cfg.SigPath != "" {
		sigbz, err := os.ReadFile(cfg.SigPath)
		if err != nil {
			return nil, err
		}

		// Unmarshal transaction Amino JSON
		var sig std.Signature
		if err := amino.UnmarshalJSON(sigbz, &sig); err != nil {
			return nil, fmt.Errorf("unable to unmarshal signature, %w", err)
		}

		if sig.Signature == nil {
			return nil, errors.New("no signature found in the signature file")
		}

		return sig.Signature, nil
	}

	// from -signature flag
	if cfg.Signature != "" {
		sig, err := base64.StdEncoding.DecodeString(cfg.Signature)
		if err != nil {
			return nil, fmt.Errorf("unable to decode signature, %w", err)
		}
		return sig, nil
	}

	// default: from tx
	if tx.Signatures != nil && len(tx.Signatures) > 0 {
		return tx.Signatures[0].Signature, nil
	}

	return nil, errors.New("no signature found in the transaction")
}

func getTxParameters(ctx context.Context, cfg *VerifyCfg, info keys.Info, io commands.IO) {
	// Query the chain if -offline=false and account number or sequence are equal to 0.
	if !cfg.Offline && (cfg.AccountNumber == 0 || cfg.AccountSequence == 0) {
		if !cfg.RootCfg.BaseOptions.Quiet {
			io.Println("Querying account from chain...")
		}
		// Query the account from the chain.
		baseAccount, err := queryBaseAccount(ctx, cfg, info)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not query account from chain, use default values: %v\n", err)
		} else {
			if cfg.AccountNumber == 0 {
				if !cfg.RootCfg.BaseOptions.Quiet {
					io.Printf("account-number set to %d\n", baseAccount.AccountNumber)
				}
				cfg.AccountNumber = baseAccount.AccountNumber
			}
			if cfg.AccountSequence == 0 {
				if !cfg.RootCfg.BaseOptions.Quiet {
					io.Printf("account-sequence set to %d\n", baseAccount.Sequence)
				}
				cfg.AccountSequence = baseAccount.Sequence
			}
		}
	}
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
