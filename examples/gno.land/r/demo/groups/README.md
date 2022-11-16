### - test package
	./build/gnodev test examples/gno.land/r/demo/groups/

### - add pkg
	./build/gnokey maketx addpkg pushkar --pkgdir "examples/gno.land/r/demo/groups" --deposit 100000000ugnot --gas-fee 1000000ugnot --gas-wanted 10000000 --broadcast true --chainid dev --remote 0.0.0.0:26657 --pkgpath "gno.land/r/demo/groups"

	
### - create group
	./build/gnokey maketx call pushkar --func "CreateGroup" --args "dao_trinity_ngo" --gas-fee "1000000ugnot" --gas-wanted "4000000" --broadcast true --chainid dev --remote 0.0.0.0:26657 --pkgpath "gno.land/r/demo/groups"

### - add member
	./build/gnokey maketx call pushkar --func "AddMember" --args "1" --args "g1hd3gwzevxlqmd3jsf64mpfczag8a8e5j2wdn3c" --args 12 --args "i am new user" --gas-fee "1000000ugnot" --gas-wanted "4000000" --broadcast true --chainid dev --remote 0.0.0.0:26657 --pkgpath "gno.land/r/demo/groups"


### - delete member
	./build/gnokey maketx call pushkar --func "DeleteMember" --args "1" --args "0" --gas-fee "1000000ugnot" --gas-wanted "4000000" --broadcast true --chainid dev --remote 0.0.0.0:26657 --pkgpath "gno.land/r/demo/groups"

### - delete group
	./build/gnokey maketx call pushkar --func "DeleteGroup" --args "1" --gas-fee "1000000ugnot" --gas-wanted "4000000" --broadcast true --chainid dev --remote 0.0.0.0:26657 --pkgpath "gno.land/r/demo/groups"
