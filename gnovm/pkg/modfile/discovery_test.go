package modfile_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/modfile"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListModules(t *testing.T) {
	// Helper type for defining expected module structure in tests
	type expectedModuleInfo struct {
		Name            string
		PkgPath         string
		Draft           bool
		DirRel          string // Expected directory, relative to t.TempDir()
		ExpectedFolders []struct {
			PathRel          string // Relative to module root
			GnoFiles         []string
			TestGnoFiles     []string
			FiletestGnoFiles []string
			OtherFiles       []string
		}
	}

	testCases := []struct {
		desc              string
		setup             dirSpec // Defines the structure to create under t.TempDir()
		searchRootRelPath string  // Relative path from t.TempDir() to start ListModules search
		expectedModules   []expectedModuleInfo
		errShouldContain  string
	}{
		{
			desc: "single module, no subdirs, few files",
			setup: dirSpec{
				Name:       "module_a", // This will be created inside tempDir
				HasGnoToml: &modfile.Modfile{PkgPath: "gno.land/p/module_a"},
				Files: []fileSpec{
					{Name: "a.gno", Content: "package a"}, // Import analysis not tested here
					{Name: "a_test.gno", Content: "package a_test"},
					{Name: "other.txt", Content: "text"},
				},
			},
			searchRootRelPath: ".", // Search from the root of TempDir
			expectedModules: []expectedModuleInfo{
				{
					Name: "gno.land/p/module_a", PkgPath: "gno.land/p/module_a", DirRel: "module_a",
					ExpectedFolders: []struct {
						PathRel          string
						GnoFiles         []string
						TestGnoFiles     []string
						FiletestGnoFiles []string
						OtherFiles       []string
					}{
						{PathRel: ".", GnoFiles: []string{"a.gno"}, TestGnoFiles: []string{"a_test.gno"}, OtherFiles: []string{"other.txt"}},
					},
				},
			},
		},
		{
			desc: "module with subdirectories",
			setup: dirSpec{
				Name:       "module_b",
				HasGnoToml: &modfile.Modfile{PkgPath: "gno.land/p/module_b", Draft: true},
				Files:      []fileSpec{{Name: "root.gno", Content: "package b"}},
				SubDirs: []dirSpec{
					{
						Name:  "sub1",
						Files: []fileSpec{{Name: "s1.gno"}, {Name: "data.json"}},
						SubDirs: []dirSpec{
							{Name: "sub2", Files: []fileSpec{{Name: "s2_test.gno"}}},
						},
					},
					{Name: "sub_empty", Files: []fileSpec{}},
				},
			},
			searchRootRelPath: ".",
			expectedModules: []expectedModuleInfo{
				{
					Name: "gno.land/p/module_b", PkgPath: "gno.land/p/module_b", Draft: true, DirRel: "module_b",
					ExpectedFolders: []struct {
						PathRel          string
						GnoFiles         []string
						TestGnoFiles     []string
						FiletestGnoFiles []string
						OtherFiles       []string
					}{
						{PathRel: ".", GnoFiles: []string{"root.gno"}},
						{PathRel: "sub1", GnoFiles: []string{"s1.gno"}, OtherFiles: []string{"data.json"}},
						{PathRel: filepath.Join("sub1", "sub2"), TestGnoFiles: []string{"s2_test.gno"}},
						{PathRel: "sub_empty", GnoFiles: []string{}, TestGnoFiles: []string{}, FiletestGnoFiles: []string{}, OtherFiles: []string{}},
					},
				},
			},
		},
		{
			desc: "nested modules", // outer_module contains inner_module
			setup: dirSpec{
				Name: "project_root", // A common root for the test setup
				SubDirs: []dirSpec{
					{
						Name:       "outer_module",
						HasGnoToml: &modfile.Modfile{PkgPath: "gno.land/p/outer"},
						Files:      []fileSpec{{Name: "outer.gno"}},
						SubDirs: []dirSpec{
							{
								Name:       "inner_module", // This is a nested module
								HasGnoToml: &modfile.Modfile{PkgPath: "gno.land/p/outer/inner"},
								Files:      []fileSpec{{Name: "inner.gno"}},
								SubDirs:    []dirSpec{{Name: "inner_sub", Files: []fileSpec{{Name: "in_sub.gno"}}}},
							},
							{
								Name:  "outer_sub", // A regular subdirectory of outer_module
								Files: []fileSpec{{Name: "out_sub.gno"}},
							},
						},
					},
				},
			},
			searchRootRelPath: "project_root", // Start search from the common root
			expectedModules: []expectedModuleInfo{
				{ // Outer module
					Name: "gno.land/p/outer", PkgPath: "gno.land/p/outer", DirRel: filepath.Join("project_root", "outer_module"),
					ExpectedFolders: []struct {
						PathRel          string
						GnoFiles         []string
						TestGnoFiles     []string
						FiletestGnoFiles []string
						OtherFiles       []string
					}{
						{PathRel: ".", GnoFiles: []string{"outer.gno"}},
						{PathRel: "outer_sub", GnoFiles: []string{"out_sub.gno"}},
						// inner_module and its subdirs are NOT listed as folders of outer_module
					},
				},
				{ // Inner module
					Name: "gno.land/p/outer/inner", PkgPath: "gno.land/p/outer/inner", DirRel: filepath.Join("project_root", "outer_module", "inner_module"),
					ExpectedFolders: []struct {
						PathRel          string
						GnoFiles         []string
						TestGnoFiles     []string
						FiletestGnoFiles []string
						OtherFiles       []string
					}{
						{PathRel: ".", GnoFiles: []string{"inner.gno"}},
						{PathRel: "inner_sub", GnoFiles: []string{"in_sub.gno"}},
					},
				},
			},
		},
		{
			desc: "no gno.toml found in search root",
			setup: dirSpec{
				Name:  "not_a_module", // Created inside tempDir
				Files: []fileSpec{{Name: "some.gno"}},
			},
			searchRootRelPath: "not_a_module", // Search starts here
			expectedModules:   nil,            // Expect empty list
		},
		{
			desc: "search root itself is a module",
			setup: dirSpec{
				Name:       ".", // Special name to use tempDir itself as the module root
				HasGnoToml: &modfile.Modfile{PkgPath: "gno.land/p/rootmod"},
				Files:      []fileSpec{{Name: "root.gno"}},
				SubDirs:    []dirSpec{{Name: "sub", Files: []fileSpec{{Name: "sub.gno"}}}},
			},
			searchRootRelPath: ".",
			expectedModules: []expectedModuleInfo{
				{
					Name: "gno.land/p/rootmod", PkgPath: "gno.land/p/rootmod", DirRel: ".",
					ExpectedFolders: []struct {
						PathRel          string
						GnoFiles         []string
						TestGnoFiles     []string
						FiletestGnoFiles []string
						OtherFiles       []string
					}{
						{PathRel: ".", GnoFiles: []string{"root.gno"}},
						{PathRel: "sub", GnoFiles: []string{"sub.gno"}},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			tempDir := t.TempDir()
			// Setup the initial directory structure under tempDir.
			// If tc.setup.Name is ".", it means the setup applies directly to tempDir.
			if tc.setup.Name == "." {
				setupTestDirStructure(t, tempDir, tc.setup)
			} else {
				// Create a root for the setup if ds.Name is not "."
				// This ensures that the setup's ds.Name is a subdirectory of tempDir.
				err := os.MkdirAll(filepath.Join(tempDir, tc.setup.Name), 0755)
				require.NoError(t, err)
				setupTestDirStructure(t, tempDir, tc.setup)
			}

			searchRoot := tempDir
			if tc.searchRootRelPath != "." && tc.searchRootRelPath != "" {
				searchRoot = filepath.Join(tempDir, tc.searchRootRelPath)
			}

			modules, err := modfile.ListModules(searchRoot)

			if tc.errShouldContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errShouldContain)
				return
			}
			require.NoError(t, err)
			require.Len(t, modules, len(tc.expectedModules), "Number of modules found mismatch")

			// Sort both actual and expected modules by Dir for consistent comparison
			sort.Slice(modules, func(i, j int) bool { return modules[i].Dir < modules[j].Dir })
			sort.Slice(tc.expectedModules, func(i, j int) bool {
				return filepath.Join(tempDir, tc.expectedModules[i].DirRel) < filepath.Join(tempDir, tc.expectedModules[j].DirRel)
			})

			for i, expectedMod := range tc.expectedModules {
				actualMod := modules[i]
				expectedModAbsDir := filepath.Join(tempDir, expectedMod.DirRel)
				if expectedMod.DirRel == "." { // Handle case where module is at searchRoot
					expectedModAbsDir = searchRoot
				}

				assert.Equal(t, expectedModAbsDir, actualMod.Dir, "Module Dir mismatch for %s", expectedMod.Name)
				assert.Equal(t, expectedMod.Name, actualMod.Name, "Module Name mismatch for %s", expectedMod.Name)
				assert.Equal(t, expectedMod.PkgPath, actualMod.Modfile.PkgPath, "Module Modfile.PkgPath mismatch for %s", expectedMod.Name)
				assert.Equal(t, expectedMod.Draft, actualMod.Modfile.Draft, "Module Modfile.Draft mismatch for %s", expectedMod.Name)

				require.Len(t, actualMod.Folders, len(expectedMod.ExpectedFolders), "Number of folders mismatch for module %s", actualMod.Name)

				// Sort actualMod.Folders by Dir for consistent comparison
				sort.Slice(actualMod.Folders, func(k, l int) bool { return actualMod.Folders[k].Dir < actualMod.Folders[l].Dir })
				// Sort expectedMod.ExpectedFolders by constructing absolute paths then comparing relative paths
				// This ensures that the comparison with assertFolderContents works correctly if expected folders are not pre-sorted.
				sort.SliceStable(expectedMod.ExpectedFolders, func(k, l int) bool {
					return expectedMod.ExpectedFolders[k].PathRel < expectedMod.ExpectedFolders[l].PathRel
				})

				for j, expectedFolder := range expectedMod.ExpectedFolders {
					// actualFolder should correspond to expectedFolder after sorting
					// In a more robust setup, we might search actualMod.Folders by expectedFolder.PathRel
					// For now, assume sorted order matches.
					if j < len(actualMod.Folders) {
						assertFolderContents(t, actualMod.Folders, actualMod.Dir, expectedFolder.PathRel,
							expectedFolder.GnoFiles, expectedFolder.TestGnoFiles, expectedFolder.FiletestGnoFiles, expectedFolder.OtherFiles)
					} else {
						t.Errorf("Missing expected folder with PathRel: %s for module %s", expectedFolder.PathRel, actualMod.Name)
					}
				}
			}
		})
	}
}

func TestSortModules(t *testing.T) {
	t.Parallel()
	type testCase struct {
		desc          string
		in            modfile.ModuleList
		expectedOrder []string // PkgPaths in expected sorted order
		shouldErr     bool
		errContain    string
	}

	// Define some modules for testing sort
	// Note: Dir, Modfile, Folders fields are not relevant for Sort logic, only Name and Imports.
	modA := &modfile.Module{Name: "gno.land/p/a", Imports: []string{"gno.land/p/b"}}
	modB := &modfile.Module{Name: "gno.land/p/b", Imports: []string{"gno.land/p/c"}}
	modC := &modfile.Module{Name: "gno.land/p/c"}
	modD := &modfile.Module{Name: "gno.land/p/d", Imports: []string{"gno.land/p/nonexistent"}} // Missing dep
	modE := &modfile.Module{Name: "gno.land/p/e", Imports: []string{"gno.land/p/f"}}
	modF := &modfile.Module{Name: "gno.land/p/f", Imports: []string{"gno.land/p/e"}} // Cycle with E

	testCases := []testCase{
		{desc: "empty list", in: modfile.ModuleList{}, expectedOrder: []string{}},
		{desc: "single module", in: modfile.ModuleList{modC}, expectedOrder: []string{"gno.land/p/c"}},
		{desc: "already sorted (c,b,a) -> (c,b,a)", in: modfile.ModuleList{modC, modB, modA}, expectedOrder: []string{"gno.land/p/c", "gno.land/p/b", "gno.land/p/a"}},
		{desc: "reverse order (a,b,c) -> (c,b,a)", in: modfile.ModuleList{modA, modB, modC}, expectedOrder: []string{"gno.land/p/c", "gno.land/p/b", "gno.land/p/a"}},
		{desc: "missing dependency", in: modfile.ModuleList{modD}, shouldErr: true, errContain: "missing dependency"},
		{desc: "cycle detected", in: modfile.ModuleList{modE, modF}, shouldErr: true, errContain: "cycle detected"},
		{
			desc: "more complex graph",
			in: modfile.ModuleList{
				{Name: "gno.land/p/x", Imports: []string{"gno.land/p/y", "gno.land/p/z"}},
				{Name: "gno.land/p/y"},
				{Name: "gno.land/p/z"},
			},
			expectedOrder: []string{"gno.land/p/y", "gno.land/p/z", "gno.land/p/x"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			sorted, err := tc.in.Sort()
			if tc.shouldErr {
				require.Error(t, err)
				if tc.errContain != "" {
					assert.Contains(t, err.Error(), tc.errContain)
				}
				return
			}
			require.NoError(t, err)
			actualOrder := make([]string, len(sorted))
			for i, m := range sorted {
				actualOrder[i] = m.Name
			}
			assert.Equal(t, tc.expectedOrder, actualOrder)
		})
	}
}

// fileSpec defines a file to be created for testing.
type fileSpec struct {
	Name    string
	Content string
}

// dirSpec defines a directory to be created, possibly with files and subdirectories.
type dirSpec struct {
	Name       string // Relative path from parent
	Files      []fileSpec
	SubDirs    []dirSpec
	HasGnoToml *modfile.Modfile // If not nil, a gno.toml with this content is created
}

// setupTestDirStructure creates a directory structure based on specs.
// baseDir is the root under which this structure is created (e.g., t.TempDir()).
func setupTestDirStructure(t *testing.T, baseDir string, ds dirSpec) {
	t.Helper()
	currentPath := baseDir
	if ds.Name != "." && ds.Name != "" { // Allow specifying root of baseDir itself
		currentPath = filepath.Join(baseDir, ds.Name)
	}

	// Ensure the directory exists, especially if it's the baseDir itself with ds.Name="."
	err := os.MkdirAll(currentPath, 0755)
	require.NoError(t, err, "Failed to create dir: %s", currentPath)

	if ds.HasGnoToml != nil {
		tomlBytes, err := toml.Marshal(*ds.HasGnoToml)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(currentPath, "gno.toml"), tomlBytes, 0644)
		require.NoError(t, err)
	}

	for _, fs := range ds.Files {
		err = os.WriteFile(filepath.Join(currentPath, fs.Name), []byte(fs.Content), 0644)
		require.NoError(t, err)
	}

	for _, subDirSpec := range ds.SubDirs {
		setupTestDirStructure(t, currentPath, subDirSpec) // Recursive call with currentPath as new base
	}
}

// assertFolderContents checks if a Folder with a specific relative path exists and has expected files.
// moduleAbsPath is the absolute path to the module's root.
// expectedRelPath is the path of the folder relative to moduleAbsPath.
func assertFolderContents(t *testing.T, folders []modfile.Folder, moduleAbsPath, expectedFolderRelPath string, expectedGnoFiles, expectedTestGnoFiles, expectedFiletestGnoFiles, expectedOtherFiles []string) {
	t.Helper()
	expectedAbsPath := filepath.Join(moduleAbsPath, expectedFolderRelPath)
	found := false
	for _, f := range folders {
		if f.Dir == expectedAbsPath {
			found = true
			// Sort slices before comparing for ElementsMatch to handle order differences from os.ReadDir
			sort.Strings(f.GnoFiles)
			sort.Strings(expectedGnoFiles)
			assert.ElementsMatch(t, expectedGnoFiles, f.GnoFiles, "GnoFiles mismatch in folder %s (abs: %s)", expectedFolderRelPath, expectedAbsPath)

			sort.Strings(f.TestGnoFiles)
			sort.Strings(expectedTestGnoFiles)
			assert.ElementsMatch(t, expectedTestGnoFiles, f.TestGnoFiles, "TestGnoFiles mismatch in folder %s (abs: %s)", expectedFolderRelPath, expectedAbsPath)

			sort.Strings(f.FiletestGnoFiles)
			sort.Strings(expectedFiletestGnoFiles)
			assert.ElementsMatch(t, expectedFiletestGnoFiles, f.FiletestGnoFiles, "FiletestGnoFiles mismatch in folder %s (abs: %s)", expectedFolderRelPath, expectedAbsPath)

			sort.Strings(f.OtherFiles)
			sort.Strings(expectedOtherFiles)
			assert.ElementsMatch(t, expectedOtherFiles, f.OtherFiles, "OtherFiles mismatch in folder %s (abs: %s)", expectedFolderRelPath, expectedAbsPath)
			break
		}
	}
	assert.True(t, found, "Folder with relative path '%s' (expected abs: '%s') not found in module folders", expectedFolderRelPath, expectedAbsPath)
}
