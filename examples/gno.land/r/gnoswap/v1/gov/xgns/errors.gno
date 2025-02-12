package xgns

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

var (
	errNoPermission = errors.New("[GNOSWAP-XGNS-001] caller has no permission")
)

func addDetailToError(err error, detail string) string {
	finalErr := ufmt.Errorf("%s || %s", err.Error(), detail)
	return finalErr.Error()
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
