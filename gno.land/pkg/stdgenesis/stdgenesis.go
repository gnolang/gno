package stdgenesis

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func EmbeddedStdlibsGenesisTxs(deployer crypto.Address, fee std.Fee) ([]gnoland.TxWithMetadata, error) {
	stdlibs, err := stdlibs.EmbeddedMemPackages()
	if err != nil {
		return nil, fmt.Errorf("unable to load embedded stdlibs: %w", err)
	}

	stdlibsTxs := []gnoland.TxWithMetadata{}
	for _, memPkg := range stdlibs {
		if memPkg.Path == "testing" {
			continue
		}

		tx := gnoland.TxWithMetadata{Tx: std.Tx{
			Fee: fee,
			Msgs: []std.Msg{
				vm.MsgAddPackage{
					Creator: deployer,
					Package: memPkg,
				},
			},
		}}

		tx.Tx.Signatures = make([]std.Signature, len(tx.Tx.GetSigners()))
		stdlibsTxs = append(stdlibsTxs, tx)
	}

	return stdlibsTxs, nil
}
