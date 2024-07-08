package submodule

import (
	"github.com/gnolang/gno/tm2/pkg/amino/genproto/example/submodule2"
)

type StructSM struct {
	FieldA int
	FieldB string
	FieldC submodule2.StructSM2
}
