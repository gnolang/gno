# Running & testing Gno code

## Prerequisites

- `gno` set up. See [Installation](../users/interact-with-gnokey.md).

## Overview

In this tutorial, you will learn how to run and test your Gno code locally, by
using the `gno` binary. For this example, we will use the `Counter` code from a
[previous tutorial](../builders/anatomy-of-a-gno-package.md).

## Setup

Start by creating a folder which will contain your Gno code:

```
mkdir counter
cd counter
```

First, we should initialize a `gnomod.toml` file. This file declares the package path
of your realm & the Gno language version, and is used by Gno tools. We can do
this by using the following command:

```
gno mod init gno.land/r/<namespace>/counter
```

You can enter your username under `<namespace>`. In this case, let's use `example`.
This will create a file with the following content:

```
module gno.land/r/example/counter
```

Then, in the same folder, create two files- `counter.gno` & `counter_test.gno`:

```
touch counter.gno counter_test.gno
```

In these files, paste in the code from the previous tutorial.

`counter.gno`:
```go
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
	value := Increment(42)

	// Check result
	if value != 42 {
		t.Fatalf("Expected 42, got %d", count)
	}
}
```

## Testing

To run package tests, we can simply use the `gno test` subcommand, passing it the
directory that contains the tests. From inside the `counter/` directory, we
can run the following:

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

Apart from `-v`, other flags are also available, such as ones for setting the
test timeout, checking performance metrics, etc.

:::info Mocked testing & running environment
The `gno` binary mocks a blockchain environment when running & testing code.
See [Final remarks](#final-remarks).
:::

## Running Gno code

The `gno` binary contains a `run` subcommand, allowing users to evaluate
specific expressions in their Gno code, without a full blockchain environment.
Under the hood, the Gno Virtual Machine evaluates the given expression and simply
returns the value, without any permanent changes to contract storage.

This can be an easy way to quickly evaluate a function during development, without
having to spin up a full blockchain environment.

The GnoVM won't automatically print out return values upon evaluating expressions,
which is why we need to include a `println()` somewhere- either in the function
body itself, or modify the expression itself:

```
gno run -expr "println(Increment(42))"
```

Try running this expression for yourself:

```go gno run-expression=println(Increment(42))
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

The `run` subcommand also supports a full GnoVM debugger, which can be started
with the `-debug` flag. Read more about it [here](https://gno.land/r/gnoland/blog:p/gno-debugger).

## Final remarks

Note that executing and testing code as shown in this tutorial  utilizes a local,
mocked execution environment. During testing and expression evaluation, the GnoVM
is simply running as a language interpreter, with no connection to a real blockchain.

No real on-chain transactions occur, and the state changes are purely in-memory
for testing and development purposes. You might notice this if you run the same
expression modifying a state variable twice, with the actual value staying unchanged.

All possible imports in your code are resolved from the GnoVM's installation folder.

## Conclusion

That's it 🎉

You've successfully run local tests and expressions using the `gno` binary.
