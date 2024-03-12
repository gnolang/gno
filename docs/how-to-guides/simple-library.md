---
id: simple-library
---

# How to write a simple Gno Library (Package)

## Overview

This guide shows you how to write a simple library (Package) in Gno, which can be used by other Packages and Realms.
Packages are _stateless_, meaning they do not hold state like regular Realms (Smart Contracts). To learn more about the
intricacies of Packages, please see the [Packages concept page](../concepts/packages.md).

The Package we will be writing today will be a simple library for suggesting a random tapas dish.
We will define a set list of tapas, and define a method that randomly selects a dish from the list.

## Prerequisites

- **Internet connection**

## 1. Using Gno Playground

When using the Gno Playground, writing, testing, deploying, and sharing Gno code
is simple. This makes it perfect for getting started with Gno.

Vising the [Playground](https://play.gno.land) will greet you with a template file:

![Default](../assets/how-to-guides/simple-library/playground_welcome.png)

## 2. Start writing code

Inside `package.gno`, we will define our library logic:

[embedmd]:# (../assets/how-to-guides/simple-library/tapas.gno go)
```go
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
```

There are a few things happening here, so let's dissect them:

- We defined the logic of our library into a package called `tapas`.
- The package imports another gno package, which is deployed at `gno.land/p/demo/rand`
- We use the imported package inside of `GetTapaSuggestion` to generate a random index value for a tapa

You can view the code on [this Playground link](https://play.gno.land/p/5SQQ-r2_Vos).

## Conclusion

That's it 🎉

You have successfully built a simple tapas suggestion Package that is ready to be deployed on the Gno chain and imported
by other Packages and Realms.
