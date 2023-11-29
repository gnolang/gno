#!/bin/zsh

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/noob" \
--func "Noob" \
--args "baz" \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
demo

# This command is used to interact with the Gnokey CLI tool for creating a new transaction. 

# The `maketx` subcommand initiates this process, and we're passing several arguments to it using flags. 

# Here's what each flag does:

# - `call`: Executes a call to the realm, invoking a function.

# - `--pkgpath`: Specifies the path of the module within its repository (in our case, it's "gno.land/r/waymobetta").

# - `--func`: Specifies the function to be invoked within the realm.

# - `--args`: Sepcifies arguments to pass into the function referenced from the `--func` flag. Multiple `--args` flags can be passed in based on the number of anticipated parameters found within the function's signature.

# - `--gas-fee` and `--gas-wanted`: Set the gas fee and desired gas limit for the transaction, respectively. These values are used by the realm to calculate the actual amount of GNOT (or ugnot, a smaller denomination) required to execute this transaction on the Gno.land network. In our case, we're setting both flags to 1 million units of ugnot tokens as a way to ensure that there are enough funds available for gas fees and to avoid underestimating the required amount of ugnot during transaction execution.

# - `--broadcast`: Broadcasts the transaction hash (i.e., its unique identifier) once it's been executed on the Gno.land network, making it visible to other users who are also monitoring this address.

# - `--chainid`: Specifies which blockchain network you want to use (in our case, "dev" for Gno.land's development environment).

# - `--remote`: Sets the URL of a local Gno.land node running on your machine that will be used to execute this transaction. In our case, we're using the default value ("localhost:26657") since we've already started an instance of Gno (a Gno.land client) in another terminal window earlier on.

# - `demo`: This is a placeholder argument that you can replace with your own account (wallet) name, depending on how you want to organize and label your accounts. In our case, we're using "demo" as a simple identifier for this particular transaction.
