package gno

import (
	"context"
	"errors"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var ErrPackageNotMain = errors.New("run method code must have a package name of main")

func (v *VMKeeper) Run(ctx context.Context, code string) (string, error) {
	v.Lock()
	defer v.Unlock()
	packageName, err := getPackagename(code)
	if err != nil {
		return "", err
	}

	// The Run method expects a package name of main; enforce this here.
	if packageName != "main" {
		return "", ErrPackageNotMain
	}

	msg := vm.MsgRun{
		Package: &std.MemPackage{
			Name: "main",
			Files: []*std.MemFile{
				{Name: "main.gno", Body: code},
			},
		},
	}
	return v.instance.Run(sdk.Context{}.WithContext(ctx), msg)
}
