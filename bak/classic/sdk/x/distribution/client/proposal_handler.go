package client

import (
	"github.com/tendermint/classic/sdk/x/distribution/client/cli"
	"github.com/tendermint/classic/sdk/x/distribution/client/rest"
	govclient "github.com/tendermint/classic/sdk/x/gov/client"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
