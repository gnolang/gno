package rpc

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ctypes "github.com/tendermint/classic/rpc/core/types"
	"github.com/tendermint/go-amino-x"

	"github.com/tendermint/classic/sdk/client/context"
	"github.com/tendermint/classic/sdk/client/flags"
	"github.com/tendermint/classic/sdk/types/rest"
	"github.com/tendermint/classic/sdk/version"

	"github.com/tendermint/classic/p2p"
)

// StatusCommand returns the command to return the status of the network.
func StatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE:  printNodeStatus,
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	cmd.Flags().Bool(flags.FlagIndentResponse, false, "Add indent to JSON response")
	return cmd
}

func getNodeStatus(cliCtx context.CLIContext) (*ctypes.ResultStatus, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return &ctypes.ResultStatus{}, err
	}

	return node.Status()
}

func printNodeStatus(_ *cobra.Command, _ []string) error {
	// No need to verify proof in getting node status
	viper.Set(flags.FlagTrustNode, true)
	cliCtx := context.NewCLIContext()
	status, err := getNodeStatus(cliCtx)
	if err != nil {
		return err
	}

	var output []byte
	if cliCtx.Indent {
		output, err = amino.MarshalJSONIndent(status, "", "  ")
	} else {
		output, err = amino.MarshalJSON(status)
	}
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

// NodeInfoResponse defines a response type that contains node status and version
// information.
type NodeInfoResponse struct {
	p2p.DefaultNodeInfo `json:"node_info"`

	ApplicationVersion version.Info `json:"application_version"`
}

// REST handler for node info
func NodeInfoRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(cliCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		resp := NodeInfoResponse{
			DefaultNodeInfo:    status.NodeInfo,
			ApplicationVersion: version.NewInfo(),
		}
		rest.PostProcessResponseBare(w, cliCtx, resp)
	}
}

// SyncingResponse defines a response type that contains node syncing information.
type SyncingResponse struct {
	Syncing bool `json:"syncing"`
}

// REST handler for node syncing
func NodeSyncingRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(cliCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, SyncingResponse{Syncing: status.SyncInfo.CatchingUp})
	}
}
