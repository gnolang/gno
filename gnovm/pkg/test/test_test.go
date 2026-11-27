package test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/require"
)

// runPackageTests runs the tests of a package made of files on opts and
// returns the gas each test reports, in source order.
func runPackageTests(t *testing.T, opts *TestOptions, files map[string]string) []int64 {
	t.Helper()
	const pkgPath = "gno.land/p/test/pkggas"
	dir := t.TempDir()
	files["gnomod.toml"] = gno.GenGnoModLatest(pkgPath)
	for name, body := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
	}
	var stderr bytes.Buffer
	opts.Error = &stderr
	opts.Verbose = true
	require.NoError(t, Test(gno.MustReadMemPackage(dir, pkgPath, gno.MPAnyAll), dir, opts))

	var gas []int64
	for _, m := range regexp.MustCompile(`--- GAS:  (\d+)`).FindAllStringSubmatch(stderr.String(), -1) {
		g, err := strconv.ParseInt(m[1], 10, 64)
		require.NoError(t, err)
		gas = append(gas, g)
	}
	return gas
}

func newPackageTestOptions(t *testing.T) *TestOptions {
	t.Helper()
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)
	return NewTestOptions(rootDir, &bytes.Buffer{}, &bytes.Buffer{}, nil)
}

// Package tests charge allocations the same gas filetests do.
func TestPackageTestGasChargesAllocations(t *testing.T) {
	opts := newPackageTestOptions(t)
	// Same length names, so the two tests differ only in what they allocate.
	gas := runPackageTests(t, opts, map[string]string{
		"pkggas.gno": "package pkggas\n\nfunc Alloc(n int) []byte { return make([]byte, n) }\n",
		"pkggas_test.gno": `package pkggas

import "testing"

func TestAllocLo(t *testing.T) { Alloc(1) }

func TestAllocHi(t *testing.T) { Alloc(10000) }
`,
	})
	require.Len(t, gas, 2)

	filetest := func(n int) int64 {
		src := "package main\n\nfunc alloc(n int) []byte { return make([]byte, n) }\n\nfunc main() { alloc(" + strconv.Itoa(n) + ") }\n"
		_, g, _, err := opts.RunFiletest("alloc.gno", []byte(src), opts.TestStore)
		require.NoError(t, err)
		return g
	}
	require.Equal(t, filetest(10000)-filetest(1), gas[1]-gas[0])
}

// A filetest loading a stdlib on the shared store must not change the gas
// package tests run on that store report afterwards.
func TestPackageTestGasIndependentOfStoreHistory(t *testing.T) {
	opts := newPackageTestOptions(t)
	files := func() map[string]string {
		return map[string]string{
			"pkggas.gno": "package pkggas\n\nimport \"strings\"\n\nfunc Repeat() string { return strings.Repeat(\"a\", 3) }\n",
			"pkggas_test.gno": `package pkggas

import "testing"

func TestA(t *testing.T) { Repeat() }

func TestB(t *testing.T) { Repeat() }
`,
		}
	}
	before := runPackageTests(t, opts, files())
	_, _, _, err := opts.RunFiletest("load.gno", []byte("package main\n\nimport \"strings\"\n\nfunc main() { _ = strings.Repeat(\"a\", 3) }\n"), opts.TestStore)
	require.NoError(t, err)
	require.Equal(t, before, runPackageTests(t, opts, files()))
}
