---
id: simple-library
---

# How to write a simple Gno Library (Package)

## Overview

This guide shows you how to write a simple library (Package) in Gno, which can be used by other Packages and Realms.
Packages are _stateless_, meaning they do not hold state like regular Realms (Smart Contracts). To learn more about the
intricacies of Packages, please see the [Packages reference](../concepts/packages.md).

The Package we will be writing today will be a simple library for suggesting a random tapas dish.
We will define a set list of tapas, and define a method that randomly selects a dish from the list.

## Prerequisites

- **Text editor**

:::info Editor support
The Gno language is based on Go, but it does not have all the bells and whistles in major text editors like Go.
Advanced language features like IntelliSense are still in the works.

Currently, we officially have language support
for [ViM](https://github.com/gnolang/gno/blob/master/CONTRIBUTING.md#vim-support),
[Emacs](https://github.com/gnolang/gno/blob/master/CONTRIBUTING.md#emacs-support)
and [Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=harry-hov.gno).
:::

## 1. Setting up the work directory

We discussed Gno folder structures more in detail in
the [simple Smart Contract guide](simple-contract.md#1-setting-up-the-work-directory).
For now, we will just follow some rules outlined there.

Create the main working directory for our Package:

```bash
mkdir tapas-lib
```

Since we are building a simple tapas Package, inside our created `tapas-lib` directory, we can create another
directory named `p`, which stands for `package`:

```bash
cd tapas-lib
mkdir p
```

Additionally, we will create another subdirectory that will house our Package code, named `tapas`:

```bash
cd p
mkdir tapas
```

After setting up our work directory structure, we should have something like this:

```text
tapas-lib/
├─ p/
│  ├─ tapas/
│  │  ├─ // source code here
```

## 2. Create `tapas.gno`

Now that the work directory structure is set up, we can go into the `tapas` sub-folder, and actually create
our tapas suggestion library logic:

```bash
cd tapas
touch tapas.gno
```

Inside `tapas.gno`, we will define our library logic:

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

## Conclusion

That's it 🎉

You have successfully built a simple tapas suggestion Package that is ready to be deployed on the Gno chain and imported
by other Packages and Realms.
