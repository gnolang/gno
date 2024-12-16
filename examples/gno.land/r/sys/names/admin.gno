package names

import "github.com/gnolang/gno/gnovm/stdlibs/std"

// UpdateVerifyCall updates the method that verifies the namespace.
func UpdateVerifyCall(check VerifyNameFunc) error {
	if err := admin.CallerIsOwner(); err != nil {
		return err
	}

	checkFunc = check
	std.Emit(VerifyFuncUpdatedEvent)

	return nil
}
