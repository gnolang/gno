package multitxtest

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestSlicePopPush(t *testing.T) {
	pkgName := "main"
	pkgPath := "gno.land/r/demo/slicetest"
	RunTxs(t, []TxDefinition{
		{
			Pkg: &std.MemPackage{
				Name: pkgName,
				Path: pkgPath,
				Files: []*std.MemFile{
					{
						Name: "main1.gno",
						Body: `
							package main
							import "gno.land/r/demo/multitxtest"
							import "std"
							func main1() {
								multitxtest.Pop()
							}
						`,
					},
				},
			},
			Entrypoint: "main1",
		},
		{
			Pkg: &std.MemPackage{
				Name: pkgName,
				Path: pkgPath,
				Files: []*std.MemFile{
					{
						Name: "main2.gno",
						Body: `
							package main
							import "gno.land/r/demo/multitxtest"
							import "std"
							func main2() {
								multitxtest.Push()
							}
						`,
					},
				},
			},
			Entrypoint: "main2",
		},
		{
			Pkg: &std.MemPackage{
				Name: pkgName,
				Path: pkgPath,
				Files: []*std.MemFile{
					{
						Name: "main3.gno",
						Body: `
							package main
							import "gno.land/r/demo/multitxtest"
							import "std"
							func main3() {
								slice := multitxtest.GetSlice()
								if len(slice) != 1 || slice[0] != "new-element" {
									panic("pop/push is borked")
								}
							}
						`,
					},
				},
			},
			Entrypoint: "main3",
		},
	})
}
