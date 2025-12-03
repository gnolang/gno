package rpc

import (
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

type (
	// newQueryContextDelegate creates a new app query context (read-only)
	newQueryContextDelegate func(height int64) (sdk.Context, error)

	// vmKeeperDelegate returns the VM keeper associated with the app
	vmKeeperDelegate func() vm.VMKeeperI
)

type mockApplication struct {
	newQueryContextFn newQueryContextDelegate
	vmKeeperFn        vmKeeperDelegate
}

func (m *mockApplication) NewQueryContext(height int64) (sdk.Context, error) {
	if m.newQueryContextFn != nil {
		return m.newQueryContextFn(height)
	}

	return sdk.Context{}, nil
}

func (m *mockApplication) VMKeeper() vm.VMKeeperI {
	if m.vmKeeperFn != nil {
		return m.vmKeeperFn()
	}

	return nil
}
