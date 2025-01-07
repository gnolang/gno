package integration

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func SignTxs(txs []gnoland.TxWithMetadata, chainID string) error {
	kb := keys.NewInMemory()
	_, err := kb.CreateAccount(DefaultAccount_Name, DefaultAccount_Seed, "", "", 0, 0)
	if err != nil {
		return err
	}
	for index, tx := range txs {
		bytes, err := tx.Tx.GetSignBytes(chainID, 0, 0)
		if err != nil {
			return err
		}
		signature, publicKey, err := kb.Sign(DefaultAccount_Name, "", bytes)
		if err != nil {
			return err
		}
		txs[index].Tx.Signatures = []std.Signature{
			{
				PubKey:    publicKey,
				Signature: signature,
			},
		}
	}
	return nil
}
