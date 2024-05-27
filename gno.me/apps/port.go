package apps

import (
	"context"

	"github.com/gnolang/gno/gno.me/gno"
)

func CreatePort(vm gno.VM) error {
	_, err := vm.Create(context.Background(), portAppDefinition, false, false)
	return err
}

const portAppDefinition = `
package port

var number string

func Number() string {
	return number
}

func Set(p string) {
	number = p
}
`
