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
			ShortUsage: "verify [flags] <key-name or address> <transaction path>",
			ShortHelp:  "verifies the transaction signature",
			LongHelp:   "Verifies a signature of a <Amino JSON format transaction> against <key-name or address> in your local keybase. The sign bytes are derived from the tx using --chain-id, --account-number, and --account-sequence; these must match the values used when the signature was created. If --account-number, --account-sequence or --chain-id are not set, the command queries the chain (via --remote) to fill them; if the query fails, default values are used. Provide the signature via --sigpath; otherwise the first signature in the tx (tx.Signatures[0]) is used.",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execVerify(ctx, cfg, args, io)
		},
	)
}

func (c *VerifyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.SigPath,
		"sigpath",
		"",
		"path of signature file in Amino JSON format (mutually exclusive with -signature flag)",
	)
	fs.StringVar(
		&c.ChainID,
		"chain-id",
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

	if len(args) != 2 {
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
	tx, err := getTransaction(args[1], io)
	if err != nil {
		return err
	}

	// Verify signature.
	sig, err := getSignature(cfg, tx)
	if err != nil {
		return err
	}

	// Update cfg.ChainID if empty.
	if cfg.ChainID == "" {
		updateCfgChainID(ctx, cfg, io)
	}

	// Update account number and sequence if needed.
	if !cfg.AccountNumber.Defined || !cfg.AccountSequence.Defined {
		updateCfgAccountParams(ctx, cfg, info, io)
	}

	// Get account number and sequence if needed.
	signBytes, err := getSignBytes(cfg, tx)
	if err != nil {
		return err
	}

	err = kb.Verify(info.GetName(), signBytes, sig)
	if err == nil {
		if !cfg.RootCfg.BaseOptions.Quiet {
			io.Printf("Valid signature!\nSigning Address: %s\nPublic key: %s\nSignature: %s\n", info.GetAddress(), info.GetPubKey().String(), base64.StdEncoding.EncodeToString(sig))
		}
	}
	return err
}

func getTransaction(txPath string, io commands.IO) (*std.Tx, error) {
	// Read document to sign.
	msg, err := os.ReadFile(txPath)
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

func getSignature(cfg *VerifyCfg, tx *std.Tx) ([]byte, error) {
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

	// Default: from tx.
	if len(tx.Signatures) > 0 {
		return tx.Signatures[0].Signature, nil
	}

	return nil, errors.New("no signature found in the transaction")
}

func updateCfgChainID(ctx context.Context, cfg *VerifyCfg, io commands.IO) {
	if cfg.ChainID == "" {
		chainID, err := queryNodeStatus(ctx, cfg.RootCfg.Remote)
		if err != nil {
			io.ErrPrintfln("Warning: could not query chain-id from chain, use default value: %v", err)
		} else {
			if !cfg.RootCfg.BaseOptions.Quiet {
				io.Printf("Queried chain-id from network: %s\n", chainID)
			}
			cfg.ChainID = chainID
		}
	}
}

func queryNodeStatus(ctx context.Context, remote string) (string, error) {
	if remote == "" {
		return "", errors.New("missing remote url")
	}

	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return "", errors.Wrap(err, "new http client")
	}

	// Get the node status to query the chain ID if needed.
	nodeStatus, err := cli.Status(ctx, nil)
	if err != nil {
		return "", errors.Wrap(err, "query node status")
	}

	return nodeStatus.NodeInfo.Network, nil
}

func updateCfgAccountParams(ctx context.Context, cfg *VerifyCfg, info keys.Info, io commands.IO) {
	// Query the account from the chain.
	baseAccount, err := queryBaseAccount(ctx, cfg.RootCfg.Remote, info)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not query account from chain, use default values: %v\n", err)
	} else {
		// Update cfg with queried account number and sequence.
		if !cfg.AccountNumber.Defined {
			if !cfg.RootCfg.BaseOptions.Quiet {
				io.ErrPrintfln("Queried account number from chain: %d", baseAccount.AccountNumber)
			}
			cfg.AccountNumber.V = baseAccount.AccountNumber
		}

		if !cfg.AccountSequence.Defined {
			if !cfg.RootCfg.BaseOptions.Quiet {
				io.Printf("Queried account sequence from chain: %d\n", baseAccount.Sequence)
			}
			cfg.AccountSequence.V = baseAccount.Sequence
		}
	}
}

func queryBaseAccount(ctx context.Context, remote string, info keys.Info) (*std.BaseAccount, error) {
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

func getSignBytes(cfg *VerifyCfg, tx *std.Tx) ([]byte, error) {
	// Get the bytes to verify.
	signBytes, err := tx.GetSignBytes(
		cfg.ChainID,
		cfg.AccountNumber.V,
		cfg.AccountSequence.V,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get signature bytes, %w", err)
	}

	return signBytes, nil
}
