# Anatomy of a Gno package

In this tutorial, you will learn to make a simple `Counter` application in
Gno. We will cover the basics of the Gno language which will help you get
started writing smart contracts for Gno.land.

## Package Types Overview

Gno.land supports three types of packages:
- **Realms (`/r/`)**: Stateful user applications (smart contracts) that
  maintain persistent state between transactions
- **Pure Packages (`/p/`)**: Stateless libraries that provide reusable 
  functionality
- **Ephemeral Packages (`/e/`)**: Temporary code execution with MsgRun
  which allows a custom main() function to be run instead of a single
  function call as with MsgExec.

Ephemeral packages can be created with [`gnokey maketx
run`](../users/interact-with-gnokey.md#run). They are designed for
executing complex logic or workflows (like `maketx call` on steroids) but
exist only during their execution which is encapsulated in its own
execution environment.

### Import Rules

The import rules between different package types are:
- **Pure packages (`/p/`)** can be imported by everything: realms, other 
  packages, and ephemeral realms
- **Realms (`/r/`)** can only be imported by other realms or ephemeral realms 
  (not by pure packages)
- **Ephemeral packages (`/e/`)** cannot be imported by anything

When importing:
- Importing a **realm** gives access to its exported functions and interacts 
  with the realm's persistent state
- Importing a **pure package** gives access to its exported functions without 
  any state persistence

## Language basics

Let's dive into the `Counter` example.

First, we need to declare a package name.

```go
package counter
```

A package is an organizational unit of code; it can contain multiple files, and
as mentioned in previous tutorials, it lives on a specific package path once
deployed to the network.

Next, let us declare a top level variable of type `int`:

```go
package counter

var count int
```

In Gno, all top-level variables will automatically be persisted to the network's
state after a successful transaction modifying them. Here, you can define
variables that will store your smart contract's data.

In our case, we have defined a variable which will store the counter's state.

Next, let's define functions that users will be able to call to change the state
of the counter:

```go
package counter

var count int

func Increment(_ realm, change int) int {
	count += change
	return count
}
```

The `Increment()` function has a few important features:
- When written with the first letter in uppercase, the function is
exported. This means that calls to this function from outside the `counter`
package are allowed - be it from off-chain clients, users, or from other Gno programs
- As this function intends to change the state of the realm (incrementing the 
`count` variable), it needs to be ["crossing"](../resources/gno-interrealm.md). 
To declare the function crossing, the first argument must be of type `realm`, 
which is a custom Gno built-in keyword. In this case, as we won't be using the 
argument for anything, we can leave it unnamed (i.e., `_`).
- It takes a second argument of type `int`, called `change`. This is how the caller
will provide a specific number which will be used to increment the `counter`
- Returns the value of `count` after a successful call

Next, to make our application more user-friendly, we should define a `Render()`
function. This function will help users see the current state of the Counter
application.

```go gno path=counter.gno run_expr=println(Render(""))
package counter

import "strconv"

var count int

func Increment(_ realm, change int) int {
	count += change
	return count
}

func Render(_ string) string {
	return "Current counter value: " + strconv.Itoa(count)
}
```

In our case, we can replace the argument string with a `_`, signifying an unused
variable. Then, we can simply return a string telling us the current value of
`count`. For converting `count` to a string, we can import the `strconv` package
from the Gno standard library, as we do when writing Go code.

:::info
A valid `Render()` function needs to have the following signature:
```go
func Render(path string) string {
	...
}
```
:::

## Writing unit tests

Following best practices, developers should test their Gno applications to avoid
bugs and other problems down the line.

Let's see how we can write a simple test for the `Increment()` function.

```go gno path=counter_test.gno depends_on=counter.gno
package counter

import "testing"

func TestIncrement(t *testing.T) {
	// Check initial value
	if count != 0 {
		t.Fatalf("Expected 0, got %d", count)
	}

	// Call Increment
	value := Increment(cross, 42)

	// Check result
	if value != 42 {
		t.Fatalf("Expected 42, got %d", count)
	}
}
```

By using the `testing` package from the standard library, we can access the
`testing.T` object that exposes methods which can help us terminate tests in specific cases.
Next, to satisfy the first argument of the `Increment()` function, we will use the
built-in `cross`.

:::info
Common testing patterns found in Go, such as [TDT](https://go.dev/wiki/TableDrivenTests),
can also be used for Gno. We recommend checking out some of the many examples
found online.
:::
