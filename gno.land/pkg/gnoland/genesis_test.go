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
