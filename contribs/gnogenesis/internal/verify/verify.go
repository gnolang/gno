package verify

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

var (
	errInvalidGenesisState       = errors.New("invalid genesis state type")
	errInvalidTxSignature        = errors.New("invalid tx signature")
	errUncoveredGenesisValidator = errors.New("genesis validator has no corresponding valopers.Register migration tx")
)

// Realm path and function name for the valopers.Register MsgCall the
// coverage check matches. Mirrors the constants in
// contribs/gnogenesis/internal/fork/valoper_seed.go; kept duplicated
// here to avoid coupling verify to fork.
const (
	valopersPkgPath    = "gno.land/r/gnops/valopers"
	valopersRegisterFn = "Register"
)

type verifyCfg struct {
	common.Cfg

	skipSignatureCheck bool
}

// NewVerifyCmd creates the genesis verify subcommand
func NewVerifyCmd(io commands.IO) *commands.Command {
	cfg := &verifyCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "verify",
			ShortUsage: "[flags]",
			ShortHelp:  "verifies a genesis.json",
			LongHelp:   "Verifies a node's genesis.json",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execVerify(cfg, io)
		},
	)
}

func (c *verifyCfg) RegisterFlags(fs *flag.FlagSet) {
	c.Cfg.RegisterFlags(fs)

	fs.BoolVar(&c.skipSignatureCheck, "skip-signature-check", false,
		"skip per-tx signature verification. Genesis-mode txs can carry "+
			"signatures that intentionally don't verify (post-sign caller "+
			"overrides, valoper-seed placeholder signatures); nodes accept "+
			"them under --skip-genesis-sig-verification. Every other check "+
			"still runs.")
}

func execVerify(cfg *verifyCfg, io commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.GenesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Verify it
	if validateErr := genesis.Validate(); validateErr != nil {
		return fmt.Errorf("unable to verify genesis, %w", validateErr)
	}

	// Validate the genesis state
	if genesis.AppState != nil {
		state, ok := genesis.AppState.(gnoland.GnoGenesisState)
		if !ok {
			return errInvalidGenesisState
		}

		if err := gnoland.ValidateGenState(state); err != nil {
			return fmt.Errorf("invalid genesis state: %w", err)
		}

		// Validate the initial transactions
		for index, tx := range state.Txs {
			if validateErr := tx.Tx.ValidateBasic(); validateErr != nil {
				return fmt.Errorf("invalid transacton, %w", validateErr)
			}

			if cfg.skipSignatureCheck {
				continue
			}

			// Genesis txs can only be signed by 1 account.
			// Basic tx validation ensures there is at least 1 signer
			signer := tx.Tx.GetSignatures()[0]

			// Zero-value placeholder signatures carry no public key —
			// nothing to verify against.
			if signer.PubKey == nil {
				return fmt.Errorf(
					"%w #%d, missing signer public key",
					errInvalidTxSignature,
					index,
				)
			}

			// Grab the signature bytes of the tx.
			// Genesis transactions are signed with
			// account number and sequence set to 0
			signBytes, err := tx.Tx.GetSignBytes(genesis.ChainID, 0, 0)
			if err != nil {
				return fmt.Errorf("unable to get tx signature payload, %w", err)
			}

			// Verify the signature using the public key
			if !signer.PubKey.VerifyBytes(signBytes, signer.Signature) {
				return fmt.Errorf(
					"%w #%d, by signer %s",
					errInvalidTxSignature,
					index,
					signer.PubKey.Address(),
				)
			}
		}

		// Validate the initial balances
		for _, balance := range state.Balances {
			if err := balance.Verify(); err != nil {
				return fmt.Errorf("invalid balance: %w", err)
			}
		}

		// Hardfork-mode genesis valoper coverage: every entry in
		// GenesisDoc.Validators must have a matching valopers.Register
		// migration tx in state.Txs. See checkGenesisValoperCoverage
		// for why and the runtime gate it mirrors.
		if err := checkGenesisValoperCoverage(genesis, state); err != nil {
			return err
		}
	}

	io.Printfln("Genesis at %s is valid", cfg.GenesisPath)

	return nil
}

// checkGenesisValoperCoverage pre-flights the same invariant gnoland's
// InitChainer auto-asserts at boot under PastChainIDs (see
// shouldAssertValoperCoverage in gno.land/pkg/gnoland/app.go): every
// GenesisDoc.Validators entry must have a matching valopers.Register
// migration tx in state.Txs whose pubkey arg derives to the validator's
// signing address. Without coverage, the chain boots with orphan
// validators that v3's operator-keyed proposal flow can't manage — the
// test-13 footgun this check exists to catch.
//
// Gate mirrors the runtime exactly: PastChainIDs non-empty AND
// Validators non-empty. Fresh chains and dev/lazy-init setups skip both
// at runtime and here.
//
// Args[4] of the Register MsgCall is the signing pubkey (see
// buildRegisterTx in contribs/gnogenesis/internal/fork/valoper_seed.go).
// Unparseable pubkeys are skipped silently — the on-chain Register
// would reject them via bech32 decode, so this surface is downstream.
func checkGenesisValoperCoverage(genesis *types.GenesisDoc, state gnoland.GnoGenesisState) error {
	if len(state.PastChainIDs) == 0 || len(genesis.Validators) == 0 {
		return nil
	}

	covered := make(map[crypto.Address]bool, len(genesis.Validators))
	for _, tx := range state.Txs {
		// Runtime gno.land/pkg/gnoland/app.go:779-787 short-circuits
		// metadata.Failed=true txs before baseApp.Deliver, so a Failed
		// Register never populates valoperCache. Mirror that here so
		// verify doesn't report coverage the runtime won't honor.
		if tx.Metadata != nil && tx.Metadata.Failed {
			continue
		}
		for _, msg := range tx.Tx.Msgs {
			call, ok := msg.(vm.MsgCall)
			if !ok {
				continue
			}
			if call.PkgPath != valopersPkgPath || call.Func != valopersRegisterFn {
				continue
			}
			if len(call.Args) < 5 {
				continue
			}
			pk, err := crypto.PubKeyFromBech32(call.Args[4])
			if err != nil {
				continue
			}
			covered[pk.Address()] = true
		}
	}

	var uncovered []string
	for _, v := range genesis.Validators {
		if !covered[v.Address] {
			uncovered = append(uncovered, v.Address.String())
		}
	}
	if len(uncovered) > 0 {
		return fmt.Errorf("%w: %v", errUncoveredGenesisValidator, uncovered)
	}
	return nil
}
