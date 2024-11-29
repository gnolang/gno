// This realm demonstrates a small userbook system working with gnoweb
package userbook

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"std"
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
	signUpEvent = "SignUp"
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

	// Check if the user is already signed up
	if _, exists := tracker.Get(caller); exists {
		panic(caller + " is already signed up!")
	}

	// Sign up the user
	tracker.Set(caller, struct{}{})
	signup := Signup{
		caller,
		height,
	}

	signups = append(signups, signup)
	std.Emit(signUpEvent, "SignedUpAccount", signup.account)

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
