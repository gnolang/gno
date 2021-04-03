package client

import (
	govclient "github.com/tendermint/classic/sdk/x/gov/client"
	"github.com/tendermint/classic/sdk/x/params/client/cli"
	"github.com/tendermint/classic/sdk/x/params/client/rest"
)

// param change proposal handler
var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
