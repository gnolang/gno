# Anatomy of a Gno package

In this tutorial, you will learn to make a simple `Counter` application in
Gno. We will cover the basics of the Gno language which will help you get
started writing smart contracts for gno.land.

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

func Increment(change int) int {
	count += change
	return count
}
```

The `Increment()` function has a few important features:
- When written with the first letter in uppercase, the function is
  exported. This means that calls to this function from outside the `counter`
  package are allowed - be it from off-chain clients or from other Gno programs
- It takes an argument of type `int`, called `change`. This is how the caller
  will provide a specific number which will be used to increment the `counter`
- Returns the value of `count` after a successful call

Next, to make our application more user-friendly, we should define a `Render()`
function. This function will help users see the current state of the Counter
application.

```go gno path=counter.gno run_expr=println(Render(""))
package counter

import "strconv"

var count int

func Increment(change int) int {
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
	value := Increment(42)

	// Check result
	if value != 42 {
		t.Fatalf("Expected 42, got %d", count)
	}
}
```

By using the `testing` package from the standard library, we can access the
`testing.T` object that exposes methods which can help us terminate tests in specific cases.

:::info
Common testing patterns found in Go, such as [TDT](https://go.dev/wiki/TableDrivenTests),
can also be used for Gno. We recommend checking out some of the many examples
found online.
:::
