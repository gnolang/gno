package uflippando

import {
	"reflect"
	"errors"
}

type GameInterface interface {
	StartGame() string
	EndGame() string
}

// Current implementation's identifier
var currentImplementationKey string

// add only_owner
var registry map[string]GameInterface = make(map[string]GameInterface)

func RegisterImplementation(key string, pkg reflect.Value) error {
	// Iterate over the provided package's exported functions
	for i := 0; i < pkg.NumMethod(); i++ {
		method := pkg.Method(i)

		// Check if the method matches our naming scheme or struct tag
		if isValidGameImplementation(method.Name) {
			registry[key] = method.Interface().(GameInterface)
			return nil
		}
	}

	return errors.New("Valid game implementation not found")
}

func isValidGameImplementation(name string) bool {
	return name == "GameImplementationName" // Change this to whatever your naming scheme is
}

// Set the current implementation using its identifier
func SetImplementation(implName string) {
	// Todo: make it only owner
	if _, exists := registry[implName]; exists {
		currentImplementationKey = implName
	}
}

func StartGame() string {
	impl, exists := registry[currentImplementationKey]
	if !exists {
		return "No implementation set"
	}
	return impl.StartGame()
}

func EndGame() string {
	impl, exists := registry[currentImplementationKey]
	if !exists {
		return "No implementation set"
	}
	return impl.EndGame()
}
