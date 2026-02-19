package gnoland

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	ErrBalanceEmptyAddress = errors.New("balance address is empty")
	ErrBalanceEmptyAmount  = errors.New("balance amount is empty")
)

const (
	// XXX rename these to flagXyz.

	// flagTokenLockWhitelisted allows unrestricted transfers.
	flagTokenLockWhitelisted BitSet = 1 << iota

	// TODO: flagValidatorAccount marks an account as validator.
	flagValidatorAccount

	// TODO: flagRealmAccount marks an account as realm.
	flagRealmAccount

	flagFrozen
)

// bitSet represents a set of flags stored in a 64-bit unsigned integer.
// Each bit in the BitSet corresponds to a specific flag.
type BitSet uint64

func (bs BitSet) String() string {
	return fmt.Sprintf("0x%016X", uint64(bs)) // Show all 64 bits
}

var _ std.AccountUnrestricter = &GnoAccount{}

type GnoAccount struct {
	std.BaseAccount
	Attributes BitSet `json:"attributes" yaml:"attributes"`
}

// validFlags defines the set of all valid flags that can be used with BitSet.
const validFlags = flagTokenLockWhitelisted | flagValidatorAccount | flagRealmAccount | flagFrozen

func (ga *GnoAccount) setFlag(flag BitSet) {
	if !isValidFlag(flag) {
		panic(fmt.Sprintf("setFlag: invalid flag %d (binary: %b). Valid flags: %b", flag, flag, validFlags))
	}
	ga.Attributes |= flag
}

func (ga *GnoAccount) clearFlag(flag BitSet) {
	if !isValidFlag(flag) {
		panic(fmt.Sprintf("clearFlag: invalid flag %d (binary: %b). Valid flags: %b", flag, flag, validFlags))
	}
	ga.Attributes &= ^flag
}

func (ga *GnoAccount) hasFlag(flag BitSet) bool {
	if !isValidFlag(flag) {
		panic(fmt.Sprintf("hasFlag: invalid flag %d (binary: %b). Valid flags: %b", flag, flag, validFlags))
	}
	return ga.Attributes&flag != 0
}

// isValidFlag ensures that a given BitSet uses only the allowed subset of bits
// as defined in validFlags. This prevents accidentally setting invalid flags,
// especially since BitSet can represent all 64 bits of a uint64.
func isValidFlag(flag BitSet) bool {
	return flag&^validFlags == 0 && flag != 0
}

// SetTokenLockWhitelisted allows the account to bypass global transfer locking restrictions.
// By default, accounts are restricted with token transfer when global transfer locking is enabled.
func (ga *GnoAccount) SetTokenLockWhitelisted(whitelisted bool) {
	if whitelisted {
		ga.setFlag(flagTokenLockWhitelisted)
	} else {
		ga.clearFlag(flagTokenLockWhitelisted)
	}
}

// IsTokenLockWhitelisted checks whether the account is white listed for the token locking
func (ga *GnoAccount) IsTokenLockWhitelisted() bool {
	return ga.hasFlag(flagTokenLockWhitelisted)
}

func (ga *GnoAccount) SetFrozen(frozen bool) {
	if frozen {
		ga.setFlag(flagFrozen)
	} else {
		ga.clearFlag(flagFrozen)
	}
}

func (ga *GnoAccount) IsFrozen() bool {
	return ga.hasFlag(flagFrozen)
}

// String implements fmt.Stringer
func (ga *GnoAccount) String() string {
	return fmt.Sprintf("%s\n  Attributes:	 %s",
		ga.BaseAccount.String(),
		ga.Attributes.String(),
	)
}

func ProtoGnoAccount() std.Account {
	return &GnoAccount{}
}

type GnoGenesisState struct {
	Balances []Balance         `json:"balances"`
	Txs      []TxWithMetadata  `json:"txs"`
	Auth     auth.GenesisState `json:"auth"`
	Bank     bank.GenesisState `json:"bank"`
	VM       vm.GenesisState   `json:"vm"`
}

type TxWithMetadata struct {
	Tx       std.Tx         `json:"tx"`
	Metadata *GnoTxMetadata `json:"metadata,omitempty"`
}

type GnoTxMetadata struct {
	Timestamp int64 `json:"timestamp"`
}

// ReadGenesisTxs reads the genesis txs from the given file path
func ReadGenesisTxs(ctx context.Context, path string) ([]TxWithMetadata, error) {
	// Open the txs file
	file, loadErr := os.Open(path)
	if loadErr != nil {
		return nil, fmt.Errorf("unable to open tx file %s: %w", path, loadErr)
	}
	defer file.Close()

	var (
		txs []TxWithMetadata

		scanner = bufio.NewScanner(file)
	)

	scanner.Buffer(make([]byte, 1_000_000), 2_000_000)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Parse the amino JSON
			var tx TxWithMetadata
			if err := amino.UnmarshalJSON(scanner.Bytes(), &tx); err != nil {
				return nil, fmt.Errorf(
					"unable to unmarshal amino JSON, %w",
					err,
				)
			}

			txs = append(txs, tx)
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(
			"error encountered while reading file, %w",
			err,
		)
	}

	return txs, nil
}

// GenesisSigner defines the interface needed to sign genesis transactions.
// Both crypto.PrivKey and bft/types.Signer implement this interface.
type GenesisSigner interface {
	PubKey() crypto.PubKey
	Sign(msg []byte) ([]byte, error)
}

// SignGenesisTxs will sign all txs passed as argument using the genesis signer.
// This signature is only valid for genesis transactions as the account number and sequence are 0
func SignGenesisTxs(txs []TxWithMetadata, signer GenesisSigner, chainID string) error {
	for index, tx := range txs {
		// Upon verifying genesis transactions, the account number and sequence are considered to be 0.
		// The reason for this is that it is not possible to know the account number (or sequence!) in advance
		// when generating the genesis transaction signature
		bytes, err := tx.Tx.GetSignBytes(chainID, 0, 0)
		if err != nil {
			return fmt.Errorf("unable to get sign bytes for transaction, %w", err)
		}

		signature, err := signer.Sign(bytes)
		if err != nil {
			return fmt.Errorf("unable to sign genesis transaction, %w", err)
		}

		txs[index].Tx.Signatures = []std.Signature{
			{
				PubKey:    signer.PubKey(),
				Signature: signature,
			},
		}
	}

	return nil
}
