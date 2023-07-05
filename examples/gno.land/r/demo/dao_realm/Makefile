realmpath=r/demo/foo_dao_3
realmuri=gno.land/$(realmpath)

gnohome=$(HOME)/Code/gno

chainid=dev
gnoland=localhost:26657
gnoweb=http://localhost:8888
wallet=onblock-test

#chainid=test3
#gnoland=test3.gno.land:36657
#gnoweb=https://test3.gno.land
#wallet=dev

.PHONY: deploy.realm
deploy.realm:
	gnokey maketx addpkg  \
		-deposit="1ugnot" \
		-gas-fee="1ugnot" \
		-gas-wanted="5000000" \
		-broadcast="true" \
		-remote="$(gnoland)" \
		-chainid="$(chainid)" \
		-pkgdir="." \
		-pkgpath="$(realmuri)" \
		$(wallet)
		
.PHONY: open.realm
open.realm:
	open "$(gnoweb)/$(realmpath)"


.PHONY: open.realm-doc
open.realm-doc:
	open "$(gnoweb)/$(realmpath)?help"

.PHONY: query.vote
define queryadmindata
$(realmuri)
GetCore().ProposalModules()[0].GetBallot(0, "g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv").Vote
endef
export queryadmindata
query.vote:
	gnokey query -data="$$queryadmindata" "vm/qeval" -remote="$(gnoland)"