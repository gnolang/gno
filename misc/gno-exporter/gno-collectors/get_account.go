package gnoexporter

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func getAccount(client rpcClient.Client, address string) (std.Account, error) {
	addr, err := crypto.AddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address: %s, %w", address, err)
	}

	path := fmt.Sprintf("auth/accounts/%s", addr.String())

	queryResponse, err := client.ABCIQuery(path, []byte{})
	if err != nil {
		return nil, fmt.Errorf("unable to execute ABCI query, %w", err)
	}

	var queryData struct{ BaseAccount std.BaseAccount }

	if err := amino.UnmarshalJSON(queryResponse.Response.Data, &queryData); err != nil {
		return nil, err
	}

	return &queryData.BaseAccount, nil
}
