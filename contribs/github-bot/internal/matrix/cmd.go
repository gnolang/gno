package matrix

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func NewMatrixCmd(verbose bool) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "matrix",
			ShortUsage: "github-bot matrix [flags]",
			ShortHelp:  "parses GitHub Actions event and defines matrix accordingly",
			LongHelp:   "This tool retrieves the GitHub Actions context, parses the attached event, and defines the matrix with the pull request numbers to be processed accordingly",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, _ []string) error {
			return execMatrix()
		},
	)
}
