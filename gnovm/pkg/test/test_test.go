package test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/require"
)

// Package tests charge allocations the same gas filetests do.
func TestPackageTestGasChargesAllocations(t *testing.T) {
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)
	const pkgPath = "gno.land/p/test/allocgas"

	dir := t.TempDir()
	write := func(name, body string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
	}
	write("gnomod.toml", gno.GenGnoModLatest(pkgPath))
	write("allocgas.gno", "package allocgas\n\nfunc Alloc(n int) []byte { return make([]byte, n) }\n")
	// Same length names, so the two tests differ only in what they allocate.
	write("allocgas_test.gno", `package allocgas

import "testing"

func TestAllocLo(t *testing.T) { Alloc(1) }

func TestAllocHi(t *testing.T) { Alloc(10000) }
`)
	var stderr bytes.Buffer
	opts := NewTestOptions(rootDir, io.Discard, &stderr, nil)
	opts.Verbose = true
	require.NoError(t, Test(gno.MustReadMemPackage(dir, pkgPath, gno.MPAnyAll), dir, opts))
	gas := regexp.MustCompile(`--- GAS:  (\d+)`).FindAllStringSubmatch(stderr.String(), -1)
	require.Len(t, gas, 2)
	small, _ := strconv.ParseInt(gas[0][1], 10, 64)
	big, _ := strconv.ParseInt(gas[1][1], 10, 64)

	filetest := func(n int) int64 {
		src := "package main\n\nfunc alloc(n int) []byte { return make([]byte, n) }\n\nfunc main() { alloc(" + strconv.Itoa(n) + ") }\n"
		_, g, _, err := opts.RunFiletest("alloc.gno", []byte(src), opts.TestStore)
		require.NoError(t, err)
		return g
	}
	require.Equal(t, filetest(10000)-filetest(1), big-small)
}
