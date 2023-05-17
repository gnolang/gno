package commands

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
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
	Name       string
	ShortUsage string
	ShortHelp  string
	LongHelp   string
	Options    []ff.Option
}

// Command is a simple wrapper for gnoland
// commands that utilizes ffcli
type Command struct {
	ffcli.Command

	cfg Config
}

func NewCommand(
	meta Metadata,
	config Config,
	exec ExecMethod,
) *Command {
	command := &Command{
		Command: ffcli.Command{
			Name:       meta.Name,
			ShortHelp:  meta.ShortHelp,
			LongHelp:   meta.LongHelp,
			ShortUsage: meta.ShortUsage,
			Options:    meta.Options,
			FlagSet:    flag.NewFlagSet(meta.Name, flag.ExitOnError),
			Exec:       exec,
		},
		cfg: config,
	}

	if config != nil {
		// Register the base command flags
		config.RegisterFlags(command.FlagSet)
	}

	return command
}

// AddSubCommands adds a variable number of subcommands
// and registers common flags using the flagset
func (c *Command) AddSubCommands(cmds ...*Command) {
	for _, cmd := range cmds {
		if c.cfg != nil {
			// Register the parent flagset with the child.
			// The syntax is not intuitive, but the flagset being
			// modified is the subcommand's, using the flags defined
			// in the parent command
			c.cfg.RegisterFlags(cmd.FlagSet)

			// Register the parent flagset with all the
			// subcommands of the child as well
			// (ex. grandparent flags are available in child commands)
			registerFlagsWithSubcommands(c.cfg, &cmd.Command)

			// Register the parent options with the child.
			cmd.Options = append(cmd.Options, c.Options...)

			// Register the parent options with all the
			// subcommands of the child as well
			registerOptionsWithSubcommands(&cmd.Command)
		}

		// Append the subcommand to the parent
		c.Subcommands = append(c.Subcommands, &cmd.Command)
	}
}

// registerFlagsWithSubcommands recursively registers the passed in
// configuration's flagset with the subcommand tree. At the point of calling
// this method, the child subcommand tree should already be present, due to the
// way subcommands are built (LIFO)
func registerFlagsWithSubcommands(cfg Config, root *ffcli.Command) {
	subcommands := []*ffcli.Command{root}

	// Traverse the direct subcommand tree,
	// and register the top-level flagset with each
	// direct line subcommand
	for len(subcommands) > 0 {
		current := subcommands[0]
		subcommands = subcommands[1:]

		for _, subcommand := range current.Subcommands {
			cfg.RegisterFlags(subcommand.FlagSet)
			subcommands = append(subcommands, subcommand)
		}
	}
}

// registerOptionsWithSubcommands recursively registers the passed in
// options with the subcommand tree. At the point of calling
func registerOptionsWithSubcommands(root *ffcli.Command) {
	subcommands := []*ffcli.Command{root}

	// Traverse the direct subcommand tree,
	// and register the top-level flagset with each
	// direct line subcommand
	for len(subcommands) > 0 {
		current := subcommands[0]
		subcommands = subcommands[1:]

		for _, subcommand := range current.Subcommands {
			subcommand.Options = append(subcommand.Options, root.Options...)
			subcommands = append(subcommands, subcommand)
		}
	}

}
