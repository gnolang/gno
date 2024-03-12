---
id: simple-contract
---

# How to write a simple Gno Smart Contract (Realm)

## Overview

This guide shows you how to write a simple **Counter** Smart Contract, or rather a [Realm](../concepts/realms.md),
in [Gno](../concepts/gno-language.md). 

For actually deploying the Realm, please see the [deployment](deploy.md) guide.

Our _Counter_ Realm will have the following functionality:

- Keeping track of the current count.
- Incrementing / decrementing the count.
- Fetching the current count value.

## Prerequisites

- **Internet connection**
- **An account in a Gno.land wallet, such as [Adena](https://adena.app)**

## 1. Using Gno Playground

When using the Gno Playground, writing, testing, deploying, and sharing Gno code
is simple. This makes it perfect for getting started with Gno.

Vising the [Playground](https://play.gno.land) will greet you with a template file:

![Default](../assets/how-to-guides/simple-contract/playground_welcome.png)

## 2. Start writing code

We can now write out the logic of the **Counter** Smart Contract in `package.gno`:

[embedmd]:# (../assets/how-to-guides/simple-contract/counter.gno go)
```go
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

There are a few things happening here, so let's dissect them:

- We defined the logic of our Realm into a package called `counter`.
- The package-level `count` variable stores the active count for the Realm (it is stateful).
- `Increment` and `Decrement` are public Realm methods, and as such are callable by users.
- `Increment` and `Decrement` directly modify the `count` value by making it go up or down (change state).
- Calling the `Render` method would return the `count` value as a formatted string. Learn more about the `Render`
  method and how it's used [here](../concepts/realms.md).

Alternatively, visit [this Playground link](https://play.gno.land/p/ONBa9eUEPKJ)
to view the code.


:::info A note on constructors
Gno Realms support a concept taken from other programming languages - _constructors_.

For example, to initialize the `count` variable with custom logic, we can specify that
logic within an `init` method, that is run **only once**, upon Realm deployment:

[embedmd]:# (../assets/how-to-guides/simple-contract/init.gno go)
```go
package counter

var count int

// ...

func init() {
	count = 2 * 10 // arbitrary value
}

// ...
```

:::

## Conclusion

That's it 🎉

You have successfully built a simple **Counter** Realm that is ready to be deployed on the Gno chain and called by users.
In the upcoming guides, we will see how we can develop more complex Realm logic and have them interact
with outside tools like a wallet application.
