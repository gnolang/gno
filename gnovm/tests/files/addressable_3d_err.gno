package main

func main() {
	_ = &(&struct{}{})
}

// Error:
// main/files/addressable_3d_err.gno:4:7: cannot take address of struct {  }{}
// *** CHECK THE ERR MESSAGES ABOVE, MAKE SURE IT'S WHAT YOU EXPECTED, DELETE THIS LINE AND RUN TEST AGAIN ***
