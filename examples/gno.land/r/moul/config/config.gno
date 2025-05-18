package config

import (
	"gno.land/p/moul/authz"
)

var (
	Authority, privConfig = authz.New("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // when #4264 -> authz.NewWithOrigin()
	SafeConfig            = privConfig.Safe()
)
