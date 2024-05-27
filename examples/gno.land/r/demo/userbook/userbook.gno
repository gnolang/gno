// This realm demonstrates a small userbook system working with gnoweb
package userbook

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
var (
	tracker *avl.Tree
	router  *mux.Router
)

const (
	defaultPageSize = 20
	pathArgument    = "number"
	subPath         = "page/{" + pathArgument + "}"
)

func init() {
	// Set up tracker tree
	tracker = avl.NewTree()

	// Set up route handling
	router = mux.NewRouter()
	router.HandleFunc("", renderHelper)
	router.HandleFunc(subPath, renderHelper)

	// Sign up the deployer
	SignUp()
}

func SignUp() string {
	// Get transaction caller
	caller := std.PrevRealm().Addr().String()
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
	return ufmt.Sprintf("%s added to userbook up at block #%d!", signup.account, signup.height)
}

func GetSignupsInRange(page, pageSize int) ([]Signup, int) {
	if page < 1 {
		panic("page number cannot be less than 1")
	}

	if pageSize < 1 || pageSize > 50 {
		panic("page size must be from 1 to 50")
	}

	// Pagination
	// Calculate indexes
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	// If page does not contain any users
	if startIndex >= len(signups) {
		return nil, -1
	}

	// If page contains fewer users than the page size
	if endIndex > len(signups) {
		endIndex = len(signups)
	}

	return signups[startIndex:endIndex], endIndex
}

func renderHelper(res *mux.ResponseWriter, req *mux.Request) {
	totalSignups := len(signups)
	res.Write("# Welcome to UserBook!\n\n")

	// Get URL parameter
	page, err := strconv.Atoi(req.GetVar("number"))
	if err != nil {
		page = 1 // render first page on bad input
	}

	// Fetch paginated signups
	fetchedSignups, endIndex := GetSignupsInRange(page, defaultPageSize)
	// Handle empty page case
	if len(fetchedSignups) == 0 {
		res.Write("No users on this page!\n\n")
		res.Write("---\n\n")
		res.Write("[Back to Page #1](/r/demo/userbook:page/1)\n\n")
		return
	}

	// Write page title
	res.Write(ufmt.Sprintf("## UserBook - Page #%d:\n\n", page))

	// Write signups
	pageStartIndex := defaultPageSize * (page - 1)
	for i, signup := range fetchedSignups {
		out := ufmt.Sprintf("#### User #%d - %s - signed up at Block #%d\n", pageStartIndex+i, signup.account, signup.height)
		res.Write(out)
	}

	res.Write("---\n\n")

	// Write UserBook info
	latestSignupIndex := totalSignups - 1
	res.Write(ufmt.Sprintf("#### Total users: %d\n", totalSignups))
	res.Write(ufmt.Sprintf("#### Latest signup: User #%d at Block #%d\n", latestSignupIndex, signups[latestSignupIndex].height))

	res.Write("---\n\n")

	// Write page number
	res.Write(ufmt.Sprintf("You're viewing page #%d", page))

	// Write navigation buttons
	var prevPage string
	var nextPage string
	// If we are on any page that is not the first page
	if page > 1 {
		prevPage = ufmt.Sprintf(" - [Previous page](/r/demo/userbook:page/%d)", page-1)
	}

	// If there are more pages after the current one
	if endIndex < totalSignups {
		nextPage = ufmt.Sprintf(" - [Next page](/r/demo/userbook:page/%d)\n\n", page+1)
	}

	res.Write(prevPage)
	res.Write(nextPage)
}

func Render(path string) string {
	return router.Render(path)
}
