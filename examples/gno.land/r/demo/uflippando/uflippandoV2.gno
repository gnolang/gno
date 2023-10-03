// flippando/v2.go
package uflippando

import "gno.land/r/demo/uflippando"

type FlippandoV2 struct{}

func (v *FlippandoV2) StartGame() string {
	return "V2: Game Started"
}

func (v *FlippandoV2) EndGame() string {
	return "V2: Game Ended"
}

// Only in V2
func (v *FlippandoV2) GetUserGames() string {
	return "V2: Fetching user games..."
}

// Automatically register the implementation
func init() {
	uflippando.RegisterImplementation("FlippandoV2", &FlippandoV2{})
}

