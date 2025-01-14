package integration

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// SignTxs will sign all txs passed as argument using the private key
// this signature is only valid for genesis transactions as accountNumber and sequence are 0
func SignTxs(txs []gnoland.TxWithMetadata, privKey crypto.PrivKey, chainID string) error {
	for index, tx := range txs {
		bytes, err := tx.Tx.GetSignBytes(chainID, 0, 0)
		if err != nil {
			return fmt.Errorf("unable to get sign bytes for transaction, %w", err)
		}
		signature, err := privKey.Sign(bytes)
		if err != nil {
			return fmt.Errorf("unable to sign transaction, %w", err)
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
