package gnoweb

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	md "github.com/gnolang/gno/gno.land/pkg/markdown"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type WebClient struct {
	logger *slog.Logger
	client *gnoclient.Client
	md     goldmark.Markdown
}

func NewWebClient(log *slog.Logger, cl *gnoclient.Client, m goldmark.Markdown) *WebClient {
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
	res, err := s.query(qpath, []byte(path))
	if err != nil {
		return nil, err
	}

	return res, nil
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
		return nil, fmt.Errorf("unable render real %q: %q", data, err)
	}

	var meta Metadata
	meta.Toc, err = md.TocInspect(doc, rawres, md.TocOptions{MaxDepth: 6})
	if err != nil {
		s.logger.Warn("unable to inspect for toc elements", "err", err)
	}

	return &meta, nil
}

func (s *WebClient) query(qpath string, data []byte) ([]byte, error) {
	s.logger.Info("query", "qpath", qpath, "data", string(data))
	// XXX: move this into gnoclient
	qres, err := s.client.Query(gnoclient.QueryCfg{
		Path: qpath,
		Data: data,
	})

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
