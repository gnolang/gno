package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/peterbourgon/ff/v3"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// Config defines the command config interface
// that holds flag values and execution logic
type Config interface {
	// RegisterFlags registers the specific flags to the flagset
	RegisterFlags(*flag.FlagSet)
}

// ExecMethod executes the command using the specified config
type ExecMethod func(ctx context.Context, args []string) error

// HelpExec is a standard exec method for displaying
// help information about a command
func HelpExec(_ context.Context, _ []string) error {
	return flag.ErrHelp
}

// Metadata contains basic help
// information about a command
type Metadata struct {
	Name          string
	ShortUsage    string
	ShortHelp     string
	LongHelp      string
	Options       []ff.Option
	NoParentFlags bool
}

// Command is a simple wrapper for gnoland commands.
type Command struct {
	name          string
	shortUsage    string
	shortHelp     string
	longHelp      string
	options       []ff.Option
	cfg           Config
	flagSet       *flag.FlagSet
	subcommands   []*Command
	exec          ExecMethod
	selected      *Command
	args          []string
	noParentFlags bool
}

func NewCommand(
	meta Metadata,
	config Config,
	exec ExecMethod,
) *Command {
	command := &Command{
		name:          meta.Name,
		shortUsage:    meta.ShortUsage,
		shortHelp:     meta.ShortHelp,
		longHelp:      meta.LongHelp,
		options:       meta.Options,
		noParentFlags: meta.NoParentFlags,
		flagSet:       flag.NewFlagSet(meta.Name, flag.ContinueOnError),
		exec:          exec,
		cfg:           config,
	}

	if config != nil {
		// Register the base command flags
		config.RegisterFlags(command.flagSet)
	}

	return command
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, [os.Stderr] is used.
func (c *Command) SetOutput(output io.Writer) {
	c.flagSet.SetOutput(output)
}

// AddSubCommands adds a variable number of subcommands
// and registers common flags using the flagset
func (c *Command) AddSubCommands(cmds ...*Command) {
	for _, cmd := range cmds {
		if c.cfg != nil && !cmd.noParentFlags {
			// Register the parent flagset with the child.
			// The syntax is not intuitive, but the flagset being
			// modified is the subcommand's, using the flags defined
			// in the parent command
			c.cfg.RegisterFlags(cmd.flagSet)

			// Register the parent flagset with all the
			// subcommands of the child as well
			// (ex. grandparent flags are available in child commands)
			registerFlagsWithSubcommands(c.cfg, cmd)

			// Register the parent options with the child.
			cmd.options = append(cmd.options, c.options...)

			// Register the parent options with all the
			// subcommands of the child as well
			registerOptionsWithSubcommands(cmd)
		}

		// Append the subcommand to the parent
		c.subcommands = append(c.subcommands, cmd)
	}
}

// Execute is a helper function for command entry. It wraps ParseAndRun and
// handles the flag.ErrHelp error, ensuring that every command with -h or
// --help won't show an error message:
// 'error parsing commandline arguments: flag: help requested'
//
// Additionally, any error of type [ErrExitCode] will be handled by exiting with
// the given status code.
func (c *Command) Execute(ctx context.Context, args []string) {
	// test moonia
	defer func() {
		var m gno.Machine
        m.PrintOpStats()
	}()
	// test moonia
	if err := c.ParseAndRun(ctx, args); err != nil {
		var ece ExitCodeError
		switch {
		case errors.Is(err, flag.ErrHelp): // just exit with 1 (help already printed)
		case errors.As(err, &ece):
			os.Exit(int(ece))
		default:
			_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		}
		os.Exit(1)
	}
}

// ParseAndRun is a helper function that calls Parse and then Run in a single
// invocation. It's useful for simple command trees that don't need two-phase
// setup.
//
// Forked from peterbourgon/ff/ffcli
func (c *Command) ParseAndRun(ctx context.Context, args []string) error {
	if err := c.Parse(args); err != nil {
		return err
	}

	if err := c.Run(ctx); err != nil {
		return err
	}

	return nil
}

// Parse the commandline arguments for this command and all sub-commands
// recursively, defining flags along the way. If Parse returns without an error,
// the terminal command has been successfully identified, and may be invoked by
// calling Run.
//
// If the terminal command identified by Parse doesn't define an Exec function,
// then Parse will return NoExecError.
//
// Forked from peterbourgon/ff/ffcli
func (c *Command) Parse(args []string) error {
	if c.selected != nil {
		return nil
	}

	if c.flagSet == nil {
		c.flagSet = flag.NewFlagSet(c.name, flag.ExitOnError)
	}

	c.flagSet.Usage = func() {
		fmt.Fprintln(c.flagSet.Output(), usage(c))
	}

	c.args = []string{}
	// Use a loop to support flag declaration after arguments and subcommands.
	// At the end of each iteration:
	// - c.args receives the first argument found
	// - args is truncated by anything that has been parsed
	// The loop ends whether if:
	// - no more arguments to parse
	// - a double delimiter "--" is met
	for {
		// ff.Parse iterates over args, feeding FlagSet with the flags encountered.
		// It stops when:
		// 1) there's nothing more to parse. In that case, FlagSet.Args() is empty.
		// 2) it encounters a double delimiter "--". In that case FlagSet.Args()
		// contains everything that follows the double delimiter.
		// 3) it encounters an item that is not a flag. In that case FlagSet.Args()
		// contains that last item and everything that follows it. The item can be
		// an argument or a subcommand.
		if err := ff.Parse(c.flagSet, args, c.options...); err != nil {
			return err
		}
		if c.flagSet.NArg() == 0 {
			// 1) Nothing more to parse
			break
		}
		// Determine if ff.Parse() has been interrupted by a double delimiter.
		// This is case if the last parsed arg is a "--"
		parsedArgs := args[:len(args)-c.flagSet.NArg()]
		if n := len(parsedArgs); n > 0 && parsedArgs[n-1] == "--" {
			// 2) Double delimiter has been met, everything that follow it can be
			// considered as arguments.
			c.args = append(c.args, c.flagSet.Args()...)
			break
		}
		// 3) c.FlagSet.Arg(0) is not a flag, determine if it's an argument or a
		// subcommand.
		// NOTE: it can be a subcommand if and only if the argument list is empty.
		// In other words, a command can't have both arguments and subcommands.
		if len(c.args) == 0 {
			for _, subcommand := range c.subcommands {
				if strings.EqualFold(c.flagSet.Arg(0), subcommand.name) {
					// c.FlagSet.Arg(0) is a subcommand
					c.selected = subcommand
					return subcommand.Parse(c.flagSet.Args()[1:])
				}
			}
		}
		// c.FlagSet.Arg(0) is an argument, append it to the argument list
		c.args = append(c.args, c.flagSet.Arg(0))
		// Truncate args and continue
		args = c.flagSet.Args()[1:]
	}

	c.selected = c

	if c.exec == nil {
		return fmt.Errorf("command %s not executable", c.name)
	}

	return nil
}

// Run selects the terminal command in a command tree previously identified by a
// successful call to Parse, and calls that command's Exec function with the
// appropriate subset of commandline args.
//
// If the terminal command previously identified by Parse doesn't define an Exec
// function, then Run will return an error.
//
// Forked from peterbourgon/ff/ffcli
func (c *Command) Run(ctx context.Context) (err error) {
	var (
		unparsed = c.selected == nil
		terminal = c.selected == c && c.exec != nil
		noop     = c.selected == c && c.exec == nil
	)

	defer func() {
		if terminal && errors.Is(err, flag.ErrHelp) {
			c.flagSet.Usage()
		}
	}()

	switch {
	case unparsed:
		return fmt.Errorf("command %s not parsed", c.name)
	case terminal:
		return c.exec(ctx, c.args)
	case noop:
		return fmt.Errorf("command %s not executable", c.name)
	default:
		return c.selected.Run(ctx)
	}
}

// Forked from peterbourgon/ff/ffcli
func usage(c *Command) string {
	var b strings.Builder

	fmt.Fprintf(&b, "USAGE\n")
	if c.shortUsage != "" {
		fmt.Fprintf(&b, "  %s\n", c.shortUsage)
	} else {
		fmt.Fprintf(&b, "  %s\n", c.name)
	}
	fmt.Fprintf(&b, "\n")

	if c.longHelp != "" {
		fmt.Fprintf(&b, "%s\n\n", c.longHelp)
	} else if c.shortHelp != "" {
		fmt.Fprintf(&b, "%s.\n\n", c.shortHelp)
	}

	if len(c.subcommands) > 0 {
		fmt.Fprintf(&b, "SUBCOMMANDS\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		for _, subcommand := range c.subcommands {
			fmt.Fprintf(tw, "  %s\t%s\n", subcommand.name, subcommand.shortHelp)
		}
		tw.Flush()
		fmt.Fprintf(&b, "\n")
	}

	if countFlags(c.flagSet) > 0 {
		fmt.Fprintf(&b, "FLAGS\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		c.flagSet.VisitAll(func(f *flag.Flag) {
			space := " "
			if isBoolFlag(f) {
				space = "="
			}

			def := f.DefValue
			if def == "" {
				def = "..."
			}

			fmt.Fprintf(tw, "  -%s%s%s\t%s\n", f.Name, space, def, f.Usage)
		})
		tw.Flush()
		fmt.Fprintf(&b, "\n")
	}

	return strings.TrimSpace(b.String()) + "\n"
}

// Forked from peterbourgon/ff/ffcli
func countFlags(fs *flag.FlagSet) (n int) {
	fs.VisitAll(func(*flag.Flag) { n++ })
	return n
}

// Forked from peterbourgon/ff/ffcli
func isBoolFlag(f *flag.Flag) bool {
	b, ok := f.Value.(interface {
		IsBoolFlag() bool
	})
	return ok && b.IsBoolFlag()
}

// registerFlagsWithSubcommands recursively registers the passed in
// configuration's flagset with the subcommand tree. At the point of calling
// this method, the child subcommand tree should already be present, due to the
// way subcommands are built (LIFO)
func registerFlagsWithSubcommands(cfg Config, root *Command) {
	subcommands := []*Command{root}

	// Traverse the direct subcommand tree,
	// and register the top-level flagset with each
	// direct line subcommand
	for len(subcommands) > 0 {
		current := subcommands[0]
		subcommands = subcommands[1:]

		for _, subcommand := range current.subcommands {
			cfg.RegisterFlags(subcommand.flagSet)
			subcommands = append(subcommands, subcommand)
		}
	}
}

// registerOptionsWithSubcommands recursively registers the passed in
// options with the subcommand tree. At the point of calling
func registerOptionsWithSubcommands(root *Command) {
	subcommands := []*Command{root}

	// Traverse the direct subcommand tree,
	// and register the top-level flagset with each
	// direct line subcommand
	for len(subcommands) > 0 {
		current := subcommands[0]
		subcommands = subcommands[1:]

		for _, subcommand := range current.subcommands {
			subcommand.options = append(subcommand.options, root.options...)
			subcommands = append(subcommands, subcommand)
		}
	}
}
