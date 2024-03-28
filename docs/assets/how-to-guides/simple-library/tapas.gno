package tapas

import "std"

// List of tapas suggestions
var listOfTapas = []string{
	"Patatas Bravas",
	"Gambas al Ajillo",
	"Croquetas",
	"Tortilla Española",
	"Pimientos de Padrón",
	"Jamon Serrano",
	"Boquerones en Vinagre",
	"Calamares a la Romana",
	"Pulpo a la Gallega",
	"Tostada con Tomate",
	"Mejillones en Escabeche",
	"Chorizo a la Sidra",
	"Cazón en Adobo",
	"Banderillas",
	"Espárragos a la Parrilla",
	"Huevos Rellenos",
	"Tuna Empanada",
	"Sardinas a la Plancha",
}

// GetTapaSuggestion randomly selects and returns a tapa suggestion
func GetTapaSuggestion(userInput string) string {

	// Create a random number depending on the block height.
	// We get the block height using std.GetHeight(), which is from an imported Gno library, "std"
	// Note: this value is not fully random and is easily guessable
	randomNumber := int(std.GetHeight()) % len(listOfTapas)

	// Return the random suggestion
	return listOfTapas[randomNumber]
}
