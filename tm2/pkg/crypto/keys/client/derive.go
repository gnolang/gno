package client

import (
	"context"
	"errors"
	"flag"
	"math"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

var (
	errInvalidMnemonic     = errors.New("invalid bip39 mnemonic")
	errInvalidNumAccounts  = errors.New("invalid number of accounts")
	errInvalidAccountIndex = errors.New("invalid account index")
)

type deriveCfg struct {
	mnemonic     string
	numAccounts  uint64
	accountIndex uint64
}

// newDeriveCmd creates a new gnokey derive subcommand
func newDeriveCmd(io *commands.IO) *commands.Command {
	cfg := &deriveCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "derive",
			ShortUsage: "derive [flags]",
			ShortHelp:  "Derives the account addresses from the specified mnemonic",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execDerive(cfg, io)
		},
	)
}

func (c *deriveCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.mnemonic,
		"mnemonic",
		"",
		"the bip39 mnemonic",
	)

	fs.Uint64Var(
		&c.numAccounts,
		"num-accounts",
		10,
		"the number of accounts to derive from the mnemonic",
	)

	fs.Uint64Var(
		&c.accountIndex,
		"account-index",
		0,
		"the account index in the mnemonic",
	)
}

func execDerive(cfg *deriveCfg, io *commands.IO) error {
	// Make sure the number of accounts is valid
	if cfg.numAccounts == 0 || !isUint32(cfg.numAccounts) {
		return errInvalidNumAccounts
	}

	// Make sure the account index is valid
	if !isUint32(cfg.accountIndex) {
		return errInvalidAccountIndex
	}

	// Make sure the mnemonic is valid
	if !bip39.IsMnemonicValid(cfg.mnemonic) {
		return errInvalidMnemonic
	}

	// Generate the accounts
	accounts := generateAccounts(
		cfg.mnemonic,
		cfg.accountIndex,
		cfg.numAccounts,
	)

	io.Printf("[Generated Accounts]\n\n")
	io.Printf("Account Index: %d\n\n", cfg.accountIndex)

	// Print them out
	for index, account := range accounts {
		io.Printfln("%d. %s", index, account.String())
	}

	return nil
}

// isUint32 verifies a uint64 value can be represented
// as a uint32
func isUint32(value uint64) bool {
	return value <= math.MaxUint32
}

// generateAccounts the accounts using the provided mnemonics
func generateAccounts(mnemonic string, accountIndex, numAccounts uint64) []crypto.Address {
	addresses := make([]crypto.Address, numAccounts)

	// Generate the seed
	seed := bip39.NewSeed(mnemonic, "")

	for i := uint64(0); i < numAccounts; i++ {
		key := generateKeyFromSeed(seed, uint32(accountIndex), uint32(i))
		address := key.PubKey().Address()

		addresses[i] = address
	}

	return addresses
}

// generateKeyFromSeed generates a private key from
// the provided seed and index
func generateKeyFromSeed(seed []byte, account, index uint32) crypto.PrivKey {
	pathParams := hd.NewFundraiserParams(account, crypto.CoinType, index)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())

	return secp256k1.PrivKeySecp256k1(derivedPriv)
}
