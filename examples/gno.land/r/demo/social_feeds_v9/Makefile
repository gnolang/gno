KEY = test1
GNOKEY = gnokey maketx call \
    -pkgpath "gno.land/r/demo/social_feeds" \
    -gas-fee 1000000ugnot \
    -gas-wanted 3000000 \
    -send "" \
    -broadcast

ADDPKG = gnokey maketx addpkg \
    -deposit="1ugnot" \
    -gas-fee="1ugnot" \
    -gas-wanted="5000000" \
    -chainid="teritori-1" \
    -remote="https://testnet.gno.teritori.com:26657" \
    -broadcast="true"  \

.PHONY: create_feed
create_feed:
	${GNOKEY} \
	-func "CreateFeed" \
    -args "teritori" \
    ${KEY}

.PHONY: create_post
create_post:
	${GNOKEY} \
    -func "CreatePost" \
    -args "1" \
    -args "0" \
    -args "2" \
    -args '{"gifs": [], "files": [], "title": "", "message": "Hello world 2 !", "hashtags": [], "mentions": [], "createdAt": "2023-08-03T01:39:45.522Z", "updatedAt": "2023-08-03T01:39:45.522Z"}' \
    ${KEY}

.PHONY: tip_post
tip_post:
	${GNOKEY} \
	-send "1000000ugnot" \
	-func "TipPost" \
    -args "1" \
    -args "1" \
    ${KEY}	


.PHONY: flag_post
flag_post:
	${GNOKEY} \
	-func "FlagPost" \
    -args "1" \
    -args "1" \
    ${KEY}	

.PHONY: get_post
get_post:
	gnokey query vm/qeval --data 'gno.land/r/demo/social_feeds\nGetPosts(1, "", []uint64{}, 0, 10)'

.PHONY: propose_ban_post
propose_ban_post:
	${GNOKEY} \
    -pkgpath "gno.land/r/demo/social_feeds_dao" \
	-func "Propose" \
    -args "0" \
    -args "Ban Post" \
    -args "" \
    -args "" \
    ${KEY}	

.PHONY: vote_yes
vote_yes:
	${GNOKEY} \
    -pkgpath "gno.land/r/demo/social_feeds_dao" \
	-func "Vote" \
    -args "0" \
    -args "0" \
    -args "0" \
    -args "This is not good" \
    ${KEY}

.PHONY: execute_proposal
execute_proposal:
	${GNOKEY} \
    -pkgpath "gno.land/r/demo/social_feeds_dao" \
	-func "Execute" \
    -args "0" \
    -args "0" \
    ${KEY}

.PHONY: add_member
add_member:
	${GNOKEY} \
    -pkgpath "gno.land/r/demo/groups_v9" \
    -func "AddMember" \
    -args "0000000001" \
    -args "g1kcdd3n0d472g2p5l8svyg9t0wq6h5857nq992f" \
    -args "1" \
    -args "" \
    ${KEY}

.PHONY: add_pkg_social_feeds
add_pkg_social_feeds:
	${ADDPKG} \
    -pkgdir="." \
    -pkgpath="gno.land/r/demo/social_feeds_v7" \
    ${KEY}

.PHONY: add_pkg_social_feeds_dao
add_pkg_social_feeds_dao:
	${ADDPKG} \
    -pkgdir="../social_feeds_dao" \
    -pkgpath="gno.land/r/demo/social_feeds_dao" \
    ${KEY}

.PHONY: init
init: create_feed create_post create_post tip_post flag_post
    
.PHONY: upgrade_pkg
upgrade_pkg:
    gnokey maketx addpkg \
        -deposit="1ugnot" \
        -gas-fee="1ugnot" \
        -gas-wanted="7000000" \
        -chainid="teritori-1" \
        -remote="https://testnet.gno.teritori.com:26657" \
        -broadcast="true" \
        -pkgdir="." \
        -pkgpath="gno.land/r/demo/social_feeds_dao" \
        test1

    gnokey maketx addpkg \
        -deposit="1ugnot" \
        -gas-fee="1ugnot" \
        -gas-wanted="5000000" \
        -chainid="teritori-1" \
        -remote="https://testnet.gno.teritori.com:26657" \
        -broadcast="true" \
        -pkgdir="." \
        -pkgpath="gno.land/p/demo/jsonutil_v4" \
        test1

    gnokey maketx addpkg \
        -deposit="1ugnot" \
        -gas-fee="1ugnot" \
        -gas-wanted="5000000" \
        -chainid="teritori-1" \
        -remote="https://testnet.gno.teritori.com:26657" \
        -broadcast="true" \
        -pkgdir="." \
        -pkgpath="gno.land/p/demo/daodao/core_v6" \
        test1
