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

type Signup struct {
	account string
	height  int64
}

// signups - keep a slice of signed up addresses efficient pagination
var signups []Signup

// tracker - keep track of who signed up
var tracker *avl.Tree

const (
	pageSize     = 20
	pathArgument = "number"
)

func init() {
	tracker = avl.NewTree()
	SignUp()
}

func SignUp() string {
	caller := std.GetOrigCaller().String()
	height := std.GetHeight()

	if _, exists := tracker.Get(caller); exists {
		panic(caller + " is already signed up!")
	}

	tracker.Set(caller, struct{}{})
	signup := Signup{
		caller,
		height,
	}

	signups = append(signups, signup)
	return ufmt.Sprintf("%s signed up at block #%d!", signup.account, signup.height)
}

func GetSignedUpInRange(page, pageSize int) []Signup {
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
	if startIndex >= len(signups) {
		return nil
	}

	// If page contains fewer posts than the page size
	if endIndex > len(signups) {
		endIndex = len(signups)
	}

	return signups[startIndex:endIndex]
}

func renderHelper(res *mux.ResponseWriter, req *mux.Request) {
	totalSignups := len(signups)

	res.Write("# Welcome to Gno.land!\n\n")
	res.Write(ufmt.Sprintf("## Signed up users (%d):\n\n", totalSignups))

	page, err := strconv.Atoi(req.GetVar(pathArgument))
	if err != nil {
		page = 1 // render first page on bad input
	}

	if totalSignups == 0 {
		res.Write("No users are signed up currently.")
		return
	}

	fetchedSignups := GetSignedUpInRange(page, pageSize)
	if totalSignups == 0 {
		res.Write("No users on this page\n\n")
	}

	pageStartIndex := pageSize * (page - 1)
	for i, signup := range fetchedSignups {
		out := ufmt.Sprintf("#### User #%d - %s - signed up at Block #%d\n", pageStartIndex+i, signup.account, signup.height)
		res.Write(out)
	}

	res.Write("---\n\n")

	latestSignup := signups[totalSignups-1]
	res.Write(ufmt.Sprintf("Latest signup: %s at Block #%d\n\n", latestSignup.account, latestSignup.height))

	res.Write("---\n\n")

	res.Write(ufmt.Sprintf("You're viewing page #%d\n\n", page))

	var prevPage string
	var nextPage string
	if page >= 2 {
		prevPage = ufmt.Sprintf("[Previous page](/r/demo/pagination:page/%d)\n\n", page-1)
	}

	if len(fetchedSignups) >= pageSize {
		nextPage = ufmt.Sprintf("[Next page](/r/demo/pagination:page/%d)\n\n", page+1)
	}

	res.Write(prevPage)
	res.Write(nextPage)
}

func Render(_ string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHelper)
	path := ufmt.Sprintf("page/{%s}", pathArgument)
	router.HandleFunc(path, renderHelper)
	return router.Render(path)
}
