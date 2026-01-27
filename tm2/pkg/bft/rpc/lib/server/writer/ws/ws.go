package ws

import (
	"encoding/json"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/writer"
	"github.com/olahol/melody"
)

var _ writer.ResponseWriter = (*ResponseWriter)(nil)

type ResponseWriter struct {
	logger *slog.Logger

	s *melody.Session
}

func New(logger *slog.Logger, s *melody.Session) ResponseWriter {
	return ResponseWriter{
		logger: logger.With("writer", "ws-writer"),
		s:      s,
	}
}

func (w ResponseWriter) WriteResponse(response any) {
	jsonRaw, encodeErr := json.Marshal(response)
	if encodeErr != nil {
		w.logger.Error(
			"unable to encode JSON-RPC response",
			"err", encodeErr,
		)

		return
	}

	if err := w.s.Write(jsonRaw); err != nil {
		w.logger.Error(
			"unable to write WS response",
			"err", err,
		)
	}
}
