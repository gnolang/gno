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

	SigPath string
	TxPath  string

	ChainID         string
	AccountNumber   commands.Uint64Flag
	AccountSequence commands.Uint64Flag
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
			LongHelp:   "Verifies a signature of a <Amino JSON format transaction> with pubkey identified by <key-name or address> in your local keybase. The sign bytes are derived from the tx using --chain-id, --account-number, and --account-sequence; these must match the values used when the signature was created. If --account-number, --account-sequence or --chain-id are not set, the command queries the chain (via --remote) to fill them; if the query fails, default values are used. Provide the signature via --sigpath; otherwise the first signature in the tx (tx.Signatures[0]) is used.",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execVerify(ctx, cfg, args, io)
		},
	)
}

func (c *VerifyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.TxPath,
		"tx-path",
		"",
		"path of transaction file in Amino JSON format",
	)
	fs.StringVar(
		&c.SigPath,
		"sig-path",
		"",
		"path of signature file in Amino JSON format. If omitted, the signature in the tx itself is verified instead",
	)
	fs.StringVar(
		&c.ChainID,
		"chainid",
		"",
		"network chain ID used for signing",
	)
	fs.Var(
		&c.AccountNumber,
		"account-number",
		"account number of the signing account",
	)
	fs.Var(
		&c.AccountSequence,
		"account-sequence",
		"account sequence of the signing account",
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
	tx, err := readTransaction(cfg.TxPath)
	if err != nil {
		return fmt.Errorf("unable to read transaction from disk: %w", err)
	}

	// Fetch the signature
	var sig []byte

	if cfg.SigPath != "" {
		// The signature is in a separate file
		sig, err = readSignature(cfg.SigPath)
	} else {
		// The signature is in the tx itself
		sig, err = extractSignature(tx)
	}

	if err != nil {
		return fmt.Errorf("unable to get tx signature: %w", err)
	}

	var (
		remote = cfg.RootCfg.Remote

		chainID         = cfg.ChainID
		accountNumber   = cfg.AccountNumber.V
		accountSequence = cfg.AccountSequence.V
	)

	// Fetch the chain from the node if unset
	if chainID == "" {
		remoteChainID, err := fetchChainID(ctx, remote)
		if err != nil {
			return fmt.Errorf("unable to fetch chain ID: %w", err)
		}

		chainID = remoteChainID
	}

	// Update account number and sequence if needed.
	if !cfg.AccountNumber.Defined || !cfg.AccountSequence.Defined {
		// Fetch the latest account information
		account, err := fetchAccount(ctx, remote, info.GetAddress())
		if err != nil {
			return fmt.Errorf("unable to fetch account: %w", err)
		}

		// Update cfg with queried account number and sequence.
		if !cfg.AccountNumber.Defined {
			if !cfg.RootCfg.BaseOptions.Quiet {
				io.Printfln("Queried account number from chain: %d", account.AccountNumber)
			}

			accountNumber = account.AccountNumber
		}

		if !cfg.AccountSequence.Defined {
			if !cfg.RootCfg.BaseOptions.Quiet {
				io.Printfln("Queried account sequence from chain: %d", account.Sequence)
			}

			accountSequence = account.Sequence
		}
	}

	// Get the bytes to verify
	signBytes, err := tx.GetSignBytes(
		chainID,
		accountNumber,
		accountSequence,
	)
	if err != nil {
		return fmt.Errorf("unable to get signature bytes, %w", err)
	}

	if err = kb.Verify(info.GetName(), signBytes, sig); err != nil {
		return fmt.Errorf("unable to verify signature: %w", err)
	}

	if !cfg.RootCfg.BaseOptions.Quiet {
		io.Printf(
			"Valid signature!\nSigning Address: %s\nPublic key: %s\nSignature: %s\n",
			info.GetAddress(),
			info.GetPubKey().String(),
			base64.StdEncoding.EncodeToString(sig),
		)
	}

	return nil
}

// readTransaction loads the transaction from the given path
func readTransaction(path string) (*std.Tx, error) {
	// Read document to sign.
	msg, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read the transaction file, %w", err)
	}

	// Unmarshal Amino JSON transaction.
	var tx std.Tx
	if err := amino.UnmarshalJSON(msg, &tx); err != nil {
		return nil, fmt.Errorf("unable to unmarshal transaction, %w", err)
	}

	return &tx, nil
}

// readSignature reads the signature from the given path
func readSignature(path string) ([]byte, error) {
	// Read the signature file (separate, for multisigs)
	sigbz, err := os.ReadFile(path)
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

// extractSignature extracts the transaction signature
func extractSignature(tx *std.Tx) ([]byte, error) {
	if len(tx.Signatures) > 0 {
		// By default, TM2 always verifies and handles the first signature in the set
		return tx.Signatures[0].Signature, nil
	}

	return nil, errors.New("no signature found in the transaction")
}

// fetchChainID fetches the chain ID from the given remote
func fetchChainID(ctx context.Context, remote string) (string, error) {
	if remote == "" {
		return "", errors.New("missing remote url")
	}

	// Create the client
	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return "", fmt.Errorf("unable to create HTTP client: %w", err)
	}

	// Fetch the node status
	nodeStatus, err := cli.Status(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("unable to query node status: %w", err)
	}

	return nodeStatus.NodeInfo.Network, nil
}

// fetchAccount fetches the account from the given remote
func fetchAccount(
	ctx context.Context,
	remote string,
	address crypto.Address,
) (*std.BaseAccount, error) {
	if remote == "" {
		return nil, errors.New("missing remote url")
	}

	// Create the client
	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP client: %w", err)
	}

	// Query the account
	qres, err := cli.ABCIQuery(ctx, fmt.Sprintf("auth/accounts/%s", address), []byte{})
	if err != nil {
		return nil, fmt.Errorf("unable to query account: %w", err)
	}

	if len(qres.Response.Data) == 0 || string(qres.Response.Data) == "null" {
		return nil, fmt.Errorf("account is not initialized: %s", address.String())
	}

	var qret struct{ BaseAccount std.BaseAccount }

	if err = amino.UnmarshalJSON(qres.Response.Data, &qret); err != nil {
		return nil, fmt.Errorf("unable to unmarshal Amino JSON: %w", err)
	}

	return &qret.BaseAccount, nil
}
