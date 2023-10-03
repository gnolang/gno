// flippando/v1.go
package uflippando

import "gno.land/r/demo/uflippando"

type FlippandoV1 struct{}

func (v *FlippandoV1) GetPackageName() string {
	return "V1: Package Name: Flippando"
}

func (v *FlippandoV1) StartGame() string {
	return "V1: Game Started"
}

func (v *FlippandoV1) EndGame() string {
	return "V1: Game Ended"
}

// Automatically register the implementation
func init() {
	uflippando.RegisterImplementation("FlippandoV1", &FlippandoV1{})
}
