package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type RawSignCfg struct {
	RootCfg *BaseCfg

	PlainPath      string
	PlainEncoding  string
	OutputDocument string
}

func NewRawSignCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &RawSignCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "raw-sign",
			ShortUsage: "raw-sign [flags] <key-name or address>",
			ShortHelp:  "(UNSAFE) signs the given raw bytes. This command is for advanced users. Use the 'sign' command instead to sign transactions",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRawSign(cfg, args, io)
		},
	)
}

func (c *RawSignCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PlainPath,
		"plain-path",
		"",
		"path to the plaintext file to sign",
	)

	fs.StringVar(
		&c.PlainEncoding,
		"plain-encoding",
		"json",
		"encoding of plaintext file: 'json' or 'hex'",
	)

	fs.StringVar(
		&c.OutputDocument,
		"output-document",
		"",
		"the signature json document to save. If empty, outputs the signature in the terminal",
	)
}

func execRawSign(cfg *RawSignCfg, args []string, io commands.IO) error {
	// Make sure the key name is provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	// saveSignature saves the given signature to the given path (Amino-encoded JSON)
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

	// Get the plaintext bytes
	plainBytes, err := os.ReadFile(cfg.PlainPath)
	if err != nil {
		return fmt.Errorf("unable to read transaction file")
	}

	// Make sure there is something to actually sign
	if len(plainBytes) == 0 {
		return errInvalidTxFile
	}

	var signBytes []byte
	if cfg.PlainEncoding == "json" {
		// Make sure plainBytes is valid JSON
		var plain any
		if err := json.Unmarshal(plainBytes, &plain); err != nil {
			return fmt.Errorf("plaintext is not valid JSON %w", err)
		}
		// Bytes to sign are the JSON bytes.
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
	kOpts := keyOpts{
		keyName:     args[0],
		decryptPass: password,
	}

	// Generate the signature
	signature, err := generateRawSignature(signBytes, kb, kOpts)
	if err != nil {
		return fmt.Errorf("unable to sign transaction, %w", err)
	}

	if cfg.OutputDocument == "" {
		io.Printf("signature hex: %X\n", signature.Signature)
	} else {
		// Don't save the signature in-place, separate it
		return saveSignature(signature, cfg.OutputDocument)
	}

	return nil
}

// generateRawSignature generates a signature for the given sign bytes.
func generateRawSignature(
	signBytes []byte,
	kb keys.Keybase,
	keyOpts keyOpts,
) (*std.Signature, error) {

	// Sign the bytes.
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
