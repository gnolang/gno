package client

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type RawVerifyCfg struct {
	RootCfg *BaseCfg

	SigPath       string
	PlainPath     string
	PlainEncoding string
}

func NewRawVerifyCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &RawVerifyCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "raw-verify",
			ShortUsage: "raw-verify [flags] <key-name or address>",
			ShortHelp:  "verifies the signature against raw plaintext bytes",
			LongHelp:   "Verifies a signature (expressed as a HEX string) of raw plaintext bytes (expressed as HEX bytes or JSON bytes) with pubkey identified by <key-name or address> in your local keybase. This is for advanced users. If you want to verify a signature against a transaction it is better to use 'verify' instead.",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execRawVerify(ctx, cfg, args, io)
		},
	)
}

func (c *RawVerifyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PlainPath,
		"plain-path",
		"",
		"path of plaintext file to verify",
	)
	fs.StringVar(
		&c.PlainEncoding,
		"plain-encoding",
		"json",
		"encoding of plaintext file: 'json' or 'hex'",
	)
	fs.StringVar(
		&c.SigPath,
		"sig-path",
		"",
		"path of signature file of HEX bytes",
	)
}

func execRawVerify(ctx context.Context, cfg *RawVerifyCfg, args []string, io commands.IO) error {
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

	// Get the plaintext bytes
	plainBytes, err := os.ReadFile(cfg.PlainPath)
	if err != nil {
		return fmt.Errorf("unable to read transaction file")
	}
	if len(plainBytes) == 0 {
		return errInvalidTxFile
	}

	// Decode sign (plaintext) bytes from plainBytes
	var signBytes []byte
	if cfg.PlainEncoding == "json" {
		// Make sure plainBytes is valid JSON
		var plain any
		if err := json.Unmarshal(plainBytes, &plain); err != nil {
			return fmt.Errorf("plaintext is not valid JSON %w", err)
		}
		// Bytes to verify are the JSON bytes.
		signBytes = plainBytes
	} else if cfg.PlainEncoding == "hex" {
		decoded, err := hex.DecodeString(string(plainBytes))
		if err != nil {
			return fmt.Errorf("plaintext is not valid HEX %w", err)
		}
		signBytes = decoded
	} else {
		panic("unrecognized plain-encoding; expected 'json' or 'hex'")
	}

	// Get the signature bytes
	sigRawBytes, err := os.ReadFile(cfg.SigPath)
	if err != nil {
		return fmt.Errorf("unable to read signature file")
	}
	sigRawBytes = []byte(strings.TrimSpace(string(sigRawBytes)))
	if len(sigRawBytes) == 0 {
		return fmt.Errorf("no signature found in the signature file")
	}

	// Decode signature bytes
	sigBytes, err := hex.DecodeString(string(sigRawBytes))
	if err != nil {
		return fmt.Errorf("signature is not valid HEX %w", err)
	}

	// Verify signature against sign bytes
	if err = kb.Verify(info.GetName(), signBytes, sigBytes); err != nil {
		return fmt.Errorf("unable to verify signature: %w", err)
	}

	if !cfg.RootCfg.BaseOptions.Quiet {
		io.Printf(
			"Valid signature!\nSigning Address: %s\nPublic key: %s\nSignature: %s\n",
			info.GetAddress(),
			info.GetPubKey().String(),
			base64.StdEncoding.EncodeToString(sigBytes),
		)
	}

	return nil
}
