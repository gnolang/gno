package example

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func createGenesisTxsFile(outputPath string, privKey secp256k1.PrivKeySecp256k1, chainID string) error {
	var txs []gnoland.TxWithMetadata

	// Create a MsgCall transaction
	caller := privKey.PubKey().Address()
	callTx := gnoland.TxWithMetadata{
		Tx: std.Tx{
			Msgs: []std.Msg{
				vm.MsgCall{
					Caller:  caller,
					PkgPath: "gno.land/r/demo/users",
					Func:    "Register",
					Args:    []string{"myusername"},
					Send:    std.Coins{},
				},
			},
			Fee: std.NewFee(2000000, std.MustParseCoin("1000000ugnot")),
		},
		Metadata: &gnoland.GnoTxMetadata{
			Timestamp: time.Now().Unix(),
		},
	}
	txs = append(txs, callTx)

	// Sort transactions by timestamp (required for deterministic ordering)
	slices.SortStableFunc(txs, func(a, b gnoland.TxWithMetadata) int {
		if a.Metadata == nil || b.Metadata == nil {
			return 0
		}
		return cmp.Compare(a.Metadata.Timestamp, b.Metadata.Timestamp)
	})

	// Sign transactions
	if err := gnoland.SignGenesisTxs(txs, privKey, chainID); err != nil {
		return err
	}

	// Write transactions to JSONL file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, tx := range txs {
		encoded, err := amino.MarshalJSON(tx)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(file, "%s\n", encoded); err != nil {
			return err
		}
	}

	return nil
}
