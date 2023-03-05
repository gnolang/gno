package tests

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	gno "github.com/gnolang/gno/pkgs/gnolang"
)

func TestStdCall(t *testing.T) {
	stdin := new(bytes.Buffer)
	//stdout := new(bytes.Buffer)
	stdout := os.Stdout
	stderr := new(bytes.Buffer)
	store := TestStore("..", "", stdin, stdout, stderr, ImportModeStdlibsPreferred)
	store.SetLogStoreOps(true)

	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "main",
		Output:  stdout,
		Store:   store,
		Context: nil,
	})

	c := `package main
	import "std"
	import "gno.land/p/demo/grc/grc721"
	import "gno.land/p/demo/testutils"

	type Token interface {
		Mint(std.Address,string) grc721.TokenID 
	}

	func main() {
		result := std.Call("gno.land/r/demo/nft","GetToken",nil)
		println(result.Value)
		println(len(result.Value))
		tt := result.Value[0].(Token)

		addr1 := testutils.TestAddress("addr1")
		tid := tt.Mint(addr1,"hello")
		println("tid: ",tid)

	}
`
	n := gno.MustParseFile("test", c)
	m.RunFiles(n)
	m.RunMain()

	fmt.Printf("output: %v\n", stdout)

	err := m.CheckEmpty()
	if err != nil {
		t.Log("last state: \n", m.String())
		panic(fmt.Sprintf("machine not empty after main: %v", err))
	}

}
