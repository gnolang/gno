package gno

import (
	"context"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func (v *VMKeeper) Create(ctx context.Context, code string, isPackage, syncable bool) (string, error) {
	v.Lock()
	defer v.Unlock()
	defer v.store.Commit()

	packageName, err := getPackagename(code)
	if err != nil {
		return "", err
	}

	prefix := AppPrefix
	if isPackage {
		prefix = PkgPrefix
	}

	msg := vm.MsgAddPackage{
		Package: &std.MemPackage{
			Name: packageName,
			Path: prefix + packageName,
			Files: []*std.MemFile{
				{
					Name: packageName + gnoFileSuffix,
					Body: code,
				},
			},
			Syncable: syncable,
		},
	}
	return packageName, v.instance.AddPackage(sdk.Context{}.WithContext(ctx), msg)
}

func (v *VMKeeper) CreateMemPackage(ctx context.Context, memPackage *std.MemPackage) error {
	v.Lock()
	defer v.Unlock()
	defer v.store.Commit()

	msg := vm.MsgAddPackage{
		Package: memPackage,
	}

	return v.instance.AddPackage(sdk.Context{}.WithContext(ctx), msg)
}
