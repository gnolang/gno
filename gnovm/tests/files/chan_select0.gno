// select statement is rejected at parse time.
package main

func main() {
	select {}
}

// Error:
// chan_select0.gno:5:2: select statements are not permitted
