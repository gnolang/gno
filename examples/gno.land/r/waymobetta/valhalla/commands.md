### commands

#### deploy

gnokey maketx \
addpkg \
--pkgpath "gno.land/r/waymobetta/valhalla" \
--pkgdir "examples/gno.land/r/waymobetta/valhalla" \
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
--pkgpath "gno.land/r/waymobetta/valhalla" \
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
--pkgpath "gno.land/r/waymobetta/valhalla" \
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
-data "gno.land/r/waymobetta/valhalla
IsMember(\"test\")" \
-remote localhost:26657

#### query member list

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/valhalla
MemberList" \
-remote localhost:26657

#### query member count

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/valhalla
MemberCount()" \
-remote localhost:26657

#### query admin

gnokey query \
"vm/qeval" \
-data "gno.land/r/waymobetta/valhalla
GetAdmin()" \
-remote localhost:26657
