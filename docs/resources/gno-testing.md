# Running & Testing Gno code

## Prerequisites

- `gno` set up. See [Installation](../users/interact-with-gnokey.md).

## Overview

In this tutorial, you will learn how to run and test your Gno code locally, by
using the `gno` binary. For this example, we will use the `Counter` code from a
[previous tutorial](../builders/anatomy-of-a-gno-package.md).

## Setup

Start by creating a directory which will contain your Gno code:

```
mkdir counter
cd counter
```

Next, initialize a `gnomod.toml` file. This file declares the package path of your
realm and the Gno language version, and is required by Gno tooling. You can do
this using the following command:

```
gno mod init gno.land/r/<namespace>/counter
```

Replace `<namespace>` with your username. In this example, we’ll use `example`.
This command creates a file with the following content:

```
module gno.land/r/example/counter
```

Then, in the same directory, create two files- `counter.gno` & `counter_test.gno`:

```
touch counter.gno counter_test.gno
```

Paste the code from the previous tutorial into these files.

`counter.gno`:
```go
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

`counter_test.gno`:
```go
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

## Testing

To run package tests, use the `gno test` subcommand and pass it the directory
containing the tests. From inside the `counter/` directory, run:

```
$ gno test .
ok      .       0.81s
```

To see a detailed list of tests that were executed, we can add the verbose flag:

```
$ gno test . -v
=== RUN   TestIncrement
--- PASS: TestIncrement (0.00s)
ok      .       0.81s
```

In addition to -v, other flags are available, such as options for setting test
timeouts, checking performance metrics, and more.

:::info Mocked testing & running environment
The `gno` binary mocks a blockchain environment when running & testing code.
See [Final remarks](#final-remarks).
:::

## Running Gno code

The `gno` binary also provides a `run` subcommand, which allows you to evaluate
specific expressions in your Gno code without starting a full blockchain
environment. Internally, the Gno Virtual Machine (GnoVM) evaluates the given
expression and returns its value, without making any permanent changes to
contract storage.

This is a convenient way to quickly test or evaluate a function during
development without spinning up a full blockchain.

By default, the GnoVM does not automatically print return values when evaluating
expressions. For this reason, you need to include a `println()` call—either inside
the function itself or directly in the evaluated expression:

```
gno run -expr "println(Increment(42))"
```

Try running this expression for yourself:

```go gno run-expression=println(Increment(42))
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

The `run` subcommand also supports a full GnoVM debugger, which can be started
with the `-debug` flag. Read more about it [here](https://gno.land/r/gnoland/blog:p/gno-debugger).

## Final remarks

Note that executing and testing code as shown in this tutorial utilizes a local,
mocked execution environment. During testing and expression evaluation, the GnoVM
is simply running as a language interpreter, with no connection to a real blockchain.

No real on-chain transactions occur, and the state changes are purely in-memory
for testing and development purposes. You might notice this if you run the same
expression modifying a state variable twice, with the actual value staying unchanged.

All imports used in your code are resolved from the GnoVM’s installation
directory.

## Conclusion

That's it 🎉

You've successfully run local tests and expressions using the `gno` binary.
