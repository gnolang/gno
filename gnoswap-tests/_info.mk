# r/demo/wugnot from gno
ADDR_WUGNOT := g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6

# based on v1
ADDR_POOL := g148tjamj80yyrm309z7rk690an22thd2l3z8ank
ADDR_POSITION := g1q646ctzhvn60v492x8ucvyqnrj2w30cwh6efk5
ADDR_ROUTER := g1lm2l7tf49h3mykesct7rhfml30yx8dw5xrval7
ADDR_STAKER := g1cceshmzzlmrh7rr3z30j2t5mrvsq9yccysw9nu
ADDR_PROTOCOL_FEE := g1f7wpek7q67tkns27sw495u5yuu3a5wwjxw5l6l
ADDR_GOV_STAKER := g17e3ykyqk9jmqe2y9wxe9zhep3p7cw56davjqwa
ADR_GOV_GOV := g17s8w2ve7k85fwfnrk59lmlhthkjdted8whvqxd
ADDR_LAUNCHPAD := g122mau2lp2rc0scs8d27pkkuys4w54mdy2tuer3
ADDR_GNS := g1jgqwaa2le3yr63d533fj785qkjspumzv22ys5m
ADDR_GNFT := g1wxv2rdfn53qc84nt3nn646f9yh3nly8lm7j89t

# username address
ADDR_GNOSWAP := g1tzl3sgre0c2zgxfpws9xhq0c069wf7zqh6aqqy
ADDR_ADMIN := g1tzl3sgre0c2zgxfpws9xhq0c069wf7zqh6aqqy
ADDR_TEST := g1tzl3sgre0c2zgxfpws9xhq0c069wf7zqh6aqqy

# INCENTIVE_START
TOMORROW_MIDNIGHT := $(shell (gdate -ud 'tomorrow 00:00:00' +%s))
INCENTIVE_END := $(shell expr $(TOMORROW_MIDNIGHT) + 7776000) # 7776000 SECONDS = 90 DAY

# MAX_UINT64 := 18446744073709551615
MAX_APPROVE := 9223372036854775806
TX_EXPIRE := 9999999999

MAKEFILE := $(shell realpath $(firstword $(MAKEFILE_LIST)))
ROOT_DIR:=$(shell dirname $(MAKEFILE))/../


# TODO: change below 2 values based on which chain to deploy
GNOLAND_RPC_URL ?= localhost:26657
CHAINID ?= dev
