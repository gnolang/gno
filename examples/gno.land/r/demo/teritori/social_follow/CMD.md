gnokey maketx addpkg  \
  -deposit="1ugnot" \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgdir="." \
  -pkgpath="gno.land/r/demo/social_follow_2" \
  mykey2

gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/social_follow_2" \
  -func="Follow" \
  -args="g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq" \
  mykey3

gnokey query "vm/qeval" -data='gno.land/r/demo/social_follow_2
Followers("g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq",0,1)' -remote="51.15.236.215:26657"

gnokey query "vm/qeval" -data='gno.land/r/demo/social_follow_2
FollowersCount("g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq")' -remote="51.15.236.215:26657"

gnokey query "vm/qeval" -data='gno.land/r/demo/social_follow_2
Followed("g1j3ylca07vlhklzftrznw7jyquzqf2wtxvjdm4r",0,1)' -remote="51.15.236.215:26657"

gnokey query "vm/qeval" -data='gno.land/r/demo/social_follow_2
FollowedCount("g1j3ylca07vlhklzftrznw7jyquzqf2wtxvjdm4r")' -remote="51.15.236.215:26657"

gnokey maketx call \
  -gas-fee="1ugnot" \
  -gas-wanted="5000000" \
  -broadcast="true" \
  -remote="51.15.236.215:26657" \
  -chainid="teritori-1" \
  -pkgpath="gno.land/r/demo/social_follow_2" \
  -func="Unfollow" \
  -args="g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq" \
  mykey3

gnokey query "vm/qeval" -data='gno.land/r/demo/social_follow_2
Followers("g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq",0,1)' -remote="51.15.236.215:26657"

gnokey query "vm/qeval" -data='gno.land/r/demo/social_follow_2
FollowersCount("g1c5y8jpe585uezcvlmgdjmk5jt2glfw88wxa3xq")' -remote="51.15.236.215:26657"

gnokey query "vm/qeval" -data='gno.land/r/demo/social_follow_2
Followed("g1j3ylca07vlhklzftrznw7jyquzqf2wtxvjdm4r",0,1)' -remote="51.15.236.215:26657"

gnokey query "vm/qeval" -data='gno.land/r/demo/social_follow_2
FollowedCount("g1j3ylca07vlhklzftrznw7jyquzqf2wtxvjdm4r")' -remote="51.15.236.215:26657"