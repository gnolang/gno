---
id: testing-gno
---

# How to test Gno Code

## Overview

In this guide, we will explore the available tooling in testing out the Gno Realms and Packages we write.
We will go over different CLI tools available to developers, gno testing libraries as well as
testing techniques that involve data mocking.

## Prerequisites

- **`gno` set up. Reference the [Installation](../getting-started/local-setup/local-setup.md#3-installing-other-gno-tools) guide
  for steps**

## Example Realm

For the purpose of this guide, we will be testing the simple *Counter* Realm created in
the [How to write a simple Gno Smart Contract (Realm)](simple-contract.md) guide.

[embedmd]:# (../assets/how-to-guides/testing-gno/counter-1.gno go)
```go
// counter-app/r/counter/counter.gno

package counter

import (
	"gno.land/p/demo/ufmt"
)

var count int

func Increment() {
	count++
}

func Decrement() {
	count--
}

func Render(_ string) string {
	return ufmt.Sprintf("Count: %d", count)
}
```

## 1. Writing the Gno test

Gno tests are written in the same manner and format as regular Go tests, just in `_test.gno` files.

We can place the Gno tests for the `Counter` Realm in the same directory as `counter.gno`:

```text
counter-app/
├─ r/
│  ├─ counter/
│  │  ├─ counter.gno
│  │  ├─ counter_test.gno  <--- the test source code
```

```bash
cd counter
touch counter_test.gno
```

What should be tested in this _Counter_ Realm example?
Mainly, we want to verify that:

- Increment increments the value.
- Decrement decrements the value.
- Render returns a valid formatted value.

Let's write the required unit tests:

[embedmd]:# (../assets/how-to-guides/testing-gno/counter-2.gno go)
```go
// counter-app/r/counter/counter_test.gno

package counter

import "testing"

func TestCounter_Increment(t *testing.T) {
	// Reset the value
	count = 0

	// Verify the initial value is 0
	if count != 0 {
		t.Fatalf("initial value != 0")
	}

	// Increment the value
	Increment()

	// Verify the initial value is 1
	if count != 1 {
		t.Fatalf("initial value != 1")
	}
}

func TestCounter_Decrement(t *testing.T) {
	// Reset the value
	count = 0

	// Verify the initial value is 0
	if count != 0 {
		t.Fatalf("initial value != 0")
	}

	// Decrement the value
	Decrement()

	// Verify the initial value is 1
	if count != -1 {
		t.Fatalf("initial value != -1")
	}
}

func TestCounter_Render(t *testing.T) {
	// Reset the value
	count = 0

	// Verify the Render output
	if Render("") != "Count: 0" {
		t.Fatalf("invalid Render value")
	}
}
```

:::warning Testing package-level variables

In practice, it is not advisable to test and validate package level variables like this, as their value is mutated
between test runs. For the sake of keeping this guide simple, we went ahead and reset the variable value for each test,
however,
you should employ more robust test strategies.

:::

## 2. Running the Gno test

To run the prepared Gno tests, we can utilize the `gno test` CLI tool.

Simply point it to the location containing our testing source code, and the tests will execute.
For example, we can run the following command from the `counter-app/r/counter` directory:

```bash
gno test -v .
```

Let's look into the different parts of this command:

- `-v` enables the verbose output.
- `-root-dir` specifies the root directory to our cloned `gno` GitHub repository
- `.` specifies the location containing our test files. Since we are already located in that directory, we specify
  a `.`.

Running the test command should produce a successful output:

```bash
=== RUN   TestCounter_Increment
--- PASS: TestCounter_Increment (0.00s)
=== RUN   TestCounter_Decrement
--- PASS: TestCounter_Decrement (0.00s)
=== RUN   TestCounter_Render
--- PASS: TestCounter_Render (0.00s)
ok      ./. 	1.00s
```

## Additional test support

As we grow more familiar with Gno development, our Realm / Package logic can become more complex. As such, we need
more robust testing support in the form of mocking values ahead of time that would normally be only available on a
live (deployed) Realm / Package.

Luckily, the Gno standard library provides ample support for functionality such as setting predefined values ahead of
time, such as the request caller address, or the calling package address.

You can learn more about these methods, that are importable using the `std` import declaration,
in the [Standard Library](../concepts/stdlibs/stdlibs.md) reference section.

## Conclusion

That's it 🎉

You have successfully written and tested Gno code. Additionally, you have utilized the `gno test` tool, and understood
how it can be configured to make the developer experience smooth.