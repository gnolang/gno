package gnoland

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Verify(t *testing.T) {
	tests := []struct {
		name      string
		genesis   GnoGenesisState
		expectErr bool
	}{
		{"default GenesisState", DefaultGenState(), false},
		{
			"invalid GenesisState Auth",
			GnoGenesisState{
				Auth: auth.GenesisState{},
				Bank: bank.DefaultGenesisState(),
				VM:   vmm.DefaultGenesisState(),
			},
			true,
		},
		{
			"invalid GenesisState Bank",
			GnoGenesisState{
				Auth: auth.DefaultGenesisState(),
				Bank: bank.GenesisState{
					Params: bank.Params{
						RestrictedDenoms: []string{"INVALID!!!"},
					},
				},
				VM: vmm.DefaultGenesisState(),
			},
			true,
		},
		{
			"invalid GenesisState VM",
			GnoGenesisState{
				Auth: auth.DefaultGenesisState(),
				Bank: bank.DefaultGenesisState(),
				VM:   vmm.GenesisState{},
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGenState(tc.genesis)
			if tc.expectErr {
				assert.Error(t, err, fmt.Sprintf("TestGenesis_Verify: %s", tc.name))
			} else {
				assert.NoError(t, err, fmt.Sprintf("TestGenesis_Verify: %s", tc.name))
			}
		})
	}
}

func TestLoadPackagesFromDir_Creator(t *testing.T) {
	defaultCreator := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	customCreator := crypto.MustAddressFromString("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

	tests := []struct {
		name     string
		packages []struct {
			dir      string
			gnomod   string
			goFile   string
			fileName string
		}
		expectError   bool
		errorContains string
		verify        func(t *testing.T, txs []TxWithMetadata)
	}{
		{
			name: "package without creator uses default",
			packages: []struct {
				dir      string
				gnomod   string
				goFile   string
				fileName string
			}{
				{
					dir: "pkg1",
					gnomod: `module = "gno.land/p/test/pkg1"
gno = "0.9"
`,
					goFile: `package pkg1

func Hello() string {
	return "Hello from pkg1"
}
`,
					fileName: "pkg1.gno",
				},
			},
			verify: func(t *testing.T, txs []TxWithMetadata) {
				t.Helper()
				require.Len(t, txs, 1)
				msg, ok := txs[0].Tx.Msgs[0].(vmm.MsgAddPackage)
				require.True(t, ok)
				assert.Equal(t, defaultCreator, msg.Creator)
				assert.Equal(t, "gno.land/p/test/pkg1", msg.Package.Path)
			},
		},
		{
			name: "package with creator uses custom address",
			packages: []struct {
				dir      string
				gnomod   string
				goFile   string
				fileName string
			}{
				{
					dir: "pkg2",
					gnomod: `module = "gno.land/p/test/pkg2"
gno = "0.9"

[addpkg]
creator = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
`,
					goFile: `package pkg2

func World() string {
	return "World from pkg2"
}
`,
					fileName: "pkg2.gno",
				},
			},
			verify: func(t *testing.T, txs []TxWithMetadata) {
				t.Helper()
				require.Len(t, txs, 1)
				msg, ok := txs[0].Tx.Msgs[0].(vmm.MsgAddPackage)
				require.True(t, ok)
				assert.Equal(t, customCreator, msg.Creator)
				assert.Equal(t, "gno.land/p/test/pkg2", msg.Package.Path)
			},
		},
		{
			name: "mixed packages with and without creator",
			packages: []struct {
				dir      string
				gnomod   string
				goFile   string
				fileName string
			}{
				{
					dir: "pkg1",
					gnomod: `module = "gno.land/p/test/pkg1"
gno = "0.9"
`,
					goFile: `package pkg1

func Hello() string {
	return "Hello"
}
`,
					fileName: "pkg1.gno",
				},
				{
					dir: "pkg2",
					gnomod: `module = "gno.land/p/test/pkg2"
gno = "0.9"

[addpkg]
creator = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
`,
					goFile: `package pkg2

func World() string {
	return "World"
}
`,
					fileName: "pkg2.gno",
				},
			},
			verify: func(t *testing.T, txs []TxWithMetadata) {
				t.Helper()
				require.Len(t, txs, 2)
				creators := make(map[string]crypto.Address)
				for _, tx := range txs {
					msg, ok := tx.Tx.Msgs[0].(vmm.MsgAddPackage)
					require.True(t, ok)
					creators[msg.Package.Path] = msg.Creator
				}
				assert.Equal(t, defaultCreator, creators["gno.land/p/test/pkg1"])
				assert.Equal(t, customCreator, creators["gno.land/p/test/pkg2"])
			},
		},
		{
			name: "invalid creator address",
			packages: []struct {
				dir      string
				gnomod   string
				goFile   string
				fileName string
			}{
				{
					dir: "pkg",
					gnomod: `module = "gno.land/p/test/pkg"
gno = "0.9"

[addpkg]
creator = "invalid_address"
`,
					goFile: `package pkg

func Test() string {
	return "test"
}
`,
					fileName: "pkg.gno",
				},
			},
			expectError:   true,
			errorContains: "invalid creator address",
		},
		{
			name: "empty creator address uses default",
			packages: []struct {
				dir      string
				gnomod   string
				goFile   string
				fileName string
			}{
				{
					dir: "pkg",
					gnomod: `module = "gno.land/p/test/pkg"
gno = "0.9"

[addpkg]
creator = ""
`,
					goFile: `package pkg

func Test() string {
	return "test"
}
`,
					fileName: "pkg.gno",
				},
			},
			verify: func(t *testing.T, txs []TxWithMetadata) {
				t.Helper()
				require.Len(t, txs, 1)
				msg, ok := txs[0].Tx.Msgs[0].(vmm.MsgAddPackage)
				require.True(t, ok)
				assert.Equal(t, defaultCreator, msg.Creator)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create test packages
			for _, pkg := range tc.packages {
				pkgDir := filepath.Join(tempDir, pkg.dir)
				require.NoError(t, os.MkdirAll(pkgDir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "gnomod.toml"), []byte(pkg.gnomod), 0644))
				require.NoError(t, os.WriteFile(filepath.Join(pkgDir, pkg.fileName), []byte(pkg.goFile), 0644))
			}

			// Load packages
			txs, err := LoadPackagesFromDir(tempDir, defaultCreator, std.Fee{})

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				if tc.verify != nil {
					tc.verify(t, txs)
				}
			}
		})
	}
}

func TestLoadPackagesFromDir_Realm(t *testing.T) {
	defaultCreator := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	customCreator := crypto.MustAddressFromString("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

	tests := []struct {
		name     string
		packages []struct {
			dir      string
			gnomod   string
			goFile   string
			fileName string
		}
		verify func(t *testing.T, txs []TxWithMetadata)
	}{
		{
			name: "realm with creator",
			packages: []struct {
				dir      string
				gnomod   string
				goFile   string
				fileName string
			}{
				{
					dir: "creatortest",
					gnomod: `module = "gno.land/r/test/creatortest"
gno = "0.9"

[addpkg]
creator = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
`,
					goFile: `package creatortest

import "std"

var realmCreator string

func init() {
	realmCreator = string(std.GetOrigCaller())
}

func GetCreator() string {
	return realmCreator
}

func Render(path string) string {
	return "Realm creator: " + realmCreator
}
`,
					fileName: "creatortest.gno",
				},
			},
			verify: func(t *testing.T, txs []TxWithMetadata) {
				t.Helper()
				require.Len(t, txs, 1)
				msg, ok := txs[0].Tx.Msgs[0].(vmm.MsgAddPackage)
				require.True(t, ok)
				assert.Equal(t, customCreator, msg.Creator)
				assert.Equal(t, "gno.land/r/test/creatortest", msg.Package.Path)
				assert.Equal(t, "creatortest", msg.Package.Name)

				// Find .gno file
				var gnoFile *std.MemFile
				for _, file := range msg.Package.Files {
					if filepath.Ext(file.Name) == ".gno" {
						gnoFile = file
						break
					}
				}
				require.NotNil(t, gnoFile)
				assert.Equal(t, "creatortest.gno", gnoFile.Name)
			},
		},
		{
			name: "multiple packages with different creators",
			packages: []struct {
				dir      string
				gnomod   string
				goFile   string
				fileName string
			}{
				{
					dir: "pkg1",
					gnomod: `module = "gno.land/p/test/pkg1"
gno = "0.9"

[addpkg]
creator = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
`,
					goFile: `package pkg1

func Hello() string {
	return "Hello from pkg1"
}
`,
					fileName: "pkg1.gno",
				},
				{
					dir: "pkg2",
					gnomod: `module = "gno.land/p/test/pkg2"
gno = "0.9"

[addpkg]
creator = "g1g3lsfxhvaqgdv4ccemwpnms4fv6t3aq3p5z6u7"
`,
					goFile: `package pkg2

func World() string {
	return "World from pkg2"
}
`,
					fileName: "pkg2.gno",
				},
				{
					dir: "pkg3",
					gnomod: `module = "gno.land/p/test/pkg3"
gno = "0.9"
`,
					goFile: `package pkg3

func Test() string {
	return "Test from pkg3"
}
`,
					fileName: "pkg3.gno",
				},
			},
			verify: func(t *testing.T, txs []TxWithMetadata) {
				t.Helper()
				require.Len(t, txs, 3)

				// Build creator map
				creators := make(map[string]crypto.Address)
				for _, tx := range txs {
					msg, ok := tx.Tx.Msgs[0].(vmm.MsgAddPackage)
					require.True(t, ok)
					creators[msg.Package.Path] = msg.Creator
				}

				// Verify creators
				assert.Equal(t, customCreator, creators["gno.land/p/test/pkg1"])
				assert.Equal(t, crypto.MustAddressFromString("g1g3lsfxhvaqgdv4ccemwpnms4fv6t3aq3p5z6u7"), creators["gno.land/p/test/pkg2"])
				assert.Equal(t, defaultCreator, creators["gno.land/p/test/pkg3"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create test packages
			for _, pkg := range tc.packages {
				pkgDir := filepath.Join(tempDir, pkg.dir)
				require.NoError(t, os.MkdirAll(pkgDir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "gnomod.toml"), []byte(pkg.gnomod), 0644))
				require.NoError(t, os.WriteFile(filepath.Join(pkgDir, pkg.fileName), []byte(pkg.goFile), 0644))
			}

			// Load packages
			txs, err := LoadPackagesFromDir(tempDir, defaultCreator, std.Fee{})
			require.NoError(t, err)
			tc.verify(t, txs)
		})
	}
}

// TestLoadGenesisParamsFile_RealmSection covers the realm part of a [vm:<x>]
// section.
//
// These keys are written as "vm:"+key, and the vm's own params live at
// vm:p:<field>, so a section named [vm:p] sets those rather than a realm's --
// and gets past the [vm] section, which accepts only two named fields. The
// realm-param loop also runs after the vm params are set, so it wins.
func TestLoadGenesisParamsFile_RealmSection(t *testing.T) {
	t.Parallel()

	load := func(t *testing.T, body string) error {
		t.Helper()
		path := filepath.Join(t.TempDir(), "params.toml")
		require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
		ggs := &GnoGenesisState{VM: vmm.DefaultGenesisState()}
		return LoadGenesisParamsFile(path, ggs)
	}

	t.Run("a vm submodule is refused", func(t *testing.T) {
		t.Parallel()
		err := load(t, "[\"vm:p\"]\n  \"run_submitters.strings\" = [\"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\"]\n")
		require.Error(t, err, "[vm:p] must not be able to set a vm parameter")
		assert.Contains(t, err.Error(), "is not a realm path")
	})

	t.Run("a colon in the name is refused", func(t *testing.T) {
		t.Parallel()
		// This used to pass the loader and then panic the node at boot, which is
		// the outcome the loader guard exists to prevent. The loader had a copy
		// of the realm-path rule and neither of the other two.
		err := load(t, "[\"vm:gno.land/r/demo/foo\"]\n  \"a:b\" = \"x\"\n")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not contain a colon")
		assert.Contains(t, err.Error(), "vm:gno.land/r/demo/foo", "the error must name the section")
	})

	t.Run("a realm path is accepted and reaches RealmParams", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "params.toml")
		require.NoError(t, os.WriteFile(path,
			[]byte("[\"vm:gno.land/r/demo/foo\"]\n  bar = \"baz\"\n"), 0o644))
		ggs := &GnoGenesisState{VM: vmm.DefaultGenesisState()}
		require.NoError(t, LoadGenesisParamsFile(path, ggs))
		require.Len(t, ggs.VM.RealmParams, 1)
		assert.Equal(t, "gno.land/r/demo/foo:bar", ggs.VM.RealmParams[0].Key)
	})
}

// The balance sheet gnoland reads at start is the same file gnogenesis writes,
// so a vesting entry has to work in both. This one used to be rejected as a
// malformed line.
func TestLoadGenesisBalancesFile_Vesting(t *testing.T) {
	t.Parallel()

	const addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

	write := func(t *testing.T, body string) string {
		t.Helper()
		p := filepath.Join(t.TempDir(), "balances.txt")
		require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
		return p
	}

	t.Run("a vesting entry loads with its schedule", func(t *testing.T) {
		t.Parallel()

		bs, err := LoadGenesisBalancesFile(write(t,
			"# a comment\n\n"+addr+"=1000ugnot;vesting=1000ugnot,100,200\n"))
		require.NoError(t, err)
		require.Len(t, bs, 1)

		bal := bs[crypto.MustAddressFromString(addr)]
		require.NotNil(t, bal.Vesting, "the schedule must survive loading")
		assert.Equal(t, int64(200), bal.Vesting.EndTime)
		assert.Equal(t, int64(1000), bal.Amount.AmountOf("ugnot"))
	})

	t.Run("a delayed entry keeps its type", func(t *testing.T) {
		t.Parallel()

		bs, err := LoadGenesisBalancesFile(write(t,
			addr+"=1000ugnot;vesting=1000ugnot,0,200;type=delayed\n"))
		require.NoError(t, err)

		bal := bs[crypto.MustAddressFromString(addr)]
		require.NotNil(t, bal.Vesting)
		assert.Equal(t, std.VestingDelayed, bal.Vesting.Type)
	})

	t.Run("a plain entry still loads, and carries no schedule", func(t *testing.T) {
		t.Parallel()

		bs, err := LoadGenesisBalancesFile(write(t, addr+"=1000ugnot\n"))
		require.NoError(t, err)

		bal := bs[crypto.MustAddressFromString(addr)]
		assert.Nil(t, bal.Vesting)
		assert.Equal(t, int64(1000), bal.Amount.AmountOf("ugnot"))
	})

	t.Run("a malformed entry is still refused", func(t *testing.T) {
		t.Parallel()

		_, err := LoadGenesisBalancesFile(write(t, "not-an-entry\n"))
		require.Error(t, err)
	})
}
