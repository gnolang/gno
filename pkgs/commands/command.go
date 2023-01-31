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

	// Exec executes the command using the specified config
	Exec(ctx context.Context, args []string) error
}

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

func (c *Command) GetConfig() Config {
	return c.cfg
}

func NewCommand(
	name string,
	meta Metadata,
	config Config,
) *Command {
	// Create the flagset
	fs := flag.NewFlagSet(name, flag.ExitOnError)

	command := &Command{
		Command: ffcli.Command{
			Name:       meta.Name,
			ShortHelp:  meta.ShortHelp,
			LongHelp:   meta.LongHelp,
			ShortUsage: meta.ShortUsage,
			Options:    meta.Options,
			FlagSet:    fs,
			Exec:       config.Exec,
		},
		cfg: config,
	}

	// Register the base command flags
	config.RegisterFlags(command.FlagSet)

	return command
}

// AddSubCommands adds a variable number of subcommands
// and registers common flags using the flagset
func (c *Command) AddSubCommands(cmds ...*Command) {
	for _, cmd := range cmds {
		// Register the parent flagset
		c.cfg.RegisterFlags(cmd.FlagSet)

		// Append the subcommand to the parent
		c.Subcommands = append(c.Subcommands, &cmd.Command)
	}
}
