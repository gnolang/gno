package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type VerifyDataCfg struct {
	RootCfg *BaseCfg

	ChainID   string
	Origin    string
	Purpose   string
	Data      string
	DataPath  string
	ProofPath string
}

func NewVerifyDataCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &VerifyDataCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "verify-data",
			ShortUsage: "verify-data [flags] <address or key-name>",
			ShortHelp:  "verifies a proof made by sign-data",
			LongHelp: `Verifies a proof of control of an address, as gnokey sign-data makes it. The
proof holds the public key, so the address needs no key in the keybase; a key
name is looked up in it.

The proof passes only if it is signed by the address's key and signs exactly
-data on -chainid, for -origin and -purpose: a valid signature over another chain,
service, purpose or text is refused. A proof stays valid forever: a service
hands out each challenge once, with an expiry, and accepts it once.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execVerifyData(cfg, args, io)
		},
	)
}

func (c *VerifyDataCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.ChainID,
		"chainid",
		"",
		"the ID of the chain the proof must be for (required)",
	)

	fs.StringVar(
		&c.Origin,
		"origin",
		"",
		"the service the proof must be for (required)",
	)

	fs.StringVar(
		&c.Purpose,
		"purpose",
		"",
		"what the proof must be for (required)",
	)

	fs.StringVar(
		&c.Data,
		"data",
		"",
		"the text the proof must sign, typically the challenge handed out",
	)

	fs.StringVar(
		&c.DataPath,
		"data-path",
		"",
		"path to a file holding the text the proof must sign, as it is (instead of -data)",
	)

	fs.StringVar(
		&c.ProofPath,
		"proof-path",
		"",
		`path to the proof, as JSON; "-" reads it from stdin (required)`,
	)
}

func execVerifyData(cfg *VerifyDataCfg, args []string, cio commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	if err := checkProofScope(cfg.ChainID, cfg.Origin, cfg.Purpose); err != nil {
		return err
	}

	if cfg.ProofPath == "" {
		return errors.New(`no proof: give it with -proof-path, or "-proof-path -" for stdin`)
	}

	data, err := readSignDataText(cfg.Data, cfg.DataPath)
	if err != nil {
		return err
	}

	signer, err := resolveSigner(cfg.RootCfg.Home, args[0])
	if err != nil {
		return err
	}

	// No proof matches expectations that make no document, such as an origin
	// with a trailing slash: say so, rather than report a mismatch.
	if _, err := std.SignDataPayload(std.SignedData{
		ChainID: cfg.ChainID,
		Origin:  cfg.Origin,
		Purpose: cfg.Purpose,
		Signer:  signer,
		Data:    data,
	}); err != nil {
		return err
	}

	var raw []byte
	if cfg.ProofPath == "-" {
		raw, err = io.ReadAll(cio.In())
	} else {
		raw, err = os.ReadFile(cfg.ProofPath)
	}
	if err != nil {
		return fmt.Errorf("unable to read proof, %w", err)
	}

	signed, err := std.VerifySignDataProof(raw)
	if err != nil {
		return err
	}

	switch {
	case signed.ChainID != cfg.ChainID:
		return fmt.Errorf("the proof is for chain %q, not %q", signed.ChainID, cfg.ChainID)
	case signed.Origin != cfg.Origin:
		return fmt.Errorf("the proof is for origin %q, not %q", signed.Origin, cfg.Origin)
	case signed.Purpose != cfg.Purpose:
		return fmt.Errorf("the proof is for purpose %q, not %q", signed.Purpose, cfg.Purpose)
	case signed.Signer != signer:
		return fmt.Errorf("the proof is signed by %s, not %s", signed.Signer, signer)
	case signed.Data != data:
		return fmt.Errorf("the proof signs other data: %q", signed.Data)
	}

	cio.Printf("Valid proof: %s signed the data for %s (%s) on %s\n", signed.Signer, signed.Origin, signed.Purpose, signed.ChainID)

	return nil
}

// resolveSigner returns the address nameOrBech32 names: itself, or the address
// of the key by that name.
func resolveSigner(home, nameOrBech32 string) (crypto.Address, error) {
	if addr, err := crypto.AddressFromBech32(nameOrBech32); err == nil {
		return addr, nil
	}

	kb, err := keys.NewKeyBaseFromDir(home)
	if err != nil {
		return crypto.Address{}, fmt.Errorf("unable to load keybase, %w", err)
	}

	info, err := kb.GetByName(nameOrBech32)
	if err != nil {
		return crypto.Address{}, fmt.Errorf("unable to get key from keybase, %w", err)
	}

	return info.GetAddress(), nil
}
