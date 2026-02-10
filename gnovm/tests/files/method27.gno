package main

import (
	"fmt"

	"filetests/extern/net/http"
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
