package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	errInvalidMultisigKey   = errors.New("provided key is not a multisig reference")
	errNoSignaturesProvided = errors.New("no signatures provided")
)

type MultisignCfg struct {
	RootCfg *BaseCfg

	TxPath     string
	Signatures commands.StringArr
}

func NewMultisignCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &MultisignCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "multisign",
			ShortUsage: "multisign [flags] <multisig key-name or address>",
			ShortHelp:  "combines the multisigs for the tx document and saves it to disk",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMultisign(cfg, args, io)
		},
	)
}

func (c *MultisignCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.TxPath,
		"tx-path",
		"",
		"path to the Amino JSON-encoded tx (file) to sign",
	)

	fs.Var(
		&c.Signatures,
		"signature",
		"the path to the signature.json",
	)
}

func execMultisign(cfg *MultisignCfg, args []string, io commands.IO) error {
	// Make sure the multisig key name is provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	// Make sure at least one signature is provided
	if len(cfg.Signatures) == 0 {
		return errNoSignaturesProvided
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

	// Make sure the key is referencing a multisig key
	if info.GetType() != keys.TypeMulti {
		return fmt.Errorf("%w: %q", errInvalidMultisigKey, args[0])
	}

	// Get the transaction bytes
	txRaw, err := os.ReadFile(cfg.TxPath)
	if err != nil {
		return fmt.Errorf("unable to read transaction file: %w", err)
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

	var (
		pubKey = info.GetPubKey()

		// The code below perfectly highlights the overengineering of the SDK.
		// The keybase works exclusively with an abstraction of a public key (crypto.PubKey),
		// so even if the key type is multisig (multiple public keys), there is no way
		// to access the crypto.PubKey's internal fields (nor should there be, as this is an abstraction).
		//
		// This becomes a problem when we need to create a multisig, since to initialize
		// the multisig object, we need to specify _the exact number of keys in the multisig_ (N),
		// even though we just want to aggregate a few signatures together (K).
		// As there is no way to extract this information from a crypto.PubKey, we need to _cast_
		// the public key to a multisig one, and access its fields, in order to find N.
		// So much for the abstraction dance.
		multisigPub = pubKey.(multisig.PubKeyMultisigThreshold)
		multisigSig = multisig.NewMultisig(len(multisigPub.PubKeys))
	)

	for _, sigPath := range cfg.Signatures {
		// Load the signature
		sigRaw, err := os.ReadFile(sigPath)
		if err != nil {
			return fmt.Errorf("unable to read signature file %q: %w", sigPath, err)
		}

		var sig std.Signature
		if err = amino.UnmarshalJSON(sigRaw, &sig); err != nil {
			return fmt.Errorf("unable to parse signature file %q: %w", sigPath, err)
		}

		// Add it to the multisig
		if err = multisigSig.AddSignatureFromPubKey(
			sig.Signature,
			sig.PubKey,
			multisigPub.PubKeys,
		); err != nil {
			return fmt.Errorf("unable to add signature: %w", err)
		}
	}

	// Construct the signature
	sig := &std.Signature{
		PubKey:    pubKey,
		Signature: multisigSig.Marshal(),
	}

	// Save the signature to the tx
	if err = addSignature(&tx, sig); err != nil {
		return fmt.Errorf("unable to add signature to the tx: %w", err)
	}

	// Save the tx to disk
	if err = saveTx(&tx, cfg.TxPath); err != nil {
		return fmt.Errorf("unable to save tx: %w", err)
	}

	io.Printf("\nTx successfully signed and saved to %s\n", cfg.TxPath)

	return nil
}

// saveTx saves the given transaction to the given path (Amino-encoded JSON)
func saveTx(tx *std.Tx, path string) error {
	// Encode the transaction
	encodedTx, err := amino.MarshalJSON(tx)
	if err != nil {
		return fmt.Errorf("unable to marshal tx to JSON, %w", err)
	}

	// Save the transaction
	if err := os.WriteFile(path, encodedTx, 0o644); err != nil {
		return fmt.Errorf("unable to write tx to %s, %w", path, err)
	}

	return nil
}
