package gnoweb

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type WebClient struct {
	logger *slog.Logger
	client *client.RPCClient
	md     goldmark.Markdown
}

func NewWebClient(log *slog.Logger, cl *client.RPCClient, m goldmark.Markdown) *WebClient {
	m.Parser().AddOptions(parser.WithAutoHeadingID())
	return &WebClient{
		logger: log,
		client: cl,
		md:     m,
	}
}

func (s *WebClient) Functions(pkgPath string) ([]vm.FunctionSignature, error) {
	const qpath = "vm/qfuncs"

	args := fmt.Sprintf("gno.land/%s", strings.Trim(pkgPath, "/"))
	res, err := s.query(qpath, []byte(args))
	if err != nil {
		return nil, fmt.Errorf("unable query funcs list: %w", err)
	}

	var fsigs vm.FunctionSignatures
	if err := amino.UnmarshalJSON(res, &fsigs); err != nil {
		s.logger.Warn("unable to unmarshal fsigs, client is probably outdated ?")
		return nil, fmt.Errorf("unable to unamarshal fsigs: %w", err)
	}

	return fsigs, nil
}

func (s *WebClient) SourceFile(path, fileName string) ([]byte, error) {
	const qpath = "vm/qfile"

	fileName = strings.TrimSpace(fileName) // sanitize filename
	if fileName == "" {
		return nil, errors.New("empty filename given") // XXX -> ErrXXX
	}

	// XXX: move this into gnoclient ?
	path = fmt.Sprintf("gno.land/%s", strings.Trim(path, "/"))
	path = filepath.Join(path, fileName)
	return s.query(qpath, []byte(path))
}

func (s *WebClient) Sources(path string) ([]string, error) {
	const qpath = "vm/qfile"

	// XXX: move this into gnoclient
	path = fmt.Sprintf("gno.land/%s", strings.Trim(path, "/"))
	res, err := s.query(qpath, []byte(path))
	if err != nil {
		return nil, err
	}

	files := strings.Split(string(res), "\n")
	return files, nil
}

type Metadata struct {
	*md.Toc
}

func (s *WebClient) Render(w io.Writer, pkgPath string, args string) (*Metadata, error) {
	const qpath = "vm/qrender"

	data := []byte(gnoPath(pkgPath, args))
	rawres, err := s.query(qpath, data)
	if err != nil {
		return nil, err
	}

	doc := s.md.Parser().Parse(text.NewReader(rawres))
	if err := s.md.Renderer().Render(w, rawres, doc); err != nil {
		return nil, fmt.Errorf("unable render real %q: %w", data, err)
	}

	var meta Metadata
	meta.Toc, err = md.TocInspect(doc, rawres, md.TocOptions{MaxDepth: 6, MinDepth: 2})
	if err != nil {
		s.logger.Warn("unable to inspect for toc elements", "err", err)
	}

	return &meta, nil
}

func (s *WebClient) query(qpath string, data []byte) ([]byte, error) {
	s.logger.Info("query", "qpath", qpath, "data", string(data))

	qres, err := s.client.ABCIQuery(qpath, data)
	if err != nil {
		s.logger.Error("request error", "path", qpath, "data", string(data), "error", err)
		return nil, fmt.Errorf("unable to query path %q: %w", qpath, err)
	}
	if qres.Response.Error != nil {
		s.logger.Error("response error", "path", qpath, "log", qres.Response.Log)
		return nil, qres.Response.Error
	}

	return qres.Response.Data, nil
}

func gnoPath(pkgPath, args string) string {
	pkgPath = strings.Trim(pkgPath, "/")
	return fmt.Sprintf("gno.land/%s:%s", pkgPath, args)
}
