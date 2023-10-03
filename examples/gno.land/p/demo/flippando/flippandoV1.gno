// flippando/v1.go
package flippando

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
