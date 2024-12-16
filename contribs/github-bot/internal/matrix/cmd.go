package matrix

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type matrixFlags struct {
	verbose   *bool
	matrixKey string
	flagSet   *flag.FlagSet
}

func NewMatrixCmd(verbose *bool) *commands.Command {
	flags := &matrixFlags{verbose: verbose}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "matrix",
			ShortUsage: "github-bot matrix [flags]",
			ShortHelp:  "parses GitHub Actions event and defines matrix accordingly",
			LongHelp:   "This tool retrieves the GitHub Actions context, parses the attached event, and defines the matrix with the pull request numbers to be processed accordingly",
		},
		flags,
		func(_ context.Context, _ []string) error {
			flags.validateFlags()
			return execMatrix(flags)
		},
	)
}

func (flags *matrixFlags) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&flags.matrixKey,
		"matrix-key",
		"",
		"key of the matrix to set in Github Actions output (required)",
	)

	flags.flagSet = fs
}

func (flags *matrixFlags) validateFlags() {
	if flags.matrixKey == "" {
		fmt.Fprintf(flags.flagSet.Output(), "Error: no matrix-key provided\n\n")
		flags.flagSet.Usage()
		os.Exit(1)
	}
}
