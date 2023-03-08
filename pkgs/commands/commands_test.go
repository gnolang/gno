package commands

import (
	"flag"
	"testing"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/stretchr/testify/assert"
)

type configDelegate func(*flag.FlagSet)

type mockConfig struct {
	configFn configDelegate
}

func (c *mockConfig) RegisterFlags(fs *flag.FlagSet) {
	if c.configFn != nil {
		c.configFn(fs)
	}
}

func TestCommand_AddSubCommands(t *testing.T) {
	t.Parallel()

	// Test setup //

	type testCmd struct {
		cmd     *Command
		subCmds []*testCmd
	}

	getSubcommands := func(t *testCmd) []*Command {
		res := make([]*Command, len(t.subCmds))

		for i, subCmd := range t.subCmds {
			res[i] = subCmd.cmd
		}

		return res
	}

	generateTestCmd := func(name string) *Command {
		return NewCommand(
			Metadata{
				Name: name,
			},
			&mockConfig{
				func(fs *flag.FlagSet) {
					fs.String(
						name,
						"",
						"",
					)
				},
			},
			HelpExec,
		)
	}

	var postorderCommands func(root *testCmd) []*testCmd

	postorderCommands = func(root *testCmd) []*testCmd {
		if root == nil {
			return nil
		}

		res := make([]*testCmd, 0)

		for _, child := range root.subCmds {
			res = append(res, postorderCommands(child)...)
		}

		return append(res, root)
	}

	// Cases //

	testTable := []struct {
		name   string
		topCmd *testCmd
	}{
		{
			name: "no subcommands",
			topCmd: &testCmd{
				cmd:     generateTestCmd("level0"),
				subCmds: nil,
			},
		},
		{
			name: "single subcommand level",
			topCmd: &testCmd{
				cmd: generateTestCmd("level0"),
				subCmds: []*testCmd{
					{
						cmd:     generateTestCmd("level1"),
						subCmds: nil,
					},
				},
			},
		},
		{
			name: "multiple subcommand levels",
			topCmd: &testCmd{
				cmd: generateTestCmd("level0"),
				subCmds: []*testCmd{
					{
						cmd: generateTestCmd("level1"),
						subCmds: []*testCmd{
							{
								cmd:     generateTestCmd("level2"),
								subCmds: nil,
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var validateSubcommandTree func(flag string, root *ffcli.Command)

			validateSubcommandTree = func(flag string, root *ffcli.Command) {
				assert.NotNil(t, root.FlagSet.Lookup(flag))

				for _, subcommand := range root.Subcommands {
					validateSubcommandTree(flag, subcommand)
				}
			}

			// Register the subcommands in LIFO order (postorder), starting from the
			// leaf of the command tree (mimics how the commands package is used)
			commandOrder := postorderCommands(testCase.topCmd)

			for _, currCmd := range commandOrder {
				// For the current command, register its subcommand tree
				currCmd.cmd.AddSubCommands(getSubcommands(currCmd)...)

				// Validate that the entire subcommand tree has root command flags
				for _, subCmd := range currCmd.cmd.Subcommands {
					// For each root command flag, validate
					currCmd.cmd.FlagSet.VisitAll(func(f *flag.Flag) {
						validateSubcommandTree(f.Name, subCmd)
					})
				}
			}
		})
	}
}
