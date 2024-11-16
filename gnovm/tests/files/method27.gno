package main

import (
	"fmt"

	"github.com/gnolang/gno/_test/net/http"
)

type AuthenticatedRequest struct {
	http.Request
	Username string
}

func main() {
	a := &AuthenticatedRequest{}
	fmt.Println("ua:", a.UserAgent(), ";")
}

// Output:
// ua:  ;
