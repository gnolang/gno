#!/bin/zsh

gnokey query \
"vm/qeval" \
--data "gno.land/r/waymobetta/noob
GetOwner()" \
--remote localhost:26657

# This command is used to interact with the Gnokey CLI tool for reading data from the chain.

# The `query` subcommand initiates this process of retrieving data from the chain, and we're passing several arguments to it using flags. 

# Here's what each flag does:

# - `"vm/qeval"`: Is a call to the VM for querying data from the chain.

# - `--data`: Sets the argument to pass into "vm/qeval"; in this case, a string including the realm path along with the function that we are wishing to call.

# - `--remote`: Sets the URL of a local Gno.land node running on your machine that will be used to execute this transaction. In our case, we're using the default value ("localhost:26657") since we've already started an instance of Gno (a Gno.land client) in another terminal window earlier on.
