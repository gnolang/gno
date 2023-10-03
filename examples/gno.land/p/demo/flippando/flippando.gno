package flippando

type GameInterface interface {
	StartGame() string
	EndGame() string
}

var currentImplementation GameInterface

func SetImplementation(impl GameInterface) {
	// todo: make it only owner
	currentImplementation = impl
}

func StartGame() string {
	if currentImplementation == nil {
		return "No implementation set"
	}
	return currentImplementation.StartGame()
}

func EndGame() string {
	if currentImplementation == nil {
		return "No implementation set"
	}
	return currentImplementation.EndGame()
}
