package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type SignDataCfg struct {
	RootCfg *BaseCfg

	ChainID        string
	Data           string
	DataPath       string
	OutputDocument string
}

func NewSignDataCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &SignDataCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "sign-data",
			ShortUsage: "sign-data [flags] <key-name or address>",
			ShortHelp:  "signs a text to prove control of the key's address, off-chain",
			LongHelp: `Signs a text, typically a challenge handed out by a service, to prove that the
key's address is yours. The signed document is the one Adena's SignArbitrary
signs (ADR-036's sign/MsgSignData, bound to -chainid): it is not a transaction,
and no node can decode it as one.

The output is Adena's SignArbitrary answer, as JSON: the signed document, the
public key and the signature. Check it with std.VerifySignData.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSignData(cfg, args, io)
		},
	)
}

func (c *SignDataCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.ChainID,
		"chainid",
		"",
		"the ID of the chain the proof is for (required)",
	)

	fs.StringVar(
		&c.Data,
		"data",
		"",
		"the text to sign",
	)

	fs.StringVar(
		&c.DataPath,
		"data-path",
		"",
		"path to a file holding the text to sign, signed as it is (instead of -data)",
	)

	fs.StringVar(
		&c.OutputDocument,
		"output-document",
		"",
		"the file to save the proof to. If empty, prints it",
	)
}

// signDataProof is Adena's SignArbitrary answer, so that a verifier handles a
// proof from either the same way.
type signDataProof struct {
	Signed    json.RawMessage   `json:"signed"`
	Signature signDataSignature `json:"signature"`
}

type signDataSignature struct {
	PubKey    signDataPubKey `json:"pubKey"`
	Signature string         `json:"signature"` // base64, 64 bytes R||S
}

type signDataPubKey struct {
	TypeURL string `json:"typeUrl"`
	Value   string `json:"value"` // base64 of the compressed key itself
}

func execSignData(cfg *SignDataCfg, args []string, io commands.IO) error {
	// Make sure the key name is provided
	if len(args) != 1 {
		return flag.ErrHelp
	}

	if cfg.ChainID == "" {
		return errors.New("-chainid is required: a proof is only good on the chain it names")
	}

	data := cfg.Data
	switch {
	case cfg.Data != "" && cfg.DataPath != "":
		return errors.New("give the data with -data or -data-path, not both")
	case cfg.DataPath != "":
		raw, err := os.ReadFile(cfg.DataPath)
		if err != nil {
			return fmt.Errorf("unable to read data file, %w", err)
		}
		data = string(raw)
	case cfg.Data == "":
		return errors.New("no data to sign: give it with -data or -data-path")
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

	pub, ok := info.GetPubKey().(secp256k1.PubKeySecp256k1)
	if !ok {
		return fmt.Errorf("sign-data needs a secp256k1 key, %s is a %T", args[0], info.GetPubKey())
	}

	payload, err := std.SignDataPayload(cfg.ChainID, info.GetAddress(), data)
	if err != nil {
		return err
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

	sig, _, err := kb.Sign(args[0], password, payload)
	if err != nil {
		return fmt.Errorf("unable to sign data, %w", err)
	}

	proof, err := json.MarshalIndent(signDataProof{
		Signed: payload,
		Signature: signDataSignature{
			PubKey: signDataPubKey{
				TypeURL: "/tm.PubKeySecp256k1",
				Value:   base64.StdEncoding.EncodeToString(pub[:]),
			},
			Signature: base64.StdEncoding.EncodeToString(sig),
		},
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal proof to JSON, %w", err)
	}

	if cfg.OutputDocument == "" {
		io.Println(string(proof))
		return nil
	}

	if err := os.WriteFile(cfg.OutputDocument, proof, 0o644); err != nil {
		return fmt.Errorf("unable to write proof to %s, %w", cfg.OutputDocument, err)
	}

	io.Printf("\nProof generated and successfully saved to %s\n", cfg.OutputDocument)

	return nil
}
