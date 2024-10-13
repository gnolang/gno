package gnoexporter

import (
	"net/http"
)

type Collector interface {
	Pattern() string
	Collect() http.HandlerFunc
}
