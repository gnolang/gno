package fmt

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)


func X_typeString(v gnolang.TypedValue) string {
	if v.IsUndefined() {
		return "<nil>"
	}
	return v.T.String()
}
