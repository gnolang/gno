package apps

import (
	"context"

	"github.com/gnolang/gno/gno.me/gno"
)

func CreatePort(vm gno.VM) error {
	return vm.Create(context.Background(), portAppDefinition, false)
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
