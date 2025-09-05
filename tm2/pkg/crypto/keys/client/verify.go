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

	DocPath         string
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
			ShortUsage: "verify [flags] <key-name or address> <signature>",
			ShortHelp:  "verifies the document signature",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execVerify(cfg, args, io)
		},
	)
}

func (c *VerifyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.DocPath,
		"docpath",
		"",
		"path of document file to verify",
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
		"Offline mode. Do not query the chain for account number and sequence.",
	)
}

func execVerify(cfg *VerifyCfg, args []string, io commands.IO) error {
	var (
		kb  keys.Keybase
		err error
	)

	if len(args) != 2 {
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

	var msg []byte
	docpath := cfg.DocPath

	// read document to sign
	if docpath == "" { // from stdin.
		msgstr, err := io.GetString(
			"Enter document to sign.",
		)
		if err != nil {
			return err
		}
		msg = []byte(msgstr)
	} else { // from file
		msg, err = os.ReadFile(docpath)
		if err != nil {
			return err
		}
	}

	// Unmarshal transaction Amino JSON
	var tx std.Tx
	if err := amino.UnmarshalJSON(msg, &tx); err != nil {
		return fmt.Errorf("unable to unmarshal transaction, %w", err)
	}

	// Query the chain if -offline=false and account number or sequence are equal to 0.
	if !cfg.Offline && (cfg.AccountNumber == 0 || cfg.AccountSequence == 0) {
		if !cfg.RootCfg.BaseOptions.Quiet {
			io.Println("Querying account from chain...")
		}
		// Query the account from the chain.
		baseAccount, err := queryBaseAccount(cfg, args, info)
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

	// Get the bytes to verify
	signBytes, err := tx.GetSignBytes(
		cfg.ChainID,
		cfg.AccountNumber,
		cfg.AccountSequence,
	)
	if err != nil {
		return fmt.Errorf("unable to get signature bytes, %w", err)
	}
	// validate document to sign.
	// XXX

	// verify signature.
	sig, err := parseSignature(args[1])
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

func queryBaseAccount(cfg *VerifyCfg, args []string, info keys.Info) (*std.BaseAccount, error) {
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

	qres, err := cli.ABCIQuery(context.Background(), path, data)
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

func parseSignature(sigstr string) ([]byte, error) {
	sig, err := base64.StdEncoding.DecodeString(sigstr)
	if err != nil {
		return nil, err
	}
	return sig, nil
}
