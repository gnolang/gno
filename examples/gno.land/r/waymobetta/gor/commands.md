### commands

```bash
#### deploy

gnokey maketx \
addpkg \
--pkgpath "gno.land/r/waymobetta/gor" \
--pkgdir "examples/gno.land/r/waymobetta/gor" \
--deposit 100000000ugnot \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
main

#### add member

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/gor" \
--func "AddMember" \
--args "test" \
--gas-fee "1000000ugnot" \
--gas-wanted "2000000" \
--broadcast \
--chainid dev \
--remote localhost:26657 \
main

#### remove member

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/gor" \
--func "RemoveMember" \
--args "test" \
--gas-fee "1000000ugnot" \
--gas-wanted "2000000" \
--broadcast \
--chainid dev \
--remote localhost:26657 \
main

#### query member

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/gor
GorMapping" \
-remote localhost:26657

#### query admin

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/gor
GetAdmin()" \
-remote localhost:26657

```
