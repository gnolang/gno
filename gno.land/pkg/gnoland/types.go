package gnoland

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	ErrBalanceEmptyAddress = errors.New("balance address is empty")
	ErrBalanceEmptyAmount  = errors.New("balance amount is empty")
)

type GnoAccount struct {
	std.BaseAccount
}

func ProtoGnoAccount() std.Account {
	return &GnoAccount{}
}

type GnoGenesisState struct {
	Balances []Balance         `json:"balances"`
	Txs      []TxWithMetadata  `json:"txs"`
	Params   []Param           `json:"params"`
	Auth     auth.GenesisState `json:"auth"`
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

// SignGenesisTxs will sign all txs passed as argument using the private key.
// This signature is only valid for genesis transactions as the account number and sequence are 0
func SignGenesisTxs(txs []TxWithMetadata, privKey crypto.PrivKey, chainID string) error {
	for index, tx := range txs {
		// Upon verifying genesis transactions, the account number and sequence are considered to be 0.
		// The reason for this is that it is not possible to know the account number (or sequence!) in advance
		// when generating the genesis transaction signature
		bytes, err := tx.Tx.GetSignBytes(chainID, 0, 0)
		if err != nil {
			return fmt.Errorf("unable to get sign bytes for transaction, %w", err)
		}

		signature, err := privKey.Sign(bytes)
		if err != nil {
			return fmt.Errorf("unable to sign genesis transaction, %w", err)
		}

		txs[index].Tx.Signatures = []std.Signature{
			{
				PubKey:    privKey.PubKey(),
				Signature: signature,
			},
		}
	}

	return nil
}
