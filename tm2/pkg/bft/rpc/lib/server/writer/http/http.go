package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/writer"
)

var _ writer.ResponseWriter = (*ResponseWriter)(nil)

type ResponseWriter struct {
	logger *slog.Logger

	w http.ResponseWriter
}

func New(logger *slog.Logger, w http.ResponseWriter) ResponseWriter {
	return ResponseWriter{
		logger: logger.With("writer", "http-writer"),
		w:      w,
	}
}

func (h ResponseWriter) WriteResponse(response any) {
	if err := json.NewEncoder(h.w).Encode(response); err != nil {
		h.logger.Error(
			"unable to encode JSON response",
			"err", err,
		)
	}
}
