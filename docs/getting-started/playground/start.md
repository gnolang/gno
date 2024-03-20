---
id: start
---

# Gno Playground

## Overview

The Gno Playground is a web editor and sandbox that enables developers to 
interactively work with the Gno language. It makes coding, testing,
and deploying simple with its diverse set of tools and features. Users can
share code, run tests, and deploy projects to Gno.land networks, 
making it the perfect tool to get started with Gno development.

## Prerequisites

- **Internet connection**
- **A keypair in a Gno.land wallet, such as [Adena](https://adena.app)**

To fully utilize Playground features, you will need to install a Gno.land web
browser wallet, such as [Adena](https://www.adena.app/), and create a keypair. This will allow you to
fully utilize the Playground.

## Playground Features

To get started, visit the Playground at [play.gno.land](https://play.gno.land). You will be greeted with a
simple `package.gno` file:

![default_playground](../../assets/getting-started/playground/default_playground.png)

The Playground has the following features:
- `Share` - Generate a unique, short, and shareable identifier for your Gno code.
- `Deploy` - Connect your wallet and deploy your code to a Gno.land network
- `Format` - Automatically adjust your Gno code's structure and style for optimal readability and consistency.
- `Run` - Execute a particular expression within your code to validate its functionality and output.
- `Test` - Execute predefined tests to verify your code's integrity and ensure it meets expected outcomes.
- `REPL` - Experiment and troubleshoot in real-time using the GnoVM with interactive REPL features.
interactive REPL features (experimental)

Let's dive into each of the Playground features.

### Share

This feature allows users to get a permanent shortlink to the code found in the
playground at the time of clicking. This way, Gno code can be shared easily. 

### Deploy

Allows users to seamlessly deploy their Gno code to the chain. After connecting 
a Gno.land wallet, users can select their desired package path and network for deployment.
as well as which network.

![default_deploy](../../assets/getting-started/playground/default_deploy.png)

After inputting your desired package path, you can select the network you would 
like to deploy to, such as [Portal Loop](../../concepts/portal-loop.md) or local,
and click deploy.

:::info
Even if you don't have testnet tokens, the Playground will automatically provide
you with enough to cover the gas cost at the time of deployment.
:::

### Format
Format will format your code, using `gofmt`.

### Run
Run will allow you to run an expression on your Gno code. Take the following code
for an example:

![run_example](../../assets/getting-started/playground/run.png)

Running `println(Render("Gnopher"))` will display the following output:

```bash
Hello Gnopher!
```

View the code [here](https://play.gno.land/p/nBq2W8drjMy).

### Test

Test will look for `_test.gno` files in your playground and run the
`gno test -v` command on them. Testing your code will open a terminal that will 
show you the output of the test. Read more about how Gno tests work
[here](../../concepts/gno-test.md).

### REPL (experimental)

This option, although experimental, will let you experiment with the GnoVM
in REPL mode. 

## Learning about Gno.land & writing Gno code

Gno.land is a complex technical system, and as such many concepts need to be 
explained to newcomers. For reading more about Gno.land, 
check out the [Concepts](../../concepts/concepts.md) section.

To get started writing Gno code, check out the
[How-to](../../how-to-guides/how-to-guides.md) section, the `examples/` folder on
the [Gno monorepo](https://github.com/gnolang/gno), or one of many community projects and tutorials found in the 
[awesome-gno](https://github.com/gnolang/awesome-gno/blob/main/README.md) repo on GitHub.





