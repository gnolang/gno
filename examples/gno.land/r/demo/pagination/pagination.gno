// This realm demonstrates a pagination system working with gnoweb
// It keeps a slice of signed up addresses for efficient pagination

package pagination

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
)

var list []string     // keep a slice of signed up addresses efficient pagination
var tracker *avl.Tree // keep track of who signed up

func init() {
	tracker = avl.NewTree()
	SignUp()
}

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
	res.Write("# Welcome to Gno.land!\n\n")
	res.Write("## Signed up users:\n\n")

	page, err := strconv.Atoi(req.GetVar("number"))
	if err != nil {
		page = 1 // render first page on bad input
	}

	addresses := GetSignedUpInRange(page, 20)
	if len(addresses) == 0 {
		res.Write("No users are signed up currently.")
	}

	for i, addr := range addresses {
		out := ufmt.Sprintf("#### User #%d - %s\n", i, addr)
		res.Write(out)
	}

	res.Write("---\n\n")
	res.Write(ufmt.Sprintf("You're viewing page #%d\n\n", page))

	nextPage := ufmt.Sprintf("[Next page](page/%d)", page+1)
	res.Write(nextPage)
}

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHelper)
	router.HandleFunc("page/{number}", renderHelper)
	return router.Render(path)
}
