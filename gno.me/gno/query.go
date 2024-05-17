package gno

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Do not use after initialization -- may not be thread-safe.
func (v *VMKeeper) QueryMemPackages(ctx context.Context) <-chan *std.MemPackage {
	return v.instance.QueryMemPackages(sdk.Context{}.WithContext(ctx))
}

func (v *VMKeeper) QueryMemPackage(ctx context.Context, appName string) *std.MemPackage {
	v.Lock()
	defer v.Unlock()

	appName = AppPrefix + appName
	return v.instance.QueryMemPackage(sdk.Context{}.WithContext(ctx), appName)
}
