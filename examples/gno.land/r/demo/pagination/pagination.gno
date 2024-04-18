// This realm demonstrates a pagination system working with gnoweb
// It keeps a slice of signed up addresses for efficient pagination

package pagination

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
	"std"
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

func GetSignedUpInRange(page, pageSize int) []string {
	if page < 1 {
		panic("page number cannot be less than 1")
	}

	if pageSize < 1 || pageSize >= 50 {
		panic("page size must be between 1 and 50")
	}

	// Pagination
	// Calculate indexes
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	// If page does not contain any posts
	if startIndex >= len(list) {
		return nil
	}

	// If page contains fewer posts than the page size
	if endIndex > len(list) {
		endIndex = len(list)
	}

	return list[startIndex:endIndex]
}

func renderHelper(res *mux.ResponseWriter, req *mux.Request) {
	num := req.GetVar("number")

	if num != "" {
		res.Write("page: " + num + "!")
	} else {
		res.Write("Hello, world!")
	}
}

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("page/{number}", renderHelper)
	return router.Render(path)
}
