package rpcserver

import (
	"encoding/hex"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

func GetParam(r *http.Request, param string) string {
	s := r.URL.Query().Get(param)
	if s == "" {
		s = r.FormValue(param)
	}
	return s
}

func GetParamByteSlice(r *http.Request, param string) ([]byte, error) {
	s := GetParam(r, param)
	return hex.DecodeString(s)
}

func GetParamInt64(r *http.Request, param string) (int64, error) {
	s := GetParam(r, param)
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.New(param, err.Error())
	}
	return i, nil
}

func GetParamInt32(r *http.Request, param string) (int32, error) {
	s := GetParam(r, param)
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, errors.New(param, err.Error())
	}
	return int32(i), nil
}

func GetParamUint64(r *http.Request, param string) (uint64, error) {
	s := GetParam(r, param)
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.New(param, err.Error())
	}
	return i, nil
}

func GetParamUint(r *http.Request, param string) (uint, error) {
	s := GetParam(r, param)
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.New(param, err.Error())
	}
	return uint(i), nil
}

func GetParamRegexp(r *http.Request, param string, re *regexp.Regexp) (string, error) {
	s := GetParam(r, param)
	if !re.MatchString(s) {
		return "", errors.New(param, "did not match regular expression %v", re.String())
	}
	return s, nil
}

func GetParamFloat64(r *http.Request, param string) (float64, error) {
	s := GetParam(r, param)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errors.New(param, err.Error())
	}
	return f, nil
}
