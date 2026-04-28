package abci_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

func TestCodecParity_ABCI(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(abci.Package)
	cdc.Seal()

	cases := []struct {
		name string
		v    any
	}{
		{"RequestInfo", &abci.RequestInfo{}},
		{"ResponseInfo", &abci.ResponseInfo{
			ABCIVersion:      "0.34",
			AppVersion:       "1.0.0",
			LastBlockHeight:  42,
			LastBlockAppHash: []byte{0xde, 0xad, 0xbe, 0xef},
		}},
		{"RequestQuery", &abci.RequestQuery{
			Data:   []byte("key"),
			Path:   "/store/bank/key",
			Height: 100,
			Prove:  true,
		}},
		{"ResponseQuery", &abci.ResponseQuery{
			Key:    []byte("key"),
			Value:  []byte("value"),
			Height: 100,
		}},
		{"ResponseCheckTx", &abci.ResponseCheckTx{
			GasWanted: 10000,
			GasUsed:   5000,
		}},
		{"ConsensusParams", &abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxTxBytes:    1024 * 1024,
				MaxDataBytes:  10 * 1024 * 1024,
				MaxBlockBytes: 20 * 1024 * 1024,
				MaxGas:        -1,
				TimeIotaMS:    1000,
			},
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{"/tm.PubKeyEd25519"}},
		}},
		{"StringError", func() *abci.StringError { e := abci.StringError("boom"); return &e }()},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
