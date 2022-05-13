package tests

import (
	"bytes"
	"os"
	"testing"

	"github.com/gnolang/gno"
)

func RunPackageTest(t *testing.T, dir string, path string) error {
	memPkg := gno.ReadMemPackage(dir, path)

	stdin := new(bytes.Buffer)
	// stdout := new(bytes.Buffer)
	stdout := os.Stdout
	stderr := new(bytes.Buffer)
	store := testStore("..", stdin, stdout, stderr, false)
	store.SetLogStoreOps(true)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "test",
		Output:  stdout,
		Store:   store,
		Context: nil,
	})
	m.TestMemPackage(t, memPkg)

	// Check that machine is empty.
	err := m.CheckEmpty()
	if err != nil {
		t.Log("last state: \n", m.String())
		return err
	}
	return nil
}
