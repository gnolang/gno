### - test package

    ./build/gno test examples/gno.land/r/demo/groups/

### - add pkg

    ./build/gnokey maketx addpkg -pkgdir "examples/gno.land/r/demo/groups" -deposit 100000000ugnot -gas-fee 1000000ugnot -gas-wanted 10000000 -broadcast -chainid dev -remote 0.0.0.0:26657 -pkgpath "gno.land/r/demo/groups" test1 

### - create group

    ./build/gnokey maketx call -func "CreateGroup" -args "dao_trinity_ngo" -gas-fee "1000000ugnot" -gas-wanted 4000000 -broadcast -chainid dev -remote 0.0.0.0:26657 -pkgpath "gno.land/r/demo/groups" test1 

### - add member

    ./build/gnokey maketx call -func "AddMember" -args "1" -args "g1hd3gwzevxlqmd3jsf64mpfczag8a8e5j2wdn3c" -args 12 -args "i am new user" -gas-fee "1000000ugnot" -gas-wanted "4000000" -broadcast -chainid dev -remote 0.0.0.0:26657 -pkgpath "gno.land/r/demo/groups" test1

### - delete member

    ./build/gnokey maketx call -func "DeleteMember" -args "1" -args "0" -gas-fee "1000000ugnot" -gas-wanted "4000000" -broadcast -chainid dev -remote 0.0.0.0:26657 -pkgpath "gno.land/r/demo/groups" test1

### - delete group

    ./build/gnokey maketx call -func "DeleteGroup" -args "1" -gas-fee "1000000ugnot" -gas-wanted "4000000" -broadcast -chainid dev -remote 0.0.0.0:26657 -pkgpath "gno.land/r/demo/groups" test1

