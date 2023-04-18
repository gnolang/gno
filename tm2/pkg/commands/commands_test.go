package commands

import (
	"context"
	"flag"
	"testing"

	"github.com/jaekwon/testify/require"
	"github.com/stretchr/testify/assert"

	"github.com/gnolang/gno/tm2/pkg/commands/ffcli"
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

func TestCommandFlagsOrder(t *testing.T) {
	type flags struct {
		b bool
		s string
		x bool
	}
	tests := []struct {
		name          string
		osArgs        []string
		expectedArgs  []string
		expectedFlags flags
		expectedError string
	}{
		{
			name:          "no args no flags",
			osArgs:        []string{},
			expectedArgs:  []string{},
			expectedFlags: flags{},
		},
		{
			name:          "only args",
			osArgs:        []string{"bar", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{},
		},
		{
			name:          "only flags",
			osArgs:        []string{"-b", "-s", "str"},
			expectedArgs:  []string{},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "unknow flag",
			osArgs:        []string{"-y", "-s", "str"},
			expectedArgs:  []string{},
			expectedError: "error parsing commandline arguments: flag provided but not defined: -y",
		},
		{
			name:          "flags before args",
			osArgs:        []string{"-b", "-s", "str", "bar", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "flags after args",
			osArgs:        []string{"bar", "baz", "-b", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "flags around args",
			osArgs:        []string{"-b", "bar", "baz", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "flags between args",
			osArgs:        []string{"bar", "-b", "-s", "str", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "subcommand no flags no args",
			osArgs:        []string{"sub"},
			expectedArgs:  []string{},
			expectedFlags: flags{},
		},
		{
			name:          "subcommand only args",
			osArgs:        []string{"sub", "bar", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{},
		},
		{
			name:          "subcommand flag before subcommand",
			osArgs:        []string{"-x", "sub"},
			expectedError: "error parsing commandline arguments: flag provided but not defined: -x",
		},
		{
			name:          "subcommand only flags",
			osArgs:        []string{"-b", "sub", "-x", "-s", "str"},
			expectedArgs:  []string{},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
		{
			name:          "subcommand flags before args",
			osArgs:        []string{"-b", "sub", "-x", "-s", "str", "bar", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
		{
			name:          "subcommand flags after args",
			osArgs:        []string{"-b", "sub", "bar", "baz", "-x", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
		{
			name:          "subcommand flags around args",
			osArgs:        []string{"-b", "sub", "-x", "bar", "baz", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
		{
			name:          "subcommand flags between args",
			osArgs:        []string{"-b", "sub", "bar", "-x", "baz", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				args  []string
				flags flags
			)
			// Create a cmd main that takes 2 flags -b and -s
			cmd := NewCommand(
				Metadata{Name: "main"},
				&mockConfig{
					configFn: func(fs *flag.FlagSet) {
						fs.BoolVar(&flags.b, "b", false, "a boolan")
						fs.StringVar(&flags.s, "s", "", "a string")
					},
				},
				func(_ context.Context, a []string) error {
					args = a
					return nil
				},
			)
			// Add a sub command to cmd with a single flag -x
			cmd.AddSubCommands(
				NewCommand(
					Metadata{Name: "sub"},
					&mockConfig{
						configFn: func(fs *flag.FlagSet) {
							fs.BoolVar(&flags.x, "x", false, "a boolan")
						},
					},
					func(_ context.Context, a []string) error {
						args = a
						return nil
					},
				),
			)

			err := cmd.ParseAndRun(context.Background(), tt.osArgs)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, args, tt.expectedArgs, "wrong args")
			require.Equal(t, flags, tt.expectedFlags, "wrong flags")
		})
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
