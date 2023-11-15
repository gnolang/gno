package tapas

import (
	"gno.land/p/demo/rand"
)

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
func GetTapaSuggestion() string {
	// Create a new instance of the random number generator.
	// Notice that this is from an imported Gno library
	generator := rand.New()

	// Generate a random index
	randomIndex := generator.Intn(len(listOfTapas))

	// Return the random suggestion
	return listOfTapas[randomIndex]
}
