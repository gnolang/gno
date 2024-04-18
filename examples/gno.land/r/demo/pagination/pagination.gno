// This realm demonstrates a pagination system working with gnoweb
// It keeps a slice of signed up addresses for efficient pagination

package pagination

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

var list []string     // keep a slice of signed up addresses efficient pagination
var tracker *avl.Tree // keep track of who signed up

func SignUp() string {
	caller := std.GetOrigCaller().String()

	_, exists := tracker.Get(caller)
	if exists {
		panic(caller + " is already signed up!")
	}

	tracker.Set(caller, struct{}{})
	list = append(list, caller)

	return caller + " signed up!"
}

func GetSignedUpInRange(page, pageSize uint) []string {
	if page < 1 {
		panic("page number cannot be less than 1")
	}

	if pageSize < 1 || pageSize >= 50 {
		panic("page size must be between 1 and 50")
	}

}

func Render(path string) string {

}
