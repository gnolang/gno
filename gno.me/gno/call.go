package gno

import (
	"context"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func (v VMKeeper) Call(
	ctx context.Context,
	appName string,
	isPackage bool,
	functionName string,
	args ...string,
) (string, error) {
	defer v.store.Commit()

	prefix := AppPrefix
	if isPackage {
		prefix = PkgPrefix
	}

	msg := vm.MsgCall{
		PkgPath: prefix + appName,
		Func:    functionName,
		Args:    append([]string{}, args...),
	}
	return v.instance.Call(sdk.Context{}.WithContext(ctx), msg)
}
