/*
Package server is used to start a new ABCI server.

It contains two server implementation:
 * socket server
 * no other server is yet implemented.

*/

package server

import (
	"fmt"

	"github.com/tendermint/classic/abci/types"
	cmn "github.com/tendermint/classic/libs/common"
)

func NewServer(protoAddr, transport string, app abci.Application) (cmn.Service, error) {
	var s cmn.Service
	var err error
	switch transport {
	case "socket":
		s = NewSocketServer(protoAddr, app)
	default:
		err = fmt.Errorf("Unknown server type %s", transport)
	}
	return s, err
}
