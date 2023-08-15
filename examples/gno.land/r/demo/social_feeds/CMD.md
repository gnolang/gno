gnokey maketx call \
    -pkgpath "gno.land/r/demo/social_feeds" \
    -func "CreateFeed" \
    -gas-fee 1000000ugnot \
    -gas-wanted 3000000 \
    -send "" \
    -broadcast \
    -args "teritori" \
    test1

gnokey maketx call \
    -pkgpath "gno.land/r/demo/social_feeds" \
    -func "CreatePost" \
    -gas-fee 1000000ugnot \
    -gas-wanted 2000000 \
    -send "" \
    -broadcast \
    -args "1" \
    -args "0" \
    -args "2" \
    -args '{"gifs": [], "files": [], "title": "", "message": "Hello world 2 !", "hashtags": [], "mentions": [], "createdAt": "2023-08-03T01:39:45.522Z", "updatedAt": "2023-08-03T01:39:45.522Z"}' \
    test1 

gnokey maketx addpkg \
    -deposit="1ugnot" \
    -gas-fee="1ugnot" \
    -gas-wanted="5000000" \
    -broadcast="true"  \
    -pkgdir="." \
    -pkgpath="gno.land/r/demo/social_feeds_v2" \
    test1

gnokey maketx call \
    -pkgpath "gno.land/r/demo/social_feeds" \
    -func "MigrateFromPreviousFeed" \
    -gas-fee 1000000ugnot \
    -gas-wanted 2000000 \
    -send "" \
    -broadcast \
    test1 

