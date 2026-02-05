// Render default realm view when there are no boards.
package main

import (
	boards2 "gno.land/r/gnoland/boards2/v1"
)

func main() {
	println(boards2.Render(""))
}

// Output:
// # Boards
// [Create Board](/r/gnoland/boards2/v1:create-board) • [List Admin Users](/r/gnoland/boards2/v1:admin-users) • [Help](/r/gnoland/boards2/v1:help)
//
// > [!NOTE] 3 THINGS YOU NEED TO GET STARTED
// > To post and comment you need: a **Gno address** (create one with [Adena](https://adena.app) or [gnokey](https://docs.gno.land/users/interact-with-gnokey)), **GNOT** to pay for transactions (get some from the [faucet](https://faucet.gno.land/)), and a **registered username** ([register yours](/r/gnoland/users/v1)).
// ---
// ### Currently there are no boards
// Be the first to [create a new board](/r/gnoland/boards2/v1:create-board) !
