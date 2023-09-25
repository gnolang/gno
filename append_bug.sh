# cd gno.land && go run ./cmd/gnoland start in an other terminal first
#
# Call Append 1
gnokey maketx call -pkgpath "gno.land/r/demo/bug/append" -func "Append" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "1" -remote "127.0.0.1:26657" $1
# Call Append 2
gnokey maketx call -pkgpath "gno.land/r/demo/bug/append" -func "Append" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "2" -remote "127.0.0.1:26657" $1
# Call Append 3
gnokey maketx call -pkgpath "gno.land/r/demo/bug/append" -func "Append" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "3" -remote "127.0.0.1:26657" $1

# Call render
gnokey maketx call -pkgpath "gno.land/r/demo/bug/append" -func "Render" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "" -remote "127.0.0.1:26657" $1
# Outputs ("1</br>2</br>3</br>" string) -> OK

# Call Pop
gnokey maketx call -pkgpath "gno.land/r/demo/bug/append" -func "Pop" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -remote "127.0.0.1:26657" $1
# Call render
gnokey maketx call -pkgpath "gno.land/r/demo/bug/append" -func "Render" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "" -remote "127.0.0.1:26657" $1
# Outputs ("1</br>2</br>" string) -> WRONG! Pop removes the first item so
# it should be ("2</br>3</br>" string)

# Call Append 42
gnokey maketx call -pkgpath "gno.land/r/demo/bug/append" -func "Append" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "42" -remote "127.0.0.1:26657" $1

# Call render
gnokey maketx call -pkgpath "gno.land/r/demo/bug/append" -func "Render" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "" -remote "127.0.0.1:26657" $1
# Ouputs ("1</br>2</br>3</br>" string) -> WTF where is 42 ???

